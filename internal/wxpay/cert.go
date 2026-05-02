package wxpay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const certURL = BaseURL + "/v3/certificates"

type CertManager struct {
	mu    sync.RWMutex
	certs map[string]*x509.Certificate

	cipher    *AESCipher
	signer    *Signer
	httpDo    func(req *http.Request) (*http.Response, error)
}

func NewCertManager(signer *Signer, cipher *AESCipher) *CertManager {
	return &CertManager{
		certs:  make(map[string]*x509.Certificate),
		cipher: cipher,
		signer: signer,
		httpDo: (&http.Client{Timeout: Timeout}).Do,
	}
}

func (m *CertManager) setHTTPDoer(fn func(req *http.Request) (*http.Response, error)) {
	m.httpDo = fn
}

func (m *CertManager) GetBySerial(serial string) (*x509.Certificate, error) {
	m.mu.RLock()
	cert, ok := m.certs[serial]
	m.mu.RUnlock()
	if ok {
		return cert, nil
	}
	if err := m.refresh(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	cert, ok = m.certs[serial]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("platform cert not found for serial: %s", serial)
	}
	return cert, nil
}

func (m *CertManager) refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, err := http.NewRequest(http.MethodGet, certURL, nil)
	if err != nil {
		return err
	}

	auth, err := m.signer.Sign(http.MethodGet, "/v3/certificates", "")
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-micro-wxpay/1.0")

	resp, err := m.httpDo(req)
	if err != nil {
		return fmt.Errorf("fetch certificates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch certificates failed, status=%d, body=%s", resp.StatusCode, string(body))
	}

	var cr CertResp
	if err := json.Unmarshal(body, &cr); err != nil {
		return err
	}

	for _, d := range cr.Data {
		plain, err := m.cipher.Decrypt(d.EncryptCertificate.Ciphertext, d.EncryptCertificate.Nonce, d.EncryptCertificate.AssociatedData)
		if err != nil {
			return fmt.Errorf("decrypt cert: %w", err)
		}
		cert, err := x509.ParseCertificate(plain)
		if err != nil {
			return fmt.Errorf("parse cert: %w", err)
		}
		if _, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
			continue
		}
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			continue
		}
		m.certs[d.SerialNo] = cert
	}
	return nil
}
