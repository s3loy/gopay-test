package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

// mockOrderRepo for payment tests
type mockOrderRepoForPayment struct {
	order   *entity.Order
	err     error
	updated bool
}

func (m *mockOrderRepoForPayment) Create(ctx context.Context, order *entity.Order) error {
	return m.err
}
func (m *mockOrderRepoForPayment) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	return m.order, m.err
}
func (m *mockOrderRepoForPayment) GetByOrderNo(ctx context.Context, no string) (*entity.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.order, nil
}
func (m *mockOrderRepoForPayment) Update(ctx context.Context, order *entity.Order) error {
	return m.err
}
func (m *mockOrderRepoForPayment) UpdateStatus(ctx context.Context, id uint64, s entity.OrderStatus) error {
	m.updated = true
	return m.err
}
func (m *mockOrderRepoForPayment) List(ctx context.Context, f repository.OrderFilter) ([]*entity.Order, int64, error) {
	return nil, 0, m.err
}

// mockPaymentRepo for payment tests
type mockPaymentRepoForPayment struct {
	payment       *entity.Payment
	err           error
	created       *entity.Payment
	updated       bool
	statusUpdated bool
}

func (m *mockPaymentRepoForPayment) Create(ctx context.Context, p *entity.Payment) error {
	m.created = p
	if m.err != nil {
		return m.err
	}
	p.ID = 1
	return nil
}
func (m *mockPaymentRepoForPayment) GetByID(ctx context.Context, id uint64) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentRepoForPayment) GetByPaymentNo(ctx context.Context, no string) (*entity.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.payment, nil
}
func (m *mockPaymentRepoForPayment) GetByThirdPartyNo(ctx context.Context, no string) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentRepoForPayment) GetByOrderID(ctx context.Context, id uint64) ([]*entity.Payment, error) {
	return nil, m.err
}
func (m *mockPaymentRepoForPayment) Update(ctx context.Context, p *entity.Payment) error {
	m.updated = true
	return m.err
}
func (m *mockPaymentRepoForPayment) UpdateStatus(ctx context.Context, id uint64, s entity.PaymentStatus) error {
	m.statusUpdated = true
	return m.err
}
func (m *mockPaymentRepoForPayment) List(ctx context.Context, f repository.PaymentFilter) ([]*entity.Payment, int64, error) {
	return nil, 0, m.err
}

// mockProviderFactory for payment tests
type mockProviderFactoryForPayment struct {
	provider service.PaymentProvider
	err      error
}

func (m *mockProviderFactoryForPayment) Get(channel entity.PaymentChannel) (service.PaymentProvider, error) {
	return m.provider, m.err
}

// mockProvider for payment tests
type mockProviderForPayment struct {
	result       *service.ProviderPaymentResult
	refundResult *service.ProviderRefundResult
	err          error
	verifyResult map[string]string
	verifyErr    error
}

func (m *mockProviderForPayment) Channel() entity.PaymentChannel { return entity.ChannelWechat }
func (m *mockProviderForPayment) CreatePayment(ctx context.Context, req service.ProviderPaymentRequest) (*service.ProviderPaymentResult, error) {
	return m.result, m.err
}
func (m *mockProviderForPayment) QueryPayment(ctx context.Context, thirdPartyNo string) (*service.ProviderPaymentResult, error) {
	return m.result, m.err
}
func (m *mockProviderForPayment) Refund(ctx context.Context, req service.ProviderRefundRequest) (*service.ProviderRefundResult, error) {
	return m.refundResult, m.err
}
func (m *mockProviderForPayment) QueryRefund(ctx context.Context, thirdPartyNo string) (*service.ProviderRefundResult, error) {
	return m.refundResult, m.err
}
func (m *mockProviderForPayment) VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (map[string]string, error) {
	return m.verifyResult, m.verifyErr
}

// mockTxMgr for payment tests
type mockTxMgrForPayment struct {
	err error
}

func (m *mockTxMgrForPayment) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

