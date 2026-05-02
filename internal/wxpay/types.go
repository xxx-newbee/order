package wxpay

import "time"

const (
	BaseURL    = "https://api.mch.weixin.qq.com"
	Schema     = "WECHATPAY2-SHA256-RSA2048"
	APIVersion = "v3"
)

var Timeout = 10 * time.Second

type Amount struct {
	Total    int    `json:"total"`
	Currency string `json:"currency,omitempty"`
}

type Payer struct {
	OpenID string `json:"openid"`
}

type SceneInfo struct {
	PayerClientIP string  `json:"payer_client_ip"`
	H5Info        *H5Info `json:"h5_info,omitempty"`
}

type H5Info struct {
	Type string `json:"type"`
}

type StoreInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	AreaCode string `json:"area_code"`
	Address  string `json:"address"`
}

type Detail struct {
	CostPrice   int    `json:"cost_price,omitempty"`
	InvoiceID   string `json:"invoice_id,omitempty"`
	GoodsDetail []Good `json:"goods_detail,omitempty"`
}

type Good struct {
	MerchantGoodsID  string `json:"merchant_goods_id"`
	WechatpayGoodsID string `json:"wechatpay_goods_id,omitempty"`
	GoodsName        string `json:"goods_name,omitempty"`
	Quantity         int    `json:"quantity"`
	UnitPrice        int    `json:"unit_price"`
}

type SettleInfo struct {
	ProfitSharing bool `json:"profit_sharing"`
}

type PaymentReq struct {
	AppID       string      `json:"appid"`
	MchID       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	TimeExpire  string      `json:"time_expire,omitempty"`
	Attach      string      `json:"attach,omitempty"`
	NotifyURL   string      `json:"notify_url"`
	GoodsTag    string      `json:"goods_tag,omitempty"`
	SupportFapiao bool      `json:"support_fapiao,omitempty"`
	Amount      *Amount     `json:"amount"`
	Payer       *Payer      `json:"payer,omitempty"`
	SceneInfo   *SceneInfo  `json:"scene_info,omitempty"`
	Detail      *Detail     `json:"detail,omitempty"`
	SettleInfo  *SettleInfo `json:"settle_info,omitempty"`
}

type PaymentResp struct {
	PrepayID string `json:"prepay_id"`
}

type NativeResp struct {
	CodeURL string `json:"code_url"`
}

type H5Resp struct {
	H5URL string `json:"h5_url"`
}

type OrderQueryResp struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeType     string `json:"trade_type"`
	TradeState    string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	BankType      string `json:"bank_type"`
	Attach        string `json:"attach"`
	SuccessTime   string `json:"success_time"`
	Payer         *struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
	Amount *struct {
		Total       int    `json:"total"`
		PayerTotal  int    `json:"payer_total"`
		Currency    string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
}

type RefundReq struct {
	TransactionID string        `json:"transaction_id,omitempty"`
	OutTradeNo    string        `json:"out_trade_no,omitempty"`
	OutRefundNo   string        `json:"out_refund_no"`
	Reason        string        `json:"reason,omitempty"`
	NotifyURL     string        `json:"notify_url,omitempty"`
	FundsAccount  string        `json:"funds_account,omitempty"`
	Amount        *RefundAmount `json:"amount"`
	GoodsDetail   []RefundGood  `json:"goods_detail,omitempty"`
}

type RefundAmount struct {
	Refund   int    `json:"refund"`
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type RefundGood struct {
	MerchantGoodsID  string `json:"merchant_goods_id"`
	WechatpayGoodsID string `json:"wechatpay_goods_id,omitempty"`
	GoodsName        string `json:"goods_name,omitempty"`
	UnitPrice        int    `json:"unit_price"`
	RefundAmount     int    `json:"refund_amount"`
	RefundQuantity   int    `json:"refund_quantity"`
}

type RefundResp struct {
	RefundID     string `json:"refund_id"`
	OutRefundNo  string `json:"out_refund_no"`
	TransactionID string `json:"transaction_id"`
	OutTradeNo   string `json:"out_trade_no"`
	Channel      string `json:"channel"`
	Status       string `json:"status"`
	UserReceivedAccount string `json:"user_received_account"`
	SuccessTime  string `json:"success_time"`
	CreateTime   string `json:"create_time"`
	Amount       *RefundAmountResp `json:"amount"`
}

type RefundAmountResp struct {
	Total       int `json:"total"`
	Refund      int `json:"refund"`
	PayerTotal  int `json:"payer_total"`
	PayerRefund int `json:"payer_refund"`
}

type CallbackResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
}

type CallbackReq struct {
	ID           string            `json:"id"`
	CreateTime   string            `json:"create_time"`
	ResourceType string            `json:"resource_type"`
	EventType    string            `json:"event_type"`
	Summary      string            `json:"summary"`
	Resource     *CallbackResource `json:"resource"`
}

type CallbackDecrypted struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeType     string `json:"trade_type"`
	TradeState    string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	BankType      string `json:"bank_type"`
	Attach        string `json:"attach"`
	SuccessTime   string `json:"success_time"`
	Payer         *struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
	Amount *struct {
		Total       int    `json:"total"`
		PayerTotal  int    `json:"payer_total"`
		Currency    string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
}

type APIErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIErr) Error() string {
	return e.Code + ": " + e.Message
}

type CertResp struct {
	Data []CertData `json:"data"`
}

type CertData struct {
	SerialNo          string          `json:"serial_no"`
	EffectiveTime     string          `json:"effective_time"`
	ExpireTime        string          `json:"expire_time"`
	EncryptCertificate EncryptCert    `json:"encrypt_certificate"`
}

type EncryptCert struct {
	Algorithm      string `json:"algorithm"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	Ciphertext     string `json:"ciphertext"`
}
