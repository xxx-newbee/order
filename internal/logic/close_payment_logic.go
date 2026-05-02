package logic

import (
	"context"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClosePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClosePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClosePaymentLogic {
	return &ClosePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClosePaymentLogic) CloseOrder(in *order.CloseOrderRequest) (*order.CloseOrderResponse, error) {
	orderMain, err := l.svcCtx.OrderMainModel.FindByOrderNo(in.OrderNo)
	if err != nil || orderMain == nil {
		return &order.CloseOrderResponse{Success: false, Msg: "订单不存在"}, nil
	}
	if orderMain.OrderStatus != 0 {
		return &order.CloseOrderResponse{Success: false, Msg: "订单状态不允许关单"}, nil
	}

	if err := l.svcCtx.WxPayClient.CloseOrder(in.OrderNo); err != nil {
		l.Logger.Errorf("关闭微信订单失败: %s", err.Error())
		return &order.CloseOrderResponse{Success: false, Msg: "关单失败: " + err.Error()}, nil
	}

	if err := l.svcCtx.OrderMainModel.UpdateStatus(in.OrderNo, 4); err != nil {
		l.Logger.Errorf("更新订单状态失败: %s", err.Error())
	}

	record, _ := l.svcCtx.PaymentRecordModel.FindByOrderNo(in.OrderNo)
	if record != nil {
		record.TradeState = "CLOSED"
		_ = l.svcCtx.PaymentRecordModel.Update(record)
	}

	return &order.CloseOrderResponse{Success: true, Msg: "关单成功"}, nil
}
