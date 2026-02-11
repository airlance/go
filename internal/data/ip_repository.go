package data

import (
	"context"
	"errors"
	"fmt"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
	"gorm.io/gorm"
)

type ipRepository struct {
	db *gorm.DB
}

func NewIPRepository(db *gorm.DB) domain.IPRepository {
	return &ipRepository{db: db}
}

func (r *ipRepository) Create(ctx context.Context, ip *domain.IP) error {
	model := toIPModel(ip)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create IP: %w", err)
	}
	ip.ID = model.ID
	return nil
}

func (r *ipRepository) GetByID(ctx context.Context, id uint) (*domain.IP, error) {
	var model IPModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIPNotFound
		}
		return nil, fmt.Errorf("failed to get IP by id: %w", err)
	}

	groupIDs, err := r.getGroupIDsForIP(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load group IDs: %w", err)
	}

	ip := toIPDomain(&model)
	ip.GroupIDs = groupIDs
	return ip, nil
}

func (r *ipRepository) GetByIP(ctx context.Context, ipAddress string) (*domain.IP, error) {
	var model IPModel
	if err := r.db.WithContext(ctx).Where("ip = ?", ipAddress).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIPNotFound
		}
		return nil, fmt.Errorf("failed to get IP by address: %w", err)
	}

	groupIDs, err := r.getGroupIDsForIP(ctx, model.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load group IDs: %w", err)
	}

	ip := toIPDomain(&model)
	ip.GroupIDs = groupIDs
	return ip, nil
}

func (r *ipRepository) GetOldestIP(ctx context.Context) (*domain.IP, error) {
	var model IPModel
	if err := r.db.WithContext(ctx).
		Order("updated_at ASC").
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIPNotFound
		}
		return nil, fmt.Errorf("failed to get oldest IP: %w", err)
	}

	groupIDs, err := r.getGroupIDsForIP(ctx, model.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load group IDs: %w", err)
	}

	ip := toIPDomain(&model)
	ip.GroupIDs = groupIDs
	return ip, nil
}

func (r *ipRepository) ListByGroupID(ctx context.Context, groupID int) ([]*domain.IP, error) {
	var models []IPModel
	if err := r.db.WithContext(ctx).
		Joins("JOIN sender_score_group_ips ON sender_score_ips.id = sender_score_group_ips.ip_id").
		Where("sender_score_group_ips.group_id = ?", groupID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list IPs by group: %w", err)
	}

	ips := make([]*domain.IP, len(models))
	for i, model := range models {
		groupIDs, err := r.getGroupIDsForIP(ctx, model.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load group IDs for IP %d: %w", model.ID, err)
		}
		ip := toIPDomain(&model)
		ip.GroupIDs = groupIDs
		ips[i] = ip
	}

	return ips, nil
}

func (r *ipRepository) Update(ctx context.Context, ip *domain.IP) error {
	model := toIPModel(ip)
	if err := r.db.WithContext(ctx).Model(&IPModel{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
		"score":      model.Score,
		"spam_trap":  model.SpamTrap,
		"blocklists": model.Blocklists,
		"complaints": model.Complaints,
		"updated_at": model.UpdatedAt,
	}).Error; err != nil {
		return fmt.Errorf("failed to update IP: %w", err)
	}
	return nil
}

func (r *ipRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&IPModel{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete IP: %w", err)
	}
	return nil
}

func (r *ipRepository) AddToGroup(ctx context.Context, ipID uint, groupID int) error {
	groupIP := GroupIPModel{
		GroupID: groupID,
		IPID:    ipID,
	}

	if err := r.db.WithContext(ctx).Create(&groupIP).Error; err != nil {
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("failed to add IP to group: %w", err)
		}
	}
	return nil
}

func (r *ipRepository) RemoveFromGroup(ctx context.Context, ipID uint, groupID int) error {
	if err := r.db.WithContext(ctx).
		Where("ip_id = ? AND group_id = ?", ipID, groupID).
		Delete(&GroupIPModel{}).Error; err != nil {
		return fmt.Errorf("failed to remove IP from group: %w", err)
	}
	return nil
}

func (r *ipRepository) IsIPInOtherGroups(ctx context.Context, ipID uint, excludeGroupID int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("sender_score_group_ips").
		Where("ip_id = ? AND group_id != ?", ipID, excludeGroupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check IP in other groups: %w", err)
	}
	return count > 0, nil
}

func (r *ipRepository) getGroupIDsForIP(ctx context.Context, ipID uint) ([]int, error) {
	var groupIDs []int
	if err := r.db.WithContext(ctx).
		Table("sender_score_group_ips").
		Where("ip_id = ?", ipID).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return nil, err
	}
	return groupIDs, nil
}
