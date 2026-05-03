package wechat

import (
	"os"

	"github.com/go-pay/gopay"
	wechatv3 "github.com/go-pay/gopay/wechat/v3"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/config"
)

type Client struct {
	client *wechatv3.ClientV3
	cfg    config.WechatConfig
}

func NewClient(cfg config.WechatConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	privateKey, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWeChatCertError).WithDetail("path", cfg.PrivateKeyPath)
	}

	client, err := wechatv3.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIV3Key, string(privateKey))
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWeChatCertError)
	}

	if cfg.PublicKeyPath != "" {
		wxPublicKey, err := os.ReadFile(cfg.PublicKeyPath)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.CodeWeChatCertError).WithDetail("path", cfg.PublicKeyPath)
		}
		if err := client.AutoVerifySignByCert(wxPublicKey, cfg.PublicKeyID); err != nil {
			return nil, apperror.Wrap(err, apperror.CodeWeChatCertError)
		}
	}

	client.DebugSwitch = gopay.DebugOff
	return &Client{client: client, cfg: cfg}, nil
}

func (c *Client) IsAvailable() bool {
	return c != nil && c.client != nil
}

func (c *Client) V3() *wechatv3.ClientV3 {
	return c.client
}

func (c *Client) NotifyURL() string {
	return c.cfg.NotifyURL
}
