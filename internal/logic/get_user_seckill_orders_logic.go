package logic

import (
	"context"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSeckillOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSeckillOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSeckillOrdersLogic {
	return &GetUserSeckillOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserSeckillOrdersLogic) GetUserSeckillOrders(in *order.GetUserSeckillOrdersRequest) (*order.GetUserSeckillOrdersResponse, error) {
	if in.UserId <= 0 {
		return &order.GetUserSeckillOrdersResponse{Success: false, Msg: "用户ID不合法"}, nil
	}

	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	orders, total, err := l.svcCtx.OrderMainModel.FindByUserId(in.UserId, page, pageSize)
	if err != nil {
		l.Logger.Errorf("查询用户订单失败: %v", err)
		return &order.GetUserSeckillOrdersResponse{Success: false, Msg: "查询失败"}, nil
	}

	var orderInfos []*order.SeckillOrderInfo
	for _, o := range orders {
		orderInfos = append(orderInfos, &order.SeckillOrderInfo{
			OrderNo:           o.OrderNo,
			UserId:            o.UserId,
			TotalAmount:       o.TotalAmount,
			PayAmount:         o.PayAmount,
			OrderStatus:       int32(o.OrderStatus),
			SeckillActivityId: o.SeckillActivityId,
			CreatedAt:         o.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &order.GetUserSeckillOrdersResponse{
		Orders:  orderInfos,
		Total:   total,
		Success: true,
		Msg:     "ok",
	}, nil
}
