package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type (
	PaymentRecord struct {
		gorm.Model
		OrderNo       string    `db:"order_no" gorm:"uniqueIndex;size:64"`
		TransactionId string    `db:"transaction_id" gorm:"size:64"`
		TradeType     string    `db:"trade_type" gorm:"size:16"`
		TradeState    string    `db:"trade_state" gorm:"size:32"`
		PayAmount     float64   `db:"pay_amount"`
		PayerTotal    int64     `db:"payer_total"`
		PrepayId      string    `db:"prepay_id" gorm:"size:64"`
		BankType      string    `db:"bank_type" gorm:"size:32"`
		RefundId      string    `db:"refund_id" gorm:"size:64"`
		RefundStatus  string    `db:"refund_status" gorm:"size:32"`
		RefundAmount  float64   `db:"refund_amount"`
		NotifyJson    string    `db:"notify_json" gorm:"type:text"`
		NotifyTime    time.Time `db:"notify_time"`
	}

	PaymentRecordModel interface {
		Insert(data *PaymentRecord) error
		FindByOrderNo(orderNo string) (*PaymentRecord, error)
		FindByTransactionId(transactionId string) (*PaymentRecord, error)
		Update(data *PaymentRecord) error
	}

	defaultPaymentRecord struct {
		db    *gorm.DB
		table string
	}
)

func (PaymentRecord) TableName() string {
	return "payment_record"
}

func NewPaymentRecordModel(db *gorm.DB) PaymentRecordModel {
	return &defaultPaymentRecord{
		db:    db,
		table: "payment_record",
	}
}

func (m *defaultPaymentRecord) Insert(data *PaymentRecord) error {
	return m.db.Create(data).Error
}

func (m *defaultPaymentRecord) FindByOrderNo(orderNo string) (*PaymentRecord, error) {
	var data PaymentRecord
	res := m.db.Table(m.table).Where("order_no = ?", orderNo).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &data, nil
}

func (m *defaultPaymentRecord) FindByTransactionId(transactionId string) (*PaymentRecord, error) {
	var data PaymentRecord
	res := m.db.Table(m.table).Where("transaction_id = ?", transactionId).First(&data)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &data, nil
}

func (m *defaultPaymentRecord) Update(data *PaymentRecord) error {
	return m.db.Table(m.table).Where("id = ?", data.ID).Updates(data).Error
}
