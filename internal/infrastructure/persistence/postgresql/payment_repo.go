package postgresql

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"gorm.io/gorm"
)

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) repository.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) getDB(ctx context.Context) *gorm.DB {
	return txFromContext(ctx, r.db)
}

func (r *paymentRepo) Create(ctx context.Context, payment *entity.Payment) error {
	m := toPaymentModel(payment)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	payment.ID = m.ID
	return nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id uint64) (*entity.Payment, error) {
	var m PaymentModel
	if err := r.getDB(ctx).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodePaymentNotFound, "payment not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toPaymentEntity(&m), nil
}

func (r *paymentRepo) GetByPaymentNo(ctx context.Context, paymentNo string) (*entity.Payment, error) {
	var m PaymentModel
	if err := r.getDB(ctx).Where("payment_no = ?", paymentNo).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodePaymentNotFound, "payment not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toPaymentEntity(&m), nil
}

func (r *paymentRepo) GetByThirdPartyNo(ctx context.Context, thirdPartyNo string) (*entity.Payment, error) {
	var m PaymentModel
	if err := r.getDB(ctx).Where("third_party_no = ?", thirdPartyNo).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodePaymentNotFound, "payment not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toPaymentEntity(&m), nil
}

func (r *paymentRepo) GetByOrderID(ctx context.Context, orderID uint64) ([]*entity.Payment, error) {
	var ms []PaymentModel
	if err := r.getDB(ctx).Where("order_id = ?", orderID).Find(&ms).Error; err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	payments := make([]*entity.Payment, len(ms))
	for i, m := range ms {
		payments[i] = toPaymentEntity(&m)
	}
	return payments, nil
}

func (r *paymentRepo) Update(ctx context.Context, payment *entity.Payment) error {
	m := toPaymentModel(payment)
	if err := r.getDB(ctx).Model(&PaymentModel{}).Where("id = ?", payment.ID).Select(
		"status", "third_party_no", "third_party_resp", "paid_at", "notify_at", "notify_count", "expire_at", "updated_at",
	).Updates(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id uint64, status entity.PaymentStatus) error {
	if err := r.getDB(ctx).Model(&PaymentModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":     int8(status),
		"updated_at": gorm.Expr("NOW()"),
	}).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *paymentRepo) List(ctx context.Context, filter repository.PaymentFilter) ([]*entity.Payment, int64, error) {
	var total int64
	query := r.getDB(ctx).Model(&PaymentModel{})
	if filter.OrderID > 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.Channel != "" {
		query = query.Where("channel = ?", string(filter.Channel))
	}
	if filter.Status >= 0 {
		query = query.Where("status = ?", int8(filter.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	var ms []PaymentModel
	offset := (filter.Page - 1) * filter.Size
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Size).Find(&ms).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	payments := make([]*entity.Payment, len(ms))
	for i, m := range ms {
		payments[i] = toPaymentEntity(&m)
	}
	return payments, total, nil
}

func toPaymentModel(e *entity.Payment) *PaymentModel {
	return &PaymentModel{
		ID:             e.ID,
		PaymentNo:      e.PaymentNo,
		OrderID:        e.OrderID,
		OrderNo:        e.OrderNo,
		Channel:        string(e.Channel),
		Method:         string(e.Method),
		Amount:         e.Amount,
		Currency:       e.Currency,
		Status:         int8(e.Status),
		ThirdPartyNo:   e.ThirdPartyNo,
		ThirdPartyResp: JSONMap(e.ThirdPartyResp),
		ClientIP:       e.ClientIP,
		NotifyURL:      e.NotifyURL,
		ReturnURL:      e.ReturnURL,
		ExpireAt:       e.ExpireAt,
		PaidAt:         e.PaidAt,
		NotifyAt:       e.NotifyAt,
		NotifyCount:    e.NotifyCount,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func toPaymentEntity(m *PaymentModel) *entity.Payment {
	return &entity.Payment{
		ID:             m.ID,
		PaymentNo:      m.PaymentNo,
		OrderID:        m.OrderID,
		OrderNo:        m.OrderNo,
		Channel:        entity.PaymentChannel(m.Channel),
		Method:         entity.PaymentMethod(m.Method),
		Amount:         m.Amount,
		Currency:       m.Currency,
		Status:         entity.PaymentStatus(m.Status),
		ThirdPartyNo:   m.ThirdPartyNo,
		ThirdPartyResp: map[string]any(m.ThirdPartyResp),
		ClientIP:       m.ClientIP,
		NotifyURL:      m.NotifyURL,
		ReturnURL:      m.ReturnURL,
		ExpireAt:       m.ExpireAt,
		PaidAt:         m.PaidAt,
		NotifyAt:       m.NotifyAt,
		NotifyCount:    m.NotifyCount,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
