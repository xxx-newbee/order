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
		Insert(data *SeckillActivity) (uint, error)
		Update(data *SeckillActivity) error
		GetActivityById(id uint) (*SeckillActivity, error)
		GetActivityByProductId(productId int64) (*SeckillActivity, error)
		FindByStatus(status int) ([]*SeckillActivity, error)
		List(page, pageSize int) ([]*SeckillActivity, int64, error)
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

func (m *defaultSeckillActivityModel) Insert(data *SeckillActivity) (uint, error) {
	if err := m.db.Create(data).Error; err != nil {
		return 0, err
	}
	return data.ID, nil
}

func (m *defaultSeckillActivityModel) Update(data *SeckillActivity) error {
	return m.db.Table(m.table).Where("id = ?", data.ID).Updates(data).Error
}

func (m *defaultSeckillActivityModel) GetActivityById(id uint) (*SeckillActivity, error) {
	var s SeckillActivity
	res := m.db.Table(m.table).Where("id = ?", id).First(&s)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
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
			return nil, nil
		}
		return nil, res.Error
	}
	return &s, nil
}

func (m *defaultSeckillActivityModel) FindByStatus(status int) ([]*SeckillActivity, error) {
	var list []*SeckillActivity
	res := m.db.Table(m.table).Where("status = ?", status).Find(&list)
	return list, res.Error
}

func (m *defaultSeckillActivityModel) List(page, pageSize int) ([]*SeckillActivity, int64, error) {
	var list []*SeckillActivity
	var total int64
	if err := m.db.Table(m.table).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	res := m.db.Table(m.table).Order("id DESC").Offset(offset).Limit(pageSize).Find(&list)
	return list, total, res.Error
}
