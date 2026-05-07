package postgresql

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"gorm.io/gorm"
)

type refundRepo struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) repository.RefundRepository {
	return &refundRepo{db: db}
}

func (r *refundRepo) getDB(ctx context.Context) *gorm.DB {
	return txFromContext(ctx, r.db)
}

func (r *refundRepo) Create(ctx context.Context, refund *entity.Refund) error {
	m := toRefundModel(refund)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	refund.ID = m.ID
	return nil
}

func (r *refundRepo) GetByID(ctx context.Context, id uint64) (*entity.Refund, error) {
	var m RefundModel
	if err := r.getDB(ctx).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeRefundNotFound, "refund not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toRefundEntity(&m), nil
}

func (r *refundRepo) GetByRefundNo(ctx context.Context, refundNo string) (*entity.Refund, error) {
	var m RefundModel
	if err := r.getDB(ctx).Where("refund_no = ?", refundNo).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeRefundNotFound, "refund not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toRefundEntity(&m), nil
}

func (r *refundRepo) GetByPaymentID(ctx context.Context, paymentID uint64) ([]*entity.Refund, error) {
	var ms []RefundModel
	if err := r.getDB(ctx).Where("payment_id = ?", paymentID).Find(&ms).Error; err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	refunds := make([]*entity.Refund, len(ms))
	for i, m := range ms {
		refunds[i] = toRefundEntity(&m)
	}
	return refunds, nil
}

func (r *refundRepo) Update(ctx context.Context, refund *entity.Refund) error {
	m := toRefundModel(refund)
	if err := r.getDB(ctx).Model(&RefundModel{}).Where("id = ?", refund.ID).Select(
		"status", "third_party_no", "third_party_resp", "notify_at", "notify_count", "updated_at",
	).Updates(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *refundRepo) UpdateStatus(ctx context.Context, id uint64, status entity.RefundStatus) error {
	if err := r.getDB(ctx).Model(&RefundModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":     int8(status),
		"updated_at": gorm.Expr("NOW()"),
	}).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *refundRepo) GetTotalRefundAmount(ctx context.Context, paymentID uint64) (int64, error) {
	var total int64
	if err := r.getDB(ctx).Model(&RefundModel{}).
		Where("payment_id = ? AND status = ?", paymentID, int8(entity.RefundStatusSuccess)).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return total, nil
}

func (r *refundRepo) List(ctx context.Context, filter repository.RefundFilter) ([]*entity.Refund, int64, error) {
	var total int64
	query := r.getDB(ctx).Model(&RefundModel{})
	if filter.PaymentID > 0 {
		query = query.Where("payment_id = ?", filter.PaymentID)
	}
	if filter.OrderID > 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.Status >= 0 {
		query = query.Where("status = ?", int8(filter.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	var ms []RefundModel
	offset := (filter.Page - 1) * filter.Size
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Size).Find(&ms).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	refunds := make([]*entity.Refund, len(ms))
	for i, m := range ms {
		refunds[i] = toRefundEntity(&m)
	}
	return refunds, total, nil
}

func toRefundModel(e *entity.Refund) *RefundModel {
	return &RefundModel{
		ID:             e.ID,
		RefundNo:       e.RefundNo,
		PaymentID:      e.PaymentID,
		PaymentNo:      e.PaymentNo,
		OrderID:        e.OrderID,
		OrderNo:        e.OrderNo,
		Channel:        string(e.Channel),
		Amount:         e.Amount,
		Reason:         e.Reason,
		Status:         int8(e.Status),
		ThirdPartyNo:   e.ThirdPartyNo,
		ThirdPartyResp: JSONMap(e.ThirdPartyResp),
		NotifyAt:       e.NotifyAt,
		NotifyCount:    e.NotifyCount,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func toRefundEntity(m *RefundModel) *entity.Refund {
	return &entity.Refund{
		ID:             m.ID,
		RefundNo:       m.RefundNo,
		PaymentID:      m.PaymentID,
		PaymentNo:      m.PaymentNo,
		OrderID:        m.OrderID,
		OrderNo:        m.OrderNo,
		Channel:        entity.PaymentChannel(m.Channel),
		Amount:         m.Amount,
		Reason:         m.Reason,
		Status:         entity.RefundStatus(m.Status),
		ThirdPartyNo:   m.ThirdPartyNo,
		ThirdPartyResp: map[string]any(m.ThirdPartyResp),
		NotifyAt:       m.NotifyAt,
		NotifyCount:    m.NotifyCount,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
