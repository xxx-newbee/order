package wxpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Signer struct {
	mchID        string
	certSerialNo string
	privateKey   *rsa.PrivateKey
}

func NewSigner(mchID, certSerialNo, privateKeyPath string) (*Signer, error) {
	pk, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &Signer{
		mchID:        mchID,
		certSerialNo: certSerialNo,
		privateKey:   pk,
	}, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func LoadPublicKeyFromBytes(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}
	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func (s *Signer) Sign(method, path, body string) (string, error) {
	nonce := generateNonce()
	ts := time.Now().Unix()

	msg := buildSignMsg(method, path, strconv.FormatInt(ts, 10), nonce, body)
	sig, err := signSHA256RSA(s.privateKey, msg)
	if err != nil {
		return "", err
	}
	return buildAuthHeader(s.mchID, s.certSerialNo, nonce, ts, sig), nil
}

func buildSignMsg(method, path, timestamp, nonce, body string) string {
	return method + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + body + "\n"
}

func signSHA256RSA(key *rsa.PrivateKey, msg string) (string, error) {
	h := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func buildAuthHeader(mchID, serialNo, nonce string, ts int64, sig string) string {
	return fmt.Sprintf(
		`%s mchid="%s",nonce_str="%s",timestamp="%d",serial_no="%s",signature="%s"`,
		Schema, mchID, nonce, ts, serialNo, sig,
	)
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
