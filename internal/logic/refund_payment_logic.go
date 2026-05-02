package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/internal/wxpay"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefundLogic) Refund(in *order.RefundRequest) (*order.RefundResponse, error) {
	orderMain, err := l.svcCtx.OrderMainModel.FindByOrderNo(in.OrderNo)
	if err != nil || orderMain == nil {
		return &order.RefundResponse{Success: false, Msg: "订单不存在"}, nil
	}
	if orderMain.OrderStatus != 1 {
		return &order.RefundResponse{Success: false, Msg: "订单状态不允许退款"}, nil
	}

	record, err := l.svcCtx.PaymentRecordModel.FindByOrderNo(in.OrderNo)
	if err != nil || record == nil {
		return &order.RefundResponse{Success: false, Msg: "支付记录不存在"}, nil
	}
	if record.TradeState != "SUCCESS" {
		return &order.RefundResponse{Success: false, Msg: "订单未支付，无法退款"}, nil
	}
	if record.RefundStatus == "SUCCESS" {
		return &order.RefundResponse{Success: false, Msg: "已退款成功，不可重复退款"}, nil
	}

	refundNo := fmt.Sprintf("REFUND_%s_%d", in.OrderNo, time.Now().UnixMilli())
	totalCents := int(record.PayerTotal)

	wxReq := &wxpay.RefundReq{
		OutTradeNo:  in.OrderNo,
		OutRefundNo: refundNo,
		Reason:      in.RefundReason,
		Amount: &wxpay.RefundAmount{
			Refund:   int(in.RefundAmount),
			Total:    totalCents,
			Currency: "CNY",
		},
	}

	resp, err := l.svcCtx.WxPayClient.Refund(wxReq)
	if err != nil {
		l.Logger.Errorf("退款失败: %s", err.Error())
		return &order.RefundResponse{Success: false, Msg: "退款失败: " + err.Error()}, nil
	}

	record.RefundId = resp.RefundID
	record.RefundStatus = resp.Status
	record.RefundAmount = float64(in.RefundAmount) / 100.0
	if err := l.svcCtx.PaymentRecordModel.Update(record); err != nil {
		l.Logger.Errorf("更新退款记录失败: %s", err.Error())
	}

	return &order.RefundResponse{
		Success:      true,
		RefundId:     resp.RefundID,
		RefundStatus: resp.Status,
		Msg:          "退款已发起",
	}, nil
}
