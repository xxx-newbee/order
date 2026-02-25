package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	SeckillActivity struct {
		gorm.Model
		ProductId    int64     `db:"product_id"`    // 商品ID
		SeckillPrice float64   `db:"seckill_price"` // 秒杀价格
		StockNum     int       `db:"stock_num"`     // 秒杀库存总数
		StartTime    time.Time `db:"start_time"`
		EndTime      time.Time `db:"end_time"`
		Status       int       `db:"status"` // 0-未开始 1-进行中 2-已结束
	}

	SeckillActivityModel interface {
		GetActivityById(id uint) (*SeckillActivity, error)
		GetActivityByProductId(productId int64) (*SeckillActivity, error)
	}

	defaultSeckillActivityModel struct {
		db    *gorm.DB
		table string
	}
)

func (SeckillActivity) TableName() string {
	return "seckill_activity"
}

func NewSeckillActivityModel(db *gorm.DB) SeckillActivityModel {
	return &defaultSeckillActivityModel{
		db:    db,
		table: "seckill_activity",
	}
}

func (m *defaultSeckillActivityModel) GetActivityById(id uint) (*SeckillActivity, error) {
	var s SeckillActivity
	res := m.db.Table(m.table).Where("id = ?", id).First(&s)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return &SeckillActivity{}, nil
		}
		return nil, res.Error
	}
	return &s, nil
}

func (m *defaultSeckillActivityModel) GetActivityByProductId(productId int64) (*SeckillActivity, error) {
	var s SeckillActivity
	res := m.db.Table(m.table).Where("product_id = ?", productId).First(&s)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return &SeckillActivity{}, nil
		}
		return nil, res.Error
	}
	return &s, nil
}
