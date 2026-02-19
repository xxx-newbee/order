package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	SeckillStock struct {
		Id           int64     `db:"id"`            // ID
		ActivityId   int64     `db:"activity_id"`   // 秒杀活动ID
		ProductId    int64     `db:"product_id"`    // 商品ID
		SurplusStock int       `db:"surplus_stock"` // 剩余库存
		Version      int       `db:"version"`       // 乐观锁版本号
		UpdateTime   time.Time `db:"update_time"`   // 更新时间
	}

	SeckillStockModel interface {
		FindByActivityId(activityId int64) (*SeckillStock, error)
		DecreaseStock(activityId int64) (int64, error)
		Update(data *SeckillStock) error
	}

	defaultSeckillStock struct {
		db    *gorm.DB
		table string
	}
)

func (s *SeckillStock) TableName() string { return "order_seckill_stock" }

func NewSeckillStockModel(db *gorm.DB) SeckillStockModel {
	return &defaultSeckillStock{
		db:    db,
		table: "order_seckill_stock",
	}
}

func (m *defaultSeckillStock) FindByActivityId(activityId int64) (*SeckillStock, error) {
	var data SeckillStock
	res := m.db.Table(m.table).Where("activity_id = ?", activityId).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &data, nil
}

// 乐观锁扣减库存，防超卖
func (m *defaultSeckillStock) DecreaseStock(activityId int64) (int64, error) {
	// 1.检查库存和版本号
	stock, err := m.FindByActivityId(activityId)
	if err != nil {
		return 0, err
	}
	if stock == nil || stock.SurplusStock <= 0 {
		return 0, errors.New("库存不足")
	}

	// 2.乐观锁更新
	res := m.db.Table(m.table).Update("surplus_stock", stock.SurplusStock-1).Update("version", stock.Version+1).Update("update_time", time.Now()).Where("activity_id = ? AND version = ? AND surplus_stock > 0", activityId, stock.Version)
	return res.RowsAffected, res.Error
}

func (m *defaultSeckillStock) Update(data *SeckillStock) error {
	res := m.db.Table(m.table).Where("activity_id = ?", data.ActivityId).Updates(&data)
	return res.Error
}
