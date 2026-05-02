package server

import (
	"context"

	"github.com/xxx-newbee/order/internal/logic"
	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"
)

type OrderServer struct {
	svcCtx *svc.ServiceContext
	order.UnimplementedOrderServer
}

func NewOrderServer(svcCtx *svc.ServiceContext) *OrderServer {
	return &OrderServer{
		svcCtx: svcCtx,
	}
}

func (s *OrderServer) SeckillOrder(ctx context.Context, in *order.SeckillOrderRequest) (*order.SeckillOrderResponse, error) {
	l := logic.NewSeckillOrderLogic(ctx, s.svcCtx)
	return l.SeckillOrder(in)
}

func (s *OrderServer) CreateSeckillActivity(ctx context.Context, in *order.CreateSeckillActivityRequest) (*order.CreateSeckillActivityResponse, error) {
	l := logic.NewCreateSeckillActivityLogic(ctx, s.svcCtx)
	return l.CreateSeckillActivity(in)
}

func (s *OrderServer) GetSeckillActivity(ctx context.Context, in *order.GetSeckillActivityRequest) (*order.GetSeckillActivityResponse, error) {
	l := logic.NewGetSeckillActivityLogic(ctx, s.svcCtx)
	return l.GetSeckillActivity(in)
}

func (s *OrderServer) LoadSeckillStock(ctx context.Context, in *order.LoadSeckillStockRequest) (*order.LoadSeckillStockResponse, error) {
	l := logic.NewLoadSeckillStockLogic(ctx, s.svcCtx)
	return l.LoadSeckillStock(in)
}

func (s *OrderServer) GetSeckillOrder(ctx context.Context, in *order.GetSeckillOrderRequest) (*order.GetSeckillOrderResponse, error) {
	l := logic.NewGetSeckillOrderLogic(ctx, s.svcCtx)
	return l.GetSeckillOrder(in)
}

func (s *OrderServer) GetUserSeckillOrders(ctx context.Context, in *order.GetUserSeckillOrdersRequest) (*order.GetUserSeckillOrdersResponse, error) {
	l := logic.NewGetUserSeckillOrdersLogic(ctx, s.svcCtx)
	return l.GetUserSeckillOrders(in)
}

func (s *OrderServer) CancelTimeoutOrder(ctx context.Context, in *order.CancelTimeoutOrderRequest) (*order.CancelTimeoutOrderResponse, error) {
	l := logic.NewCancelTimeoutOrderLogic(ctx, s.svcCtx)
	return l.CancelTimeoutOrder(in)
}

func (s *OrderServer) CreatePayment(ctx context.Context, in *order.CreatePaymentRequest) (*order.CreatePaymentResponse, error) {
	l := logic.NewCreatePaymentLogic(ctx, s.svcCtx)
	return l.CreatePayment(in)
}

func (s *OrderServer) ProcessPaymentCallback(ctx context.Context, in *order.PaymentCallbackRequest) (*order.PaymentCallbackResponse, error) {
	l := logic.NewProcessPaymentCallbackLogic(ctx, s.svcCtx)
	return l.ProcessPaymentCallback(in)
}

func (s *OrderServer) QueryPayment(ctx context.Context, in *order.QueryPaymentRequest) (*order.QueryPaymentResponse, error) {
	l := logic.NewQueryPaymentLogic(ctx, s.svcCtx)
	return l.QueryPayment(in)
}

func (s *OrderServer) Refund(ctx context.Context, in *order.RefundRequest) (*order.RefundResponse, error) {
	l := logic.NewRefundLogic(ctx, s.svcCtx)
	return l.Refund(in)
}

func (s *OrderServer) CloseOrder(ctx context.Context, in *order.CloseOrderRequest) (*order.CloseOrderResponse, error) {
	l := logic.NewClosePaymentLogic(ctx, s.svcCtx)
	return l.CloseOrder(in)
}
