package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	OrderMain struct {
		gorm.Model
		OrderNo           string    `db:"order_no"`            // 订单编号
		UserId            int64     `db:"user_id"`             // 用户ID
		TotalAmount       float64   `db:"total_amount"`        // 订单总金额
		PayAmount         float64   `db:"pay_amount"`          // 实付金额
		OrderStatus       int8      `db:"order_status"`        // 订单状态
		OrderType         int8      `db:"order_type"`          // 订单类型
		SeckillActivityId int64     `db:"seckill_activity_id"` // 秒杀活动ID
		PayTime           time.Time `db:"pay_time"`            // 支付时间
	}

	OrderMainModel interface {
		Insert(data *OrderMain) error
		FindById(id int64) (*OrderMain, error)
		FindByOrderNo(orderNo string) (*OrderMain, error)
		Update(data *OrderMain) error
		UpdataStatus(orderNo string, status int8) error
	}

	defaultOrderMain struct {
		db    *gorm.DB
		table string
	}
)

func (OrderMain) TableName() string {
	return "order_main"
}

func NewOrderMainModel(db *gorm.DB) OrderMainModel {
	return &defaultOrderMain{
		db:    db,
		table: "order_main",
	}
}

func (m *defaultOrderMain) Insert(data *OrderMain) error {
	return m.db.Create(data).Error
}

func (m *defaultOrderMain) FindById(id int64) (*OrderMain, error) {
	var data OrderMain
	res := m.db.Table(m.table).Where("id = ?", id).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &data, nil
}
func (m *defaultOrderMain) FindByOrderNo(orderNo string) (*OrderMain, error) {
	var data OrderMain
	res := m.db.Table(m.table).Where("order_no = ?", orderNo).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &data, nil
}
func (m *defaultOrderMain) Update(data *OrderMain) error {
	return m.db.Table(m.table).Where("id = ?", data.ID).Updates(data).Error
}
func (m *defaultOrderMain) UpdataStatus(orderNo string, status int8) error {
	var data OrderMain
	res := m.db.Table(m.table).Where("order_no = ?", orderNo).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return res.Error
	}

	data.OrderStatus = status

	return m.Update(&data)
}
