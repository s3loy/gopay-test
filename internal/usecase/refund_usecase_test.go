package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

// mockPaymentRepo for testing
type mockPaymentRepo struct {
	payment       *entity.Payment
	payments      []*entity.Payment
	err           error
	created       *entity.Payment
	updated       bool
	statusUpdated bool
}

func (m *mockPaymentRepo) Create(ctx context.Context, p *entity.Payment) error {
	m.created = p
	if m.err != nil {
		return m.err
	}
	p.ID = 1
	return nil
}
func (m *mockPaymentRepo) GetByID(ctx context.Context, id uint64) (*entity.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.payment, nil
}
func (m *mockPaymentRepo) GetByPaymentNo(ctx context.Context, no string) (*entity.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.payment, nil
}
func (m *mockPaymentRepo) GetByThirdPartyNo(ctx context.Context, no string) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentRepo) GetByOrderID(ctx context.Context, id uint64) ([]*entity.Payment, error) {
	return m.payments, m.err
}
func (m *mockPaymentRepo) Update(ctx context.Context, p *entity.Payment) error {
	m.updated = true
	return m.err
}
func (m *mockPaymentRepo) UpdateStatus(ctx context.Context, id uint64, s entity.PaymentStatus) error {
	m.statusUpdated = true
	return m.err
}
func (m *mockPaymentRepo) List(ctx context.Context, f repository.PaymentFilter) ([]*entity.Payment, int64, error) {
	return nil, 0, m.err
}

// mockRefundRepo for testing
type mockRefundRepo struct {
	refund        *entity.Refund
	err           error
	created       *entity.Refund
	updated       bool
	statusUpdated bool
	totalAmount   int64
}

func (m *mockRefundRepo) Create(ctx context.Context, r *entity.Refund) error {
	m.created = r
	if m.err != nil {
		return m.err
	}
	r.ID = 1
	m.totalAmount += r.Amount
	return nil
}
func (m *mockRefundRepo) GetByID(ctx context.Context, id uint64) (*entity.Refund, error) {
	return m.refund, m.err
}
func (m *mockRefundRepo) GetByRefundNo(ctx context.Context, no string) (*entity.Refund, error) {
	return m.refund, m.err
}
func (m *mockRefundRepo) GetByPaymentID(ctx context.Context, id uint64) ([]*entity.Refund, error) {
	return nil, m.err
}
func (m *mockRefundRepo) Update(ctx context.Context, r *entity.Refund) error {
	m.updated = true
	return m.err
}
func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id uint64, s entity.RefundStatus) error {
	m.statusUpdated = true
	return m.err
}
func (m *mockRefundRepo) GetTotalRefundAmount(ctx context.Context, paymentID uint64) (int64, error) {
	return m.totalAmount, m.err
}
func (m *mockRefundRepo) List(ctx context.Context, f repository.RefundFilter) ([]*entity.Refund, int64, error) {
	return nil, 0, m.err
}

// mockProvider for testing
type mockProvider struct {
	result       *service.ProviderPaymentResult
	refundResult *service.ProviderRefundResult
	err          error
}

func (m *mockProvider) Channel() entity.PaymentChannel { return entity.ChannelWechat }
func (m *mockProvider) CreatePayment(ctx context.Context, req service.ProviderPaymentRequest) (*service.ProviderPaymentResult, error) {
	return m.result, m.err
}
func (m *mockProvider) QueryPayment(ctx context.Context, thirdPartyNo string) (*service.ProviderPaymentResult, error) {
	return m.result, m.err
}
func (m *mockProvider) Refund(ctx context.Context, req service.ProviderRefundRequest) (*service.ProviderRefundResult, error) {
	return m.refundResult, m.err
}
func (m *mockProvider) QueryRefund(ctx context.Context, thirdPartyNo string) (*service.ProviderRefundResult, error) {
	return m.refundResult, m.err
}
func (m *mockProvider) VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (map[string]string, error) {
	return map[string]string{"out_trade_no": "PAY123", "trade_state": "SUCCESS", "transaction_id": "TX123"}, nil
}

// mockProviderFactory for testing
type mockProviderFactory struct {
	provider service.PaymentProvider
	err      error
}

