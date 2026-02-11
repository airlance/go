package data

import (
	"context"
	"errors"
	"fmt"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
	"gorm.io/gorm"
)

type historyRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) domain.HistoryRepository {
	return &historyRepository{db: db}
}

func (r *historyRepository) Create(ctx context.Context, history *domain.History) error {
	model := toHistoryModel(history)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}
	history.ID = model.ID
	return nil
}

func (r *historyRepository) GetByIPAndDate(ctx context.Context, ipID uint, date string) (*domain.History, error) {
	var model HistoryModel
	if err := r.db.WithContext(ctx).Where("ips_id = ? AND time = ?", ipID, date).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get history by IP and date: %w", err)
	}
	return toHistoryDomain(&model), nil
}

func (r *historyRepository) ListByIPID(ctx context.Context, ipID uint) ([]*domain.History, error) {
	var models []HistoryModel
	if err := r.db.WithContext(ctx).Where("ips_id = ?", ipID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list history by IP: %w", err)
	}

	histories := make([]*domain.History, len(models))
	for i, model := range models {
		histories[i] = toHistoryDomain(&model)
	}

	return histories, nil
}

func (r *historyRepository) Update(ctx context.Context, history *domain.History) error {
	model := toHistoryModel(history)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update history: %w", err)
	}
	return nil
}

func (r *historyRepository) DeleteByIPID(ctx context.Context, ipID uint) error {
	if err := r.db.WithContext(ctx).Where("ips_id = ?", ipID).Delete(&HistoryModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete history by IP: %w", err)
	}
	return nil
}
