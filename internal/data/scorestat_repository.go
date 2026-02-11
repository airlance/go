package data

import (
	"context"
	"fmt"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
	"gorm.io/gorm"
)

type scoreStatRepository struct {
	db *gorm.DB
}

func NewScoreStatRepository(db *gorm.DB) domain.ScoreStatRepository {
	return &scoreStatRepository{db: db}
}

func (r *scoreStatRepository) Create(ctx context.Context, stat *domain.ScoreStat) error {
	model := toScoreStatModel(stat)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create score stat: %w", err)
	}
	stat.ID = model.ID
	return nil
}

func (r *scoreStatRepository) ListByIPID(ctx context.Context, ipID uint) ([]*domain.ScoreStat, error) {
	var models []ScoreStatModel
	if err := r.db.WithContext(ctx).Where("ips_id = ?", ipID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list score stats by IP: %w", err)
	}

	stats := make([]*domain.ScoreStat, len(models))
	for i, model := range models {
		stats[i] = toScoreStatDomain(&model)
	}

	return stats, nil
}

func (r *scoreStatRepository) DeleteByIPID(ctx context.Context, ipID uint) error {
	if err := r.db.WithContext(ctx).Where("ips_id = ?", ipID).Delete(&ScoreStatModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete score stats by IP: %w", err)
	}
	return nil
}
