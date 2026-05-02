package logic

import (
	"context"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPaymentLogic {
	return &QueryPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryPaymentLogic) QueryPayment(in *order.QueryPaymentRequest) (*order.QueryPaymentResponse, error) {
	var wxResp *struct {
		OrderNo       string
		TransactionID string
		TradeState    string
		TotalFee      int64
		BankType      string
		TradeType     string
	}

	if in.TransactionId != "" {
		resp, err := l.svcCtx.WxPayClient.QueryByTransactionID(in.TransactionId)
		if err != nil {
			l.Logger.Errorf("查询微信支付失败: %s", err.Error())
			return &order.QueryPaymentResponse{Success: false, Msg: err.Error()}, nil
		}
		wxResp = &struct {
			OrderNo       string
			TransactionID string
			TradeState    string
			TotalFee      int64
			BankType      string
			TradeType     string
		}{
			OrderNo:       resp.OutTradeNo,
			TransactionID: resp.TransactionID,
			TradeState:    resp.TradeState,
			BankType:      resp.BankType,
			TradeType:     resp.TradeType,
		}
		if resp.Amount != nil {
			wxResp.TotalFee = int64(resp.Amount.Total)
		}
	} else if in.OrderNo != "" {
		resp, err := l.svcCtx.WxPayClient.QueryByOutTradeNo(in.OrderNo)
		if err != nil {
			l.Logger.Errorf("查询微信支付失败: %s", err.Error())
			return &order.QueryPaymentResponse{Success: false, Msg: err.Error()}, nil
		}
		wxResp = &struct {
			OrderNo       string
			TransactionID string
			TradeState    string
			TotalFee      int64
			BankType      string
			TradeType     string
		}{
			OrderNo:       resp.OutTradeNo,
			TransactionID: resp.TransactionID,
			TradeState:    resp.TradeState,
			BankType:      resp.BankType,
			TradeType:     resp.TradeType,
		}
		if resp.Amount != nil {
			wxResp.TotalFee = int64(resp.Amount.Total)
		}
	} else {
		return &order.QueryPaymentResponse{Success: false, Msg: "order_no和transaction_id至少提供一个"}, nil
	}

	if wxResp.TradeState == "SUCCESS" {
		record, _ := l.svcCtx.PaymentRecordModel.FindByOrderNo(wxResp.OrderNo)
		if record != nil && record.TradeState != "SUCCESS" {
			record.TransactionId = wxResp.TransactionID
			record.TradeState = "SUCCESS"
			_ = l.svcCtx.PaymentRecordModel.Update(record)
		}
	}

	return &order.QueryPaymentResponse{
		Success:       true,
		OrderNo:       wxResp.OrderNo,
		TransactionId: wxResp.TransactionID,
		TradeState:    wxResp.TradeState,
		TotalFee:      wxResp.TotalFee,
		BankType:      wxResp.BankType,
		TradeType:     wxResp.TradeType,
	}, nil
}
