package data

import (
	"context"
	"errors"
	"fmt"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) domain.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *domain.Group) error {
	model := toGroupModel(group)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	group.ID = model.ID
	return nil
}

func (r *groupRepository) GetByID(ctx context.Context, id uint) (*domain.Group, error) {
	var model GroupModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group by id: %w", err)
	}
	return toGroupDomain(&model), nil
}

func (r *groupRepository) GetByGroupID(ctx context.Context, groupID int) (*domain.Group, error) {
	var model GroupModel
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group by group_id: %w", err)
	}
	return toGroupDomain(&model), nil
}

func (r *groupRepository) List(ctx context.Context, offset, limit int) ([]*domain.Group, int64, error) {
	var models []GroupModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&GroupModel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}

	groups := make([]*domain.Group, len(models))
	for i, model := range models {
		groups[i] = toGroupDomain(&model)
	}

	return groups, total, nil
}

func (r *groupRepository) Update(ctx context.Context, group *domain.Group) error {
	model := toGroupModel(group)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	return nil
}

func (r *groupRepository) Delete(ctx context.Context, groupID int) error {
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Delete(&GroupModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	return nil
}

func (r *groupRepository) UpdateCounters(ctx context.Context, groupID int) error {
	var ipsCount int64
	var totalSpamTrap int64

	if err := r.db.WithContext(ctx).
		Table("sender_score_group_ips").
		Where("group_id = ?", groupID).
		Count(&ipsCount).Error; err != nil {
		return fmt.Errorf("failed to count IPs: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Table("sender_score_ips").
		Joins("JOIN sender_score_group_ips ON sender_score_ips.id = sender_score_group_ips.ip_id").
		Where("sender_score_group_ips.group_id = ?", groupID).
		Select("COALESCE(SUM(spam_trap), 0)").
		Scan(&totalSpamTrap).Error; err != nil {
		return fmt.Errorf("failed to sum spam traps: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&GroupModel{}).Where("group_id = ?", groupID).Updates(map[string]interface{}{
		"ips_count":       ipsCount,
		"spam_trap_count": totalSpamTrap,
	}).Error; err != nil {
		return fmt.Errorf("failed to update counters: %w", err)
	}

	return nil
}

func (r *groupRepository) GetGroupIDsByIP(ctx context.Context, ipID uint) ([]int, error) {
	var groupIDs []int
	if err := r.db.WithContext(ctx).
		Table("sender_score_group_ips").
		Where("ip_id = ?", ipID).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get group IDs by IP: %w", err)
	}
	return groupIDs, nil
}
