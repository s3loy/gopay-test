package postgresql

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"gorm.io/gorm"
)

type orderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) getDB(ctx context.Context) *gorm.DB {
	return txFromContext(ctx, r.db)
}

func (r *orderRepo) Create(ctx context.Context, order *entity.Order) error {
	m := toOrderModel(order)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	order.ID = m.ID
	return nil
}

func (r *orderRepo) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	var m OrderModel
	if err := r.getDB(ctx).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeOrderNotFound, "order not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toOrderEntity(&m), nil
}

func (r *orderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	var m OrderModel
	if err := r.getDB(ctx).Where("order_no = ?", orderNo).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.CodeOrderNotFound, "order not found")
		}
		return nil, apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return toOrderEntity(&m), nil
}

func (r *orderRepo) Update(ctx context.Context, order *entity.Order) error {
	m := toOrderModel(order)
	if err := r.getDB(ctx).Model(&OrderModel{}).Where("id = ?", order.ID).Select(
		"status", "expired_at", "paid_at", "description", "metadata", "updated_at",
	).Updates(m).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uint64, status entity.OrderStatus) error {
	if err := r.getDB(ctx).Model(&OrderModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":     int8(status),
		"updated_at": gorm.Expr("NOW()"),
	}).Error; err != nil {
		return apperror.Wrap(err, apperror.CodeDatabaseError)
	}
	return nil
}

func (r *orderRepo) List(ctx context.Context, filter repository.OrderFilter) ([]*entity.Order, int64, error) {
	var total int64
	query := r.getDB(ctx).Model(&OrderModel{})
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Status >= 0 {
		query = query.Where("status = ?", int8(filter.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	var ms []OrderModel
	offset := (filter.Page - 1) * filter.Size
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Size).Find(&ms).Error; err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDatabaseError)
	}

	orders := make([]*entity.Order, len(ms))
	for i, m := range ms {
		orders[i] = toOrderEntity(&m)
	}
	return orders, total, nil
}

func toOrderModel(e *entity.Order) *OrderModel {
	return &OrderModel{
		ID:          e.ID,
		OrderNo:     e.OrderNo,
		UserID:      e.UserID,
		Subject:     e.Subject,
		Amount:      e.Amount,
		Currency:    e.Currency,
		Status:      int8(e.Status),
		ExpiredAt:   e.ExpiredAt,
		PaidAt:      e.PaidAt,
		Description: e.Description,
		Metadata:    JSONMap(e.Metadata),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func toOrderEntity(m *OrderModel) *entity.Order {
	return &entity.Order{
		ID:          m.ID,
		OrderNo:     m.OrderNo,
		UserID:      m.UserID,
		Subject:     m.Subject,
		Amount:      m.Amount,
		Currency:    m.Currency,
		Status:      entity.OrderStatus(m.Status),
		ExpiredAt:   m.ExpiredAt,
		PaidAt:      m.PaidAt,
		Description: m.Description,
		Metadata:    map[string]any(m.Metadata),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
