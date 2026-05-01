package logic

import (
	"context"
	"fmt"

	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelTimeoutOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelTimeoutOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelTimeoutOrderLogic {
	return &CancelTimeoutOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelTimeoutOrder 取消超时未支付的秒杀订单
// 1. 校验订单状态（只有status=5秒杀中的订单可取消）
// 2. 更新订单状态为取消(4)
// 3. 回滚Redis库存计数器
// 4. 异步回滚数据库库存
func (l *CancelTimeoutOrderLogic) CancelTimeoutOrder(in *order.CancelTimeoutOrderRequest) (*order.CancelTimeoutOrderResponse, error) {
	if in.OrderNo == "" {
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "订单号不能为空"}, nil
	}

	orderMain, err := l.svcCtx.OrderMainModel.FindByOrderNo(in.OrderNo)
	if err != nil {
		l.Logger.Errorf("查询订单失败: %v", err)
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "查询失败"}, nil
	}
	if orderMain == nil {
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "订单不存在"}, nil
	}

	// 只有秒杀中状态的订单可以超时取消
	if orderMain.OrderStatus != 5 {
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "订单状态不允许取消"}, nil
	}
	if orderMain.OrderType != 1 {
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "非秒杀订单"}, nil
	}

	// 更新订单为取消状态
	if err := l.svcCtx.OrderMainModel.UpdateStatus(in.OrderNo, 4); err != nil {
		l.Logger.Errorf("取消订单失败: %v", err)
		return &order.CancelTimeoutOrderResponse{Success: false, Msg: "取消订单失败"}, nil
	}

	// 回滚Redis库存
	stockKey := fmt.Sprintf("seckill:stock:%d", orderMain.SeckillActivityId)
	if _, err := l.svcCtx.Cache.Increase(stockKey); err != nil {
		l.Logger.Errorf("Redis库存回滚失败: %v", err)
	}

	// 数据库库存回滚（乐观锁）
	if affected, err := l.svcCtx.SeckillStockModel.IncreaseStock(orderMain.SeckillActivityId); err != nil || affected == 0 {
		l.Logger.Errorf("DB库存回滚失败, activityId=%d, affected=%d, err=%v",
			orderMain.SeckillActivityId, affected, err)
	}

	l.Logger.Infof("超时取消订单成功, orderNo=%s, userId=%d", in.OrderNo, orderMain.UserId)
	return &order.CancelTimeoutOrderResponse{
		Success: true,
		Msg:     "取消成功",
	}, nil
}
