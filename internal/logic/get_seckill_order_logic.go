package logic

import (
	"context"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSeckillOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeckillOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeckillOrderLogic {
	return &GetSeckillOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSeckillOrderLogic) GetSeckillOrder(in *order.GetSeckillOrderRequest) (*order.GetSeckillOrderResponse, error) {
	if in.OrderNo == "" {
		return &order.GetSeckillOrderResponse{Success: false, Msg: "订单号不能为空"}, nil
	}

	orderMain, err := l.svcCtx.OrderMainModel.FindByOrderNo(in.OrderNo)
	if err != nil {
		l.Logger.Errorf("查询订单失败: %v", err)
		return &order.GetSeckillOrderResponse{Success: false, Msg: "查询失败"}, nil
	}
	if orderMain == nil {
		return &order.GetSeckillOrderResponse{Success: false, Msg: "订单不存在"}, nil
	}

	return &order.GetSeckillOrderResponse{
		OrderNo:           orderMain.OrderNo,
		UserId:            orderMain.UserId,
		TotalAmount:       orderMain.TotalAmount,
		PayAmount:         orderMain.PayAmount,
		OrderStatus:       int32(orderMain.OrderStatus),
		OrderType:         int32(orderMain.OrderType),
		SeckillActivityId: orderMain.SeckillActivityId,
		CreatedAt:         orderMain.CreatedAt.Format("2006-01-02 15:04:05"),
		Success:           true,
		Msg:               "ok",
	}, nil
}