func TestPaymentUsecase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Subject: "test", Amount: 1000, Currency: "CNY", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		provider := &mockProviderForPayment{result: &service.ProviderPaymentResult{ThirdPartyNo: "TP001", PayParams: map[string]any{"qr_code": "https://qr.test"}}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		payment, params, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if payment.Amount != 1000 {
			t.Errorf("Amount = %d, want 1000", payment.Amount)
		}
		if params["qr_code"] != "https://qr.test" {
			t.Errorf("params qr_code = %v, want https://qr.test", params["qr_code"])
		}
	})

	t.Run("order not found", func(t *testing.T) {
		orderRepo := &mockOrderRepoForPayment{err: apperror.New(apperror.CodeOrderNotFound, "not found")}
		uc := NewPaymentUsecase(orderRepo, nil, nil, nil)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD999"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("order cannot pay", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPaid}
		orderRepo := &mockOrderRepoForPayment{order: order}
		uc := NewPaymentUsecase(orderRepo, nil, nil, nil)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !apperror.Is(err, apperror.CodeOrderCannotClose) {
			t.Errorf("error code mismatch: %v", err)
		}
	})

	t.Run("invalid channel", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		factory := &mockProviderFactoryForPayment{err: errors.New("invalid channel")}
		uc := NewPaymentUsecase(orderRepo, nil, factory, nil)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: "invalid"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("default expire minutes", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		provider := &mockProviderForPayment{result: &service.ProviderPaymentResult{}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		payment, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1", ExpireMinutes: 0})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if time.Until(payment.ExpireAt) < 25*time.Minute {
			t.Error("ExpireAt should be ~30 minutes from now")
		}
	})

	t.Run("expire minutes capped at 1440", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		provider := &mockProviderForPayment{result: &service.ProviderPaymentResult{}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		payment, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1", ExpireMinutes: 9999})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if time.Until(payment.ExpireAt) > 1441*time.Minute {
			t.Error("ExpireAt should be capped at 1440 minutes")
		}
	})

	t.Run("payment repo create error", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{err: errors.New("db error")}
		factory := &mockProviderFactoryForPayment{provider: &mockProviderForPayment{}}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("provider error updates status", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		provider := &mockProviderForPayment{err: errors.New("provider error")}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !paymentRepo.statusUpdated {
			t.Error("payment status should be updated to failed")
		}
	})

	t.Run("provider error with update status failure", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		paymentRepo.statusUpdated = false
		provider := &mockProviderForPayment{err: errors.New("provider error")}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		_, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("update after create failure", func(t *testing.T) {
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Amount: 100, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		orderRepo := &mockOrderRepoForPayment{order: order}
		paymentRepo := &mockPaymentRepoForPayment{}
		provider := &mockProviderForPayment{result: &service.ProviderPaymentResult{ThirdPartyNo: "TP001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		payment, _, err := uc.Create(context.Background(), CreatePaymentRequest{OrderNo: "ORD123", Channel: entity.ChannelWechat, Method: entity.MethodNative, ClientIP: "127.0.0.1"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if payment.ThirdPartyNo != "TP001" {
			t.Errorf("ThirdPartyNo = %v, want TP001", payment.ThirdPartyNo)
		}
	})
}

func TestPaymentUsecase_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{PaymentNo: "PAY123", Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		uc := NewPaymentUsecase(nil, paymentRepo, nil, nil)
		got, err := uc.Get(context.Background(), "PAY123")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.PaymentNo != "PAY123" {
			t.Errorf("PaymentNo = %v, want PAY123", got.PaymentNo)
		}
	})

	t.Run("not found", func(t *testing.T) {
		paymentRepo := &mockPaymentRepoForPayment{err: apperror.New(apperror.CodePaymentNotFound, "not found")}
		uc := NewPaymentUsecase(nil, paymentRepo, nil, nil)
		_, err := uc.Get(context.Background(), "PAY999")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPaymentUsecase_HandleWechatNotify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, Status: entity.PaymentStatusPending}
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		orderRepo := &mockOrderRepoForPayment{order: order}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err != nil {
			t.Fatalf("HandleWechatNotify() error = %v", err)
		}
		if !orderRepo.updated {
			t.Error("order should be updated to paid")
		}
	})

	t.Run("duplicate notify", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err != nil {
			t.Fatalf("HandleWechatNotify() error = %v", err)
		}
	})

	t.Run("non success trade state", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "NOTPAY", "transaction_id": ""}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err != nil {
			t.Fatalf("HandleWechatNotify() error = %v", err)
		}
	})

	t.Run("verify failed", func(t *testing.T) {
		provider := &mockProviderForPayment{verifyErr: errors.New("verify failed")}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, nil, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("provider factory error", func(t *testing.T) {
		factory := &mockProviderFactoryForPayment{err: errors.New("factory error")}
		uc := NewPaymentUsecase(nil, nil, factory, nil)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("payment not found", func(t *testing.T) {
		paymentRepo := &mockPaymentRepoForPayment{err: apperror.New(apperror.CodePaymentNotFound, "not found")}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("order not found after payment success", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		orderRepo := &mockOrderRepoForPayment{err: apperror.New(apperror.CodeOrderNotFound, "not found")}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("transaction error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{err: errors.New("tx error")}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleWechatNotify(context.Background(), []byte("{}"), map[string]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPaymentUsecase_HandleAlipayNotify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, Status: entity.PaymentStatusPending}
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		orderRepo := &mockOrderRepoForPayment{order: order}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_status": "TRADE_SUCCESS", "trade_no": "ALI001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err != nil {
			t.Fatalf("HandleAlipayNotify() error = %v", err)
		}
		if !orderRepo.updated {
			t.Error("order should be updated to paid")
		}
	})

	t.Run("trade finished", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, Status: entity.PaymentStatusPending}
		order := &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour)}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		orderRepo := &mockOrderRepoForPayment{order: order}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_status": "TRADE_FINISHED", "trade_no": "ALI001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(orderRepo, paymentRepo, factory, txMgr)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err != nil {
			t.Fatalf("HandleAlipayNotify() error = %v", err)
		}
	})

	t.Run("duplicate notify", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_status": "TRADE_SUCCESS", "trade_no": "ALI001"}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err != nil {
			t.Fatalf("HandleAlipayNotify() error = %v", err)
		}
	})

	t.Run("non success trade status", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepoForPayment{payment: payment}
		provider := &mockProviderForPayment{verifyResult: map[string]string{"out_trade_no": "PAY123", "trade_status": "WAIT_BUYER_PAY", "trade_no": ""}}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, paymentRepo, factory, txMgr)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err != nil {
			t.Fatalf("HandleAlipayNotify() error = %v", err)
		}
	})

	t.Run("verify failed", func(t *testing.T) {
		provider := &mockProviderForPayment{verifyErr: errors.New("verify failed")}
		factory := &mockProviderFactoryForPayment{provider: provider}
		txMgr := &mockTxMgrForPayment{}

		uc := NewPaymentUsecase(nil, nil, factory, txMgr)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("provider factory error", func(t *testing.T) {
		factory := &mockProviderFactoryForPayment{err: errors.New("factory error")}
		uc := NewPaymentUsecase(nil, nil, factory, nil)
		err := uc.HandleAlipayNotify(context.Background(), map[string]string{"out_trade_no": "PAY123"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
