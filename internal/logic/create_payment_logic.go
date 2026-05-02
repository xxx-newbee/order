package logic

import (
	"context"
	"fmt"
	"math"

	"github.com/xxx-newbee/order/internal/model"
	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePaymentLogic) CreatePayment(in *order.CreatePaymentRequest) (*order.CreatePaymentResponse, error) {
	orderMain, err := l.svcCtx.OrderMainModel.FindByOrderNo(in.OrderNo)
	if err != nil {
		return &order.CreatePaymentResponse{Success: false, Msg: "订单查询失败"}, nil
	}
	if orderMain == nil {
		return &order.CreatePaymentResponse{Success: false, Msg: "订单不存在"}, nil
	}
	if orderMain.OrderStatus != 0 {
		return &order.CreatePaymentResponse{Success: false, Msg: "订单状态不允许支付"}, nil
	}

	existing, err := l.svcCtx.PaymentRecordModel.FindByOrderNo(in.OrderNo)
	if err != nil {
		l.Logger.Errorf("查询支付记录失败: %s", err.Error())
	}
	if existing != nil && existing.PrepayId != "" {
		return &order.CreatePaymentResponse{
			Success:  true,
			PrepayId: existing.PrepayId,
			PayInfo:  "",
			Msg:      "支付已发起，返回已有记录",
		}, nil
	}

	amountCents := yuanToCents(orderMain.PayAmount)
	desc := fmt.Sprintf("订单-%s", in.OrderNo)
	var payInfo string
	var prepayID string

	switch in.PayType {
	case "JSAPI":
		if in.Openid == "" {
			return &order.CreatePaymentResponse{Success: false, Msg: "JSAPI支付需要openid"}, nil
		}
		resp, params, e := l.svcCtx.WxPayClient.PlaceJSAPIOrder(desc, in.OrderNo, in.Openid, amountCents)
		if e != nil {
			l.Logger.Errorf("JSAPI下单失败: %s", e.Error())
			return &order.CreatePaymentResponse{Success: false, Msg: "微信支付下单失败: " + e.Error()}, nil
		}
		prepayID = resp.PrepayID
		payInfo = params
	case "NATIVE":
		resp, e := l.svcCtx.WxPayClient.PlaceNativeOrder(desc, in.OrderNo, amountCents)
		if e != nil {
			l.Logger.Errorf("Native下单失败: %s", e.Error())
			return &order.CreatePaymentResponse{Success: false, Msg: "微信支付下单失败: " + e.Error()}, nil
		}
		payInfo = resp.CodeURL
	case "APP":
		resp, params, e := l.svcCtx.WxPayClient.PlaceAppOrder(desc, in.OrderNo, amountCents)
		if e != nil {
			l.Logger.Errorf("App下单失败: %s", e.Error())
			return &order.CreatePaymentResponse{Success: false, Msg: "微信支付下单失败: " + e.Error()}, nil
		}
		prepayID = resp.PrepayID
		payInfo = params
	case "H5":
		if in.ClientIp == "" {
			return &order.CreatePaymentResponse{Success: false, Msg: "H5支付需要client_ip"}, nil
		}
		resp, e := l.svcCtx.WxPayClient.PlaceH5Order(desc, in.OrderNo, in.ClientIp, amountCents)
		if e != nil {
			l.Logger.Errorf("H5下单失败: %s", e.Error())
			return &order.CreatePaymentResponse{Success: false, Msg: "微信支付下单失败: " + e.Error()}, nil
		}
		payInfo = resp.H5URL
	default:
		return &order.CreatePaymentResponse{Success: false, Msg: "不支持的支付类型: " + in.PayType}, nil
	}

	record := &model.PaymentRecord{
		OrderNo:    in.OrderNo,
		TradeType:  in.PayType,
		TradeState: "NOTPAY",
		PayAmount:  orderMain.PayAmount,
		PrepayId:   prepayID,
	}
	if err := l.svcCtx.PaymentRecordModel.Insert(record); err != nil {
		l.Logger.Errorf("保存支付记录失败: %s", err.Error())
		return &order.CreatePaymentResponse{Success: false, Msg: "保存支付记录失败"}, nil
	}

	return &order.CreatePaymentResponse{
		Success:  true,
		PrepayId: prepayID,
		PayInfo:  payInfo,
		Msg:      "支付下单成功",
	}, nil
}

func yuanToCents(yuan float64) int {
	return int(math.Round(yuan * 100))
}
