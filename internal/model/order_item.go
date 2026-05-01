package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	OrderItem struct {
		Id           int64     `db:"id"`            // 明细ID
		OrderId      int64     `db:"order_id"`      // 订单ID
		OrderNo      string    `db:"order_no"`      // 订单编号
		ProductId    int64     `db:"product_id"`    // 商品ID
		ProductName  string    `db:"product_name"`  // 商品名称
		ProductPrice float64   `db:"product_price"` // 商品单价
		BuyNum       int       `db:"buy_num"`       // 购买数量
		CreateTime   time.Time `db:"create_time"`   // 创建时间
	}

	OrderItemModel interface {
		Insert(data *OrderItem) error
		InsertBatch(data []*OrderItem) error
		FindById(id int64) (*OrderItem, error)
		FindByOrderNo(orderNo string) ([]*OrderItem, error)
	}

	defaultOrderItem struct {
		db    *gorm.DB
		table string
	}
)

func (OrderItem) TableName() string {
	return "order_item"
}

func NewOrderItemModel(db *gorm.DB) OrderItemModel {
	return &defaultOrderItem{
		db:    db,
		table: "order_item",
	}
}

func (m *defaultOrderItem) Insert(data *OrderItem) error {
	return m.db.Create(data).Error
}

func (m *defaultOrderItem) InsertBatch(data []*OrderItem) error {
	return m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&data).Error
}

func (m *defaultOrderItem) FindById(id int64) (*OrderItem, error) {
	var item OrderItem
	err := m.db.Where("id = ?", id).First(&item).Error
	return &item, err
}

func (m *defaultOrderItem) FindByOrderNo(orderNo string) ([]*OrderItem, error) {
	var items []*OrderItem
	err := m.db.Where("order_no = ?", orderNo).Find(&items).Error
	return items, err
}
