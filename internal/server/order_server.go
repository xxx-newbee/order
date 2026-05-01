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
