package alipay

import (
	"os"
	"strings"

	alipayv3 "github.com/go-pay/gopay/alipay/v3"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/config"
)

type Client struct {
	client *alipayv3.ClientV3
	cfg    config.AlipayConfig
}

func NewClient(cfg config.AlipayConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	privateKeyBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeAlipayCertError).WithDetail("path", cfg.PrivateKeyPath)
	}

	// gopay's FormatAlipayPrivateKey adds PEM headers, so strip them first
	privateKey := strings.TrimSpace(string(privateKeyBytes))
	privateKey = strings.TrimPrefix(privateKey, "-----BEGIN RSA PRIVATE KEY-----")
	privateKey = strings.TrimPrefix(privateKey, "-----BEGIN PRIVATE KEY-----")
	privateKey = strings.TrimSuffix(privateKey, "-----END RSA PRIVATE KEY-----")
	privateKey = strings.TrimSuffix(privateKey, "-----END PRIVATE KEY-----")
	privateKey = strings.ReplaceAll(privateKey, "\n", "")
	privateKey = strings.TrimSpace(privateKey)

	client, err := alipayv3.NewClientV3(cfg.AppID, privateKey, cfg.IsProd)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeAlipayCertError)
	}

	if cfg.AppCertPath != "" && cfg.RootCertPath != "" && cfg.PublicKeyPath != "" {
		appCert, err := os.ReadFile(cfg.AppCertPath)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.CodeAlipayCertError)
		}
		rootCert, err := os.ReadFile(cfg.RootCertPath)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.CodeAlipayCertError)
		}
		publicCert, err := os.ReadFile(cfg.PublicKeyPath)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.CodeAlipayCertError)
		}
		if err := client.SetCert(appCert, rootCert, publicCert); err != nil {
			return nil, apperror.Wrap(err, apperror.CodeAlipayCertError)
		}
	}

	return &Client{client: client, cfg: cfg}, nil
}

func (c *Client) IsAvailable() bool {
	return c != nil && c.client != nil
}

func (c *Client) V3() *alipayv3.ClientV3 {
	return c.client
}

func (c *Client) NotifyURL() string {
	return c.cfg.NotifyURL
}

func (c *Client) PublicKey() string {
	if c.cfg.PublicKeyPath == "" {
		return ""
	}
	data, _ := os.ReadFile(c.cfg.PublicKeyPath)
	return string(data)
}
