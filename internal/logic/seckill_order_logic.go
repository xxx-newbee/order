package logic

import (
	"context"
	"errors"
	"fmt"
	"order/internal/model"
	"strconv"

	"order/internal/svc"
	"order/order"

	"github.com/bsm/redislock"
	"github.com/zeromicro/go-zero/core/logx"
)

type SeckillOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSeckillOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeckillOrderLogic {
	return &SeckillOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SeckillOrderLogic) SeckillOrder(in *order.SeckillOrderRequest) (*order.SeckillOrderResponse, error) {
	// 1.检查秒杀活动状态
	activityKey := fmt.Sprintf("seckill:activity:%d", in.ActivityId)
	activityStr, err := l.svcCtx.Cache.Get(activityKey)
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
	// 释放锁
	defer func() {
		if err := lock.Release(context.TODO()); err != nil {
			if errors.Is(err, redislock.ErrLockNotHeld) {
				l.Logger.Errorf("release lock err: %s", err.Error())
			} else {
				l.Logger.Infof("release lock: %s", lockerKey)
			}
		}
	}()

	// 4.生成订单
	orderNo := strconv.FormatInt(int64(33), 10)
	order := &model.OrderMain{
		OrderNo:           orderNo,
		UserId:            in.UserId,
		TotalAmount:       0,
		PayAmount:         0,
		OrderStatus:       5,
		OrderType:         1,
		SeckillActivityId: in.ActivityId,
	}
	if err = l.svcCtx.OrderMainModel.Insert(order); err != nil {
		_, _ = l.svcCtx.Cache.Increase(stockKey)
		l.Logger.Errorf("order insert err: %s", err.Error())
		return nil, err
	}

}
