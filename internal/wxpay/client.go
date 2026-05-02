package wxpay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const userAgent = "go-micro-wxpay/1.0"

type Client struct {
	mchID        string
	appID        string
	notifyURL    string
	httpClient   *http.Client
	signer       *Signer
	verifier     *Verifier
	certManager  *CertManager
	cipher       *AESCipher
}

type Config struct {
	MchID          string
	AppID          string
	MchAPIv3Key    string
	PrivateKeyPath string
	CertSerialNo   string
	NotifyURL      string
}

func NewClient(cfg Config) (*Client, error) {
	signer, err := NewSigner(cfg.MchID, cfg.CertSerialNo, cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("init signer: %w", err)
	}

	cipher, err := NewAESCipher(cfg.MchAPIv3Key)
	if err != nil {
		return nil, fmt.Errorf("init cipher: %w", err)
	}

	certManager := NewCertManager(signer, cipher)

	hc := &http.Client{Timeout: Timeout}

	cl := &Client{
		mchID:       cfg.MchID,
		appID:       cfg.AppID,
		notifyURL:   cfg.NotifyURL,
		httpClient:  hc,
		signer:      signer,
		cipher:      cipher,
		certManager: certManager,
		verifier:    NewVerifier(certManager),
	}

	certManager.setHTTPDoer(cl.do)
	return cl, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) post(path string, body interface{}, respObj interface{}) error {
	return c.request(http.MethodPost, path, body, respObj)
}

func (c *Client) get(path string, respObj interface{}) error {
	return c.request(http.MethodGet, path, nil, respObj)
}

func (c *Client) request(method, path string, reqBody interface{}, respObj interface{}) error {
	var bodyStr string
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyStr = string(b)
	}

	url := BaseURL + path
	req, err := http.NewRequest(method, url, strings.NewReader(bodyStr))
	if err != nil {
		return err
	}

	auth, err := c.signer.Sign(method, path, bodyStr)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr APIErr
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Code != "" {
			return &apiErr
		}
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(respBody))
	}

	if respObj != nil {
		return json.Unmarshal(respBody, respObj)
	}
	return nil
}

func (c *Client) HandleCallback(signature, nonce, timestamp, serial, body string) (*CallbackDecrypted, error) {
	if err := c.verifier.Verify(signature, nonce, timestamp, serial, body); err != nil {
		return nil, fmt.Errorf("verify callback: %w", err)
	}

	var cbReq CallbackReq
	if err := json.Unmarshal([]byte(body), &cbReq); err != nil {
		return nil, fmt.Errorf("parse callback body: %w", err)
	}
	if cbReq.Resource == nil {
		return nil, fmt.Errorf("empty resource in callback")
	}

	plain, err := c.cipher.Decrypt(cbReq.Resource.Ciphertext, cbReq.Resource.Nonce, cbReq.Resource.AssociatedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt resource: %w", err)
	}

	var decrypted CallbackDecrypted
	if err := json.Unmarshal(plain, &decrypted); err != nil {
		return nil, fmt.Errorf("parse decrypted resource: %w", err)
	}
	return &decrypted, nil
}

func (c *Client) buildPayReq(description, outTradeNo string, amountCents int) *PaymentReq {
	return &PaymentReq{
		AppID:       c.appID,
		MchID:       c.mchID,
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyURL:   c.notifyURL,
		Amount:      &Amount{Total: amountCents, Currency: "CNY"},
	}
}

func (c *Client) buildRefundReq(outTradeNo, refundNo, reason string, refundCents, totalCents int) *RefundReq {
	return &RefundReq{
		OutTradeNo:  outTradeNo,
		OutRefundNo: refundNo,
		Reason:      reason,
		Amount: &RefundAmount{
			Refund:   refundCents,
			Total:    totalCents,
			Currency: "CNY",
		},
	}
}
