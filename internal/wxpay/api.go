package wxpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

func (c *Client) PlaceJSAPIOrder(description, outTradeNo, openID string, amountCents int) (*PaymentResp, string, error) {
	req := c.buildPayReq(description, outTradeNo, amountCents)
	req.Payer = &Payer{OpenID: openID}

	var resp PaymentResp
	if err := c.post("/v3/pay/transactions/jsapi", req, &resp); err != nil {
		return nil, "", err
	}
	payParams := buildJSAPIParams(c.appID, resp.PrepayID, c.signer.privateKey, c.mchID, c.signer.certSerialNo)
	return &resp, payParams, nil
}

func (c *Client) PlaceAppOrder(description, outTradeNo string, amountCents int) (*PaymentResp, string, error) {
	req := c.buildPayReq(description, outTradeNo, amountCents)

	var resp PaymentResp
	if err := c.post("/v3/pay/transactions/app", req, &resp); err != nil {
		return nil, "", err
	}
	payParams := buildAppParams(c.appID, resp.PrepayID, c.signer.privateKey, c.mchID, c.signer.certSerialNo)
	return &resp, payParams, nil
}

func (c *Client) PlaceNativeOrder(description, outTradeNo string, amountCents int) (*NativeResp, error) {
	req := c.buildPayReq(description, outTradeNo, amountCents)

	var resp NativeResp
	if err := c.post("/v3/pay/transactions/native", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) PlaceH5Order(description, outTradeNo, clientIP string, amountCents int) (*H5Resp, error) {
	req := c.buildPayReq(description, outTradeNo, amountCents)
	req.SceneInfo = &SceneInfo{
		PayerClientIP: clientIP,
		H5Info:        &H5Info{Type: "Wap"},
	}

	var resp H5Resp
	if err := c.post("/v3/pay/transactions/h5", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) QueryByOutTradeNo(orderNo string) (*OrderQueryResp, error) {
	path := "/v3/pay/transactions/out-trade-no/" + orderNo + "?mchid=" + c.mchID
	var resp OrderQueryResp
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) QueryByTransactionID(transactionID string) (*OrderQueryResp, error) {
	path := "/v3/pay/transactions/id/" + transactionID + "?mchid=" + c.mchID
	var resp OrderQueryResp
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CloseOrder(orderNo string) error {
	req := map[string]string{"mchid": c.mchID}
	path := "/v3/pay/transactions/out-trade-no/" + orderNo + "/close"
	return c.post(path, &req, nil)
}

func (c *Client) Refund(refundReq *RefundReq) (*RefundResp, error) {
	var resp RefundResp
	if err := c.post("/v3/refund/domestic/refunds", refundReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) QueryRefund(outRefundNo string) (*RefundResp, error) {
	path := "/v3/refund/domestic/refunds/" + outRefundNo
	var resp RefundResp
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func buildJSAPIParams(appID, prepayID string, pk *rsa.PrivateKey, mchID, serialNo string) string {
	nonce := generateNonce()
	ts := time.Now().Unix()
	pkg := "prepay_id=" + prepayID

	msg := appID + "\n" + strconv.FormatInt(ts, 10) + "\n" + nonce + "\n" + pkg + "\n"
	sig, _ := signSHA256RSA(pk, msg)

	return fmt.Sprintf(`{"appId":"%s","timeStamp":"%d","nonceStr":"%s","package":"%s","signType":"RSA_256","paySign":"%s","signature":"%s","mchid":"%s","serialNo":"%s"}`,
		appID, ts, nonce, pkg, sig, sig, mchID, serialNo)
}

func buildAppParams(appID, prepayID string, pk *rsa.PrivateKey, mchID, serialNo string) string {
	nonce := generateNonce()
	ts := time.Now().Unix()

	msg := appID + "\n" + strconv.FormatInt(ts, 10) + "\n" + nonce + "\n" + prepayID + "\n"
	h := sha256.Sum256([]byte(msg))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, h[:])
	signature := base64.StdEncoding.EncodeToString(sig)

	return fmt.Sprintf(`{"appid":"%s","partnerid":"%s","prepayid":"%s","package":"Sign=WXPay","noncestr":"%s","timestamp":"%d","sign":"%s"}`,
		appID, mchID, prepayID, nonce, ts, signature)
}
