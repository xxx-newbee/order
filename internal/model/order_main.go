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
		OrderStatus       int8      `db:"order_status"`        // 订单状态 0-待支付 1-已付款 2-已发货 3-已完成 4-取消状态 5-秒杀中
		OrderType         int8      `db:"order_type"`          // 订单类型 0-普通订单 1-秒杀订单
		SeckillActivityId int64     `db:"seckill_activity_id"` // 秒杀活动ID
		PayTime           time.Time `db:"pay_time"`            // 支付时间
	}

	OrderMainModel interface {
		Insert(data *OrderMain) error
		FindById(id int64) (*OrderMain, error)
		FindByOrderNo(orderNo string) (*OrderMain, error)
		Update(data *OrderMain) error
		UpdateStatus(orderNo string, status int8) error
		FindByUserId(userId int64, page, pageSize int) ([]*OrderMain, int64, error)
		FindExpiredSeckillOrders(beforeTime string, limit int) ([]*OrderMain, error)
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
			return &OrderMain{}, nil
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
func (m *defaultOrderMain) UpdateStatus(orderNo string, status int8) error {
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

func (m *defaultOrderMain) FindByUserId(userId int64, page, pageSize int) ([]*OrderMain, int64, error) {
	var list []*OrderMain
	var total int64
	if err := m.db.Table(m.table).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	res := m.db.Table(m.table).Where("user_id = ?", userId).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list)
	return list, total, res.Error
}

// 查找超时未支付的秒杀订单（status=5 秒杀中，超过指定时间未支付）
func (m *defaultOrderMain) FindExpiredSeckillOrders(beforeTime string, limit int) ([]*OrderMain, error) {
	var list []*OrderMain
	res := m.db.Table(m.table).
		Where("order_status = ? AND order_type = ? AND created_at < ?", 5, 1, beforeTime).
		Order("created_at ASC").
		Limit(limit).
		Find(&list)
	return list, res.Error
}
