package robokassa

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

type Config struct {
	MerchantLogin string
	Password1     string
	Password2     string
	IsTest        bool
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

// GeneratePaymentURL generates a URL to redirect the user to Robokassa.
func (c *Client) GeneratePaymentURL(outSum float64, invID int64, description string) string {
	sumStr := fmt.Sprintf("%.2f", outSum)

	// Create signature: MerchantLogin:OutSum:InvId:Password1
	signatureRaw := fmt.Sprintf("%s:%s:%d:%s", c.cfg.MerchantLogin, sumStr, invID, c.cfg.Password1)
	signature := generateMD5(signatureRaw)

	baseURL := "https://auth.robokassa.ru/Merchant/Index.aspx"
	isTest := "0"
	if c.cfg.IsTest {
		isTest = "1"
	}

	return fmt.Sprintf("%s?MerchantLogin=%s&OutSum=%s&InvId=%d&Description=%s&SignatureValue=%s&IsTest=%s",
		baseURL,
		c.cfg.MerchantLogin,
		sumStr,
		invID,
		description,
		signature,
		isTest,
	)
}

// CheckResultSignature validates the request from Robokassa on the ResultURL.
// The signature format is: OutSum:InvId:Password2
func (c *Client) CheckResultSignature(outSum string, invID int64, signatureValue string) bool {
	signatureRaw := fmt.Sprintf("%s:%d:%s", outSum, invID, c.cfg.Password2)
	expectedSignature := generateMD5(signatureRaw)

	return strings.EqualFold(signatureValue, expectedSignature)
}

// CheckSuccessSignature validates the request from Robokassa when redirecting user to SuccessURL.
// The signature format is: OutSum:InvId:Password1
func (c *Client) CheckSuccessSignature(outSum string, invID int64, signatureValue string) bool {
	signatureRaw := fmt.Sprintf("%s:%d:%s", outSum, invID, c.cfg.Password1)
	expectedSignature := generateMD5(signatureRaw)

	return strings.EqualFold(signatureValue, expectedSignature)
}

func generateMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
