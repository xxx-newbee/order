package wxpay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

const maxTimestampDrift = 300

type Verifier struct {
	certManager *CertManager
}

func NewVerifier(cm *CertManager) *Verifier {
	return &Verifier{certManager: cm}
}

func (v *Verifier) Verify(signature, nonce, timestamp, serial, body string) error {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("parse timestamp: %w", err)
	}
	if time.Now().Unix()-ts > maxTimestampDrift || ts-time.Now().Unix() > maxTimestampDrift {
		return fmt.Errorf("timestamp expired: %d", ts)
	}

	cert, err := v.certManager.GetBySerial(serial)
	if err != nil {
		return fmt.Errorf("get platform cert: %w", err)
	}

	msg := timestamp + "\n" + nonce + "\n" + body + "\n"
	h := sha256.Sum256([]byte(msg))

	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("platform cert is not RSA")
	}

	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, h[:], sig); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

func VerifyResponse(signature, nonce, timestamp, serial, body string, pubKey *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("parse timestamp: %w", err)
	}
	if time.Now().Unix()-ts > maxTimestampDrift || ts-time.Now().Unix() > maxTimestampDrift {
		return fmt.Errorf("timestamp expired: %d", ts)
	}

	msg := timestamp + "\n" + nonce + "\n" + body + "\n"
	h := sha256.Sum256([]byte(msg))

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, h[:], sig); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}
