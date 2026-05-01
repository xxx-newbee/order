package logic

import (
	"context"
	"time"

	"github.com/xxx-newbee/order/internal/model"
	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSeckillActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSeckillActivityLogic {
	return &CreateSeckillActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSeckillActivityLogic) CreateSeckillActivity(in *order.CreateSeckillActivityRequest) (*order.CreateSeckillActivityResponse, error) {
	if in.ProductId <= 0 || in.SeckillPrice <= 0 || in.StockNum <= 0 {
		return &order.CreateSeckillActivityResponse{Success: false, Msg: "参数不合法"}, nil
	}

	startTime := time.Unix(in.StartTime, 0)
	endTime := time.Unix(in.EndTime, 0)

	if !endTime.After(startTime) || time.Now().After(endTime) {
		return &order.CreateSeckillActivityResponse{Success: false, Msg: "活动时间不合法"}, nil
	}

	activity := &model.SeckillActivity{
		ProductId:    in.ProductId,
		SeckillPrice: in.SeckillPrice,
		StockNum:     int(in.StockNum),
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       0, // 未开始
	}

	id, err := l.svcCtx.SeckillActivity.Insert(activity)
	if err != nil {
		l.Logger.Errorf("创建秒杀活动失败: %v", err)
		return &order.CreateSeckillActivityResponse{Success: false, Msg: "创建活动失败"}, nil
	}

	// 同步初始化库存表
	stock := &model.SeckillStock{
		ActivityId:   int64(id),
		ProductId:    in.ProductId,
		SurplusStock: int(in.StockNum),
		Version:      0,
		UpdateTime:   time.Now(),
	}
	if err := l.svcCtx.SeckillStockModel.Update(stock); err != nil {
		l.Logger.Errorf("初始化库存记录失败: %v", err)
	}

	l.Logger.Infof("创建秒杀活动成功, activity_id=%d", id)
	return &order.CreateSeckillActivityResponse{
		ActivityId: uint32(id),
		Success:    true,
		Msg:        "创建成功",
	}, nil
}
