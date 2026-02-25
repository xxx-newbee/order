package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xxx-newbee/order/internal/model"
	"github.com/xxx-newbee/storage"
	"github.com/xxx-newbee/storage/queue"

	"strconv"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/bsm/redislock"
	"github.com/zeromicro/go-zero/core/logx"
)

type SeckillOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSeckillOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeckillOrderLogic {
	l := &SeckillOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
	l.svcCtx.RedisQueue.Register(model.SeckillStock{}.TableName(), l.SeckillStockConsumer)
	l.svcCtx.RedisQueue.Run()
	return l
}

func (l *SeckillOrderLogic) SeckillOrder(in *order.SeckillOrderRequest) (*order.SeckillOrderResponse, error) {
	// 1.检查秒杀活动状态
	activityKey := fmt.Sprintf("seckill:activity:%d", in.ActivityId)
	activityStr, err := l.svcCtx.Cache.Get(activityKey)
	// 根据活动状态进行判断
	if err != nil || activityStr == "" {
		return &order.SeckillOrderResponse{Success: false, Msg: "秒杀活动不存在或未开始"}, nil
	}

	// 2.获取库存量，预扣减库存
	stockKey := fmt.Sprintf("seckill:stock:%d", in.ActivityId)
	surplusStock, err := l.svcCtx.Cache.Decrease(stockKey)
	if err != nil || surplusStock < 0 {
		// 库存不足，恢复Redis库存
		_, _ = l.svcCtx.Cache.Increase(stockKey)
		return &order.SeckillOrderResponse{Success: false, Msg: "秒杀库存不足"}, nil
	}

	// 3.分布式锁，防止重复下单
	lockerKey := fmt.Sprintf("seckill:locker:%d:%d", in.UserId, in.ActivityId)
	lock, err := l.svcCtx.Locker.Lock(lockerKey, 5, nil)
	if err != nil {
		// 重复下单
		l.svcCtx.Cache.Increase(stockKey)
		return &order.SeckillOrderResponse{Success: false, Msg: "请勿重复下单"}, nil
	}
	// 释放锁
	defer func() {
		if err := lock.Release(context.TODO()); err != nil {
			if errors.Is(err, redislock.ErrLockNotHeld) {
				l.Logger.Errorf("释放分布式锁失败: %s", err.Error())
			} else {
				l.Logger.Infof("释放分布式锁: %s", lockerKey)
			}
		}
	}()

	// 4.生成订单
	orderKey := fmt.Sprintf("seckill:order:%d", in.ActivityId)
	order_no, err := l.svcCtx.Cache.Increase(orderKey)
	if err != nil {
		l.Logger.Errorf("生成订单号失败: %s", err.Error())
		return &order.SeckillOrderResponse{Success: false, Msg: "生成订单号失败"}, nil
	}
	orderNo := strconv.FormatInt(order_no, 10)
	order_main := &model.OrderMain{
		OrderNo:           orderNo,
		UserId:            in.UserId,
		TotalAmount:       0, // 秒杀价格：通过活动id查
		PayAmount:         0,
		OrderStatus:       5, // 订单状态：秒杀中
		OrderType:         1, // 订单类型：秒杀订单
		SeckillActivityId: in.ActivityId,
	}
	if err = l.svcCtx.OrderMainModel.Insert(order_main); err != nil {
		_, _ = l.svcCtx.Cache.Increase(stockKey)
		l.Logger.Errorf("创建订单失败: %s", err.Error())
		return &order.SeckillOrderResponse{Success: false, Msg: "创建订单失败"}, nil
	}

	// 5.消息队列异步扣减数据库库存
	vals := make(map[string]interface{})
	vals["user_id"] = in.UserId
	vals["order_no"] = orderNo
	vals["activity_id"] = in.ActivityId
	vals["product_id"] = in.ProductId

	msg := &queue.Message{
		Stream: model.SeckillStock{}.TableName(),
		Values: vals,
	}
	if err = l.svcCtx.RedisQueue.Append(msg); err != nil {
		_, _ = l.svcCtx.Cache.Increase(stockKey)
		l.Logger.Errorf("消息入队失败: %s", err.Error())
		return &order.SeckillOrderResponse{Success: false, Msg: "数据库异步扣减失败"}, nil
	}

	return &order.SeckillOrderResponse{
		Success: true,
		OrderNo: orderNo,
		Msg:     "秒杀下单成功，请尽快支付",
	}, nil
}

func (l *SeckillOrderLogic) SeckillStockConsumer(msg storage.Messager) error {
	order := struct {
		UserId     int64  `redis:"user_id"`
		OrderNo    string `json:"order_no"`
		ActivityId int64  `redis:"activity_id"`
		ProductId  int64  `redis:"product_id"`
	}{}
	rb, err := json.Marshal(msg.GetValues())
	if err != nil {
		return err
	}

	if err = json.Unmarshal(rb, &order); err != nil {
		return err
	}

	// 乐观锁扣减库存
	affected, err := l.svcCtx.SeckillStockModel.DecreaseStock(order.ActivityId)
	if err != nil || affected == 0 {
		// 扣减失败，更新订单为取消状态
		l.Logger.Errorf("数据库库存扣减失败，error: %s", err.Error())
		_ = l.svcCtx.OrderMainModel.UpdataStatus(order.OrderNo, 4)
		return err
	}
	// 扣减成功，更新订单为待付款状态
	l.Logger.Infof("秒杀库存扣减成功，orderNo：", order.OrderNo)
	return l.svcCtx.OrderMainModel.UpdataStatus(order.OrderNo, 0)
}