func (m *mockProviderFactory) Get(channel entity.PaymentChannel) (service.PaymentProvider, error) {
	return m.provider, m.err
}

// mockTxMgr for testing
type mockTxMgr struct{}

func (m *mockTxMgr) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestRefundUsecase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", OrderID: 1, OrderNo: "ORD123", Channel: entity.ChannelWechat, Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		provider := &mockProvider{refundResult: &service.ProviderRefundResult{ThirdPartyNo: "RF001", Status: "SUCCESS"}}
		factory := &mockProviderFactory{provider: provider}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		refund, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if refund.Amount != 500 {
			t.Errorf("Amount = %d, want 500", refund.Amount)
		}
		if refund.Status != entity.RefundStatusProcessing {
			t.Errorf("Status = %v, want processing", refund.Status)
		}
	})

	t.Run("payment not found", func(t *testing.T) {
		paymentRepo := &mockPaymentRepo{err: apperror.New(apperror.CodePaymentNotFound, "not found")}
		refundRepo := &mockRefundRepo{}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY999", Amount: 100, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("payment not refundable", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusPending}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 100, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !apperror.Is(err, apperror.CodePaymentNotRefundable) {
			t.Errorf("error code mismatch: %v", err)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 0, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("amount exceeds payment", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 2000, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !apperror.Is(err, apperror.CodeRefundAmountExceedsPayment) {
			t.Errorf("error code mismatch: %v", err)
		}
	})

	t.Run("amount exceeds remaining", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 600}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !apperror.Is(err, apperror.CodeRefundAmountExceedsRemaining) {
			t.Errorf("error code mismatch: %v", err)
		}
	})

	t.Run("provider error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		provider := &mockProvider{err: errors.New("provider error")}
		factory := &mockProviderFactory{provider: provider}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !refundRepo.statusUpdated {
			t.Error("refund status should be updated to failed")
		}
	})

	t.Run("full refund updates payment to refunded", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		provider := &mockProvider{refundResult: &service.ProviderRefundResult{ThirdPartyNo: "RF001", Status: "SUCCESS"}}
		factory := &mockProviderFactory{provider: provider}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 1000, Reason: "full refund"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if !paymentRepo.statusUpdated {
			t.Error("payment status should be updated to refunded")
		}
	})

	t.Run("get total refund amount error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{err: errors.New("db error")}
		factory := &mockProviderFactory{}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("provider factory error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		factory := &mockProviderFactory{err: errors.New("factory error")}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("refund repo create error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		refundRepo.err = errors.New("db error")
		factory := &mockProviderFactory{provider: &mockProvider{}}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		_, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("refund update after provider success error", func(t *testing.T) {
		payment := &entity.Payment{ID: 1, PaymentNo: "PAY123", Amount: 1000, Status: entity.PaymentStatusSuccess}
		paymentRepo := &mockPaymentRepo{payment: payment}
		refundRepo := &mockRefundRepo{totalAmount: 0}
		provider := &mockProvider{refundResult: &service.ProviderRefundResult{ThirdPartyNo: "RF001", Status: "SUCCESS"}}
		factory := &mockProviderFactory{provider: provider}
		txMgr := &mockTxMgr{}

		uc := NewRefundUsecase(paymentRepo, refundRepo, factory, txMgr)
		refund, err := uc.Create(context.Background(), CreateRefundRequest{PaymentNo: "PAY123", Amount: 500, Reason: "test"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if refund.Status != entity.RefundStatusProcessing {
			t.Errorf("Status = %v, want processing", refund.Status)
		}
	})
}

func TestRefundUsecase_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		refund := &entity.Refund{RefundNo: "REF123", Status: entity.RefundStatusSuccess}
		refundRepo := &mockRefundRepo{refund: refund}
		uc := NewRefundUsecase(nil, refundRepo, nil, nil)
		got, err := uc.Get(context.Background(), "REF123")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.RefundNo != "REF123" {
			t.Errorf("RefundNo = %s, want REF123", got.RefundNo)
		}
	})

	t.Run("not found", func(t *testing.T) {
		refundRepo := &mockRefundRepo{err: apperror.New(apperror.CodeRefundNotFound, "not found")}
		uc := NewRefundUsecase(nil, refundRepo, nil, nil)
		_, err := uc.Get(context.Background(), "REF999")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
