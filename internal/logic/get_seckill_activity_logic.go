package logic

import (
	"context"
	"fmt"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSeckillActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeckillActivityLogic {
	return &GetSeckillActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSeckillActivityLogic) GetSeckillActivity(in *order.GetSeckillActivityRequest) (*order.GetSeckillActivityResponse, error) {
	activity, err := l.svcCtx.SeckillActivity.GetActivityById(uint(in.ActivityId))
	if err != nil {
		l.Logger.Errorf("查询活动失败: %v", err)
		return &order.GetSeckillActivityResponse{Success: false, Msg: "查询失败"}, nil
	}
	if activity == nil {
		return &order.GetSeckillActivityResponse{Success: false, Msg: "活动不存在"}, nil
	}

	// 查询剩余库存
	surplusStock := int32(0)
	stock, err := l.svcCtx.SeckillStockModel.FindByActivityId(int64(in.ActivityId))
	if err == nil && stock != nil {
		surplusStock = int32(stock.SurplusStock)
	}

	// 同时从Redis获取实时缓存状态
	activityKey := fmt.Sprintf("seckill:activity:%d", in.ActivityId)
	activityStr, _ := l.svcCtx.Cache.Get(activityKey)
	if activityStr != "" && activityStr == "1" {
		activity.Status = 1 // 修正为进行中
	}

	return &order.GetSeckillActivityResponse{
		ActivityId:    uint32(activity.ID),
		ProductId:     activity.ProductId,
		SeckillPrice:  activity.SeckillPrice,
		StockNum:      int32(activity.StockNum),
		SurplusStock:  surplusStock,
		StartTime:     activity.StartTime.Unix(),
		EndTime:       activity.EndTime.Unix(),
		Status:        int32(activity.Status),
		Success:       true,
		Msg:           "ok",
	}, nil
}
