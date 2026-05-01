package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoadSeckillStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadSeckillStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadSeckillStockLogic {
	return &LoadSeckillStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LoadSeckillStock 预热活动库存到Redis
// 此接口应在活动开始前由运营/定时任务调用
func (l *LoadSeckillStockLogic) LoadSeckillStock(in *order.LoadSeckillStockRequest) (*order.LoadSeckillStockResponse, error) {
	activity, err := l.svcCtx.SeckillActivity.GetActivityById(uint(in.ActivityId))
	if err != nil || activity == nil {
		return &order.LoadSeckillStockResponse{Success: false, Msg: "活动不存在"}, nil
	}

	// 计算缓存TTL：活动结束后再保留1小时
	ttl := int(time.Until(activity.EndTime).Seconds()) + 3600
	if ttl <= 0 {
		return &order.LoadSeckillStockResponse{Success: false, Msg: "活动已结束"}, nil
	}

	// 1. 将活动状态写入Redis缓存
	activityKey := fmt.Sprintf("seckill:activity:%d", in.ActivityId)
	if err := l.svcCtx.Cache.Set(activityKey, strconv.Itoa(activity.Status), ttl); err != nil {
		l.Logger.Errorf("缓存活动状态失败: %v", err)
		return &order.LoadSeckillStockResponse{Success: false, Msg: "缓存活动状态失败"}, nil
	}

	// 2. 初始化库存计数器到Redis
	stock, err := l.svcCtx.SeckillStockModel.FindByActivityId(int64(in.ActivityId))
	if err != nil || stock == nil {
		return &order.LoadSeckillStockResponse{Success: false, Msg: "库存记录不存在"}, nil
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", in.ActivityId)
	if err := l.svcCtx.Cache.Set(stockKey, strconv.Itoa(stock.SurplusStock), ttl); err != nil {
		l.Logger.Errorf("缓存库存失败: %v", err)
		return &order.LoadSeckillStockResponse{Success: false, Msg: "缓存库存失败"}, nil
	}

	// 3. 初始化订单编号计数器为0
	orderKey := fmt.Sprintf("seckill:order:%d", in.ActivityId)
	_ = l.svcCtx.Cache.Set(orderKey, "0", ttl)

	// 4. 更新活动状态为进行中
	activity.Status = 1
	if err := l.svcCtx.SeckillActivity.Update(activity); err != nil {
		l.Logger.Errorf("更新活动状态失败: %v", err)
	}
	// 同步更新Redis中的活动状态
	_ = l.svcCtx.Cache.Set(activityKey, "1", ttl)

	l.Logger.Infof("秒杀活动[%d]库存预热完成, stock=%d", in.ActivityId, stock.SurplusStock)
	return &order.LoadSeckillStockResponse{
		Success: true,
		Msg:     "预热成功",
	}, nil
}
