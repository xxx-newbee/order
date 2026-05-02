package logic

import (
	"context"
	"time"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessPaymentCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessPaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessPaymentCallbackLogic {
	return &ProcessPaymentCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessPaymentCallbackLogic) ProcessPaymentCallback(in *order.PaymentCallbackRequest) (*order.PaymentCallbackResponse, error) {
	decrypted, err := l.svcCtx.WxPayClient.HandleCallback(
		in.WechatpaySignature, in.WechatpayNonce, in.WechatpayTimestamp, in.WechatpaySerial, in.Body,
	)
	if err != nil {
		l.Logger.Errorf("回调验证失败: %s", err.Error())
		return &order.PaymentCallbackResponse{
			Code:    "FAIL",
			Message: "签名验证失败",
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	record, err := l.svcCtx.PaymentRecordModel.FindByOrderNo(decrypted.OutTradeNo)
	if err != nil || record == nil {
		l.Logger.Errorf("支付记录不存在: %s", decrypted.OutTradeNo)
		return &order.PaymentCallbackResponse{
			Code:    "FAIL",
			Message: "支付记录不存在",
			Success: false,
			Msg:     "payment record not found",
		}, nil
	}

	if record.TradeState == "SUCCESS" {
		return &order.PaymentCallbackResponse{
			Code:          "SUCCESS",
			Message:       "成功",
			OrderNo:       decrypted.OutTradeNo,
			TransactionId: decrypted.TransactionID,
			TradeState:    "SUCCESS",
			Success:       true,
		}, nil
	}

	if decrypted.TradeState != "SUCCESS" {
		l.Logger.Infof("支付状态非成功: %s, %s", decrypted.OutTradeNo, decrypted.TradeState)
		return &order.PaymentCallbackResponse{
			Code:       "SUCCESS",
			Message:    "成功",
			OrderNo:    decrypted.OutTradeNo,
			TradeState: decrypted.TradeState,
			Success:    true,
		}, nil
	}

	payTime, _ := time.Parse(time.RFC3339, decrypted.SuccessTime)
	if err := l.svcCtx.OrderMainModel.UpdatePayStatus(decrypted.OutTradeNo, payTime); err != nil {
		l.Logger.Errorf("更新订单支付状态失败: %s", err.Error())
	}

	payerTotal := int64(decrypted.Amount.PayerTotal)
	record.TransactionId = decrypted.TransactionID
	record.TradeState = "SUCCESS"
	record.TradeType = decrypted.TradeType
	record.BankType = decrypted.BankType
	record.PayerTotal = payerTotal
	record.NotifyJson = in.Body
	record.NotifyTime = time.Now()
	if err := l.svcCtx.PaymentRecordModel.Update(record); err != nil {
		l.Logger.Errorf("更新支付记录失败: %s", err.Error())
	}

	l.Logger.Infof("支付成功, orderNo: %s, transactionId: %s, amount: %d",
		decrypted.OutTradeNo, decrypted.TransactionID, payerTotal)

	return &order.PaymentCallbackResponse{
		Code:          "SUCCESS",
		Message:       "成功",
		OrderNo:       decrypted.OutTradeNo,
		TransactionId: decrypted.TransactionID,
		TotalFee:      payerTotal,
		TradeState:    "SUCCESS",
		Success:       true,
	}, nil
}
