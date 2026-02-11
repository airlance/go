package usecase

import (
	"context"
	"fmt"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
)

type GroupUseCase interface {
	CreateGroup(ctx context.Context, dto CreateGroupDTO) (*GroupDTO, error)
	GetGroupByID(ctx context.Context, id uint, withIPs bool) (*GroupDTO, error)
	GetGroupByGroupID(ctx context.Context, groupID int, withIPs bool) (*GroupDTO, error)
	ListGroups(ctx context.Context, pagination PaginationDTO, withIPs bool) ([]*GroupDTO, int64, error)
	UpdateCounters(ctx context.Context, groupID int) error
	DeleteGroup(ctx context.Context, groupID int) error
}

type groupUseCase struct {
	groupRepo     domain.GroupRepository
	ipRepo        domain.IPRepository
	historyRepo   domain.HistoryRepository
	scoreStatRepo domain.ScoreStatRepository
}

func NewGroupUseCase(
	groupRepo domain.GroupRepository,
	ipRepo domain.IPRepository,
	historyRepo domain.HistoryRepository,
	scoreStatRepo domain.ScoreStatRepository,
) GroupUseCase {
	return &groupUseCase{
		groupRepo:     groupRepo,
		ipRepo:        ipRepo,
		historyRepo:   historyRepo,
		scoreStatRepo: scoreStatRepo,
	}
}

func (uc *groupUseCase) CreateGroup(ctx context.Context, dto CreateGroupDTO) (*GroupDTO, error) {
	existing, err := uc.groupRepo.GetByGroupID(ctx, dto.GroupID)
	if err == nil && existing != nil {
		return nil, domain.ErrGroupAlreadyExists
	}
	if err != nil && err != domain.ErrGroupNotFound {
		return nil, fmt.Errorf("failed to check existing group: %w", err)
	}

	group := &domain.Group{
		GroupID:       dto.GroupID,
		GroupName:     dto.GroupName,
		SpamTrapCount: 0,
		IPsCount:      0,
	}

	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return uc.mapGroupToDTO(group, nil), nil
}

func (uc *groupUseCase) GetGroupByID(ctx context.Context, id uint, withIPs bool) (*GroupDTO, error) {
	group, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var ips []*domain.IP
	if withIPs {
		ips, err = uc.ipRepo.ListByGroupID(ctx, group.GroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to load IPs: %w", err)
		}
	}

	return uc.mapGroupToDTO(group, ips), nil
}

func (uc *groupUseCase) GetGroupByGroupID(ctx context.Context, groupID int, withIPs bool) (*GroupDTO, error) {
	group, err := uc.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var ips []*domain.IP
	if withIPs {
		ips, err = uc.ipRepo.ListByGroupID(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to load IPs: %w", err)
		}
	}

	return uc.mapGroupToDTO(group, ips), nil
}

func (uc *groupUseCase) ListGroups(ctx context.Context, pagination PaginationDTO, withIPs bool) ([]*GroupDTO, int64, error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 20
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	groups, total, err := uc.groupRepo.List(ctx, offset, pagination.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}

	result := make([]*GroupDTO, len(groups))
	for i, group := range groups {
		var ips []*domain.IP
		if withIPs {
			ips, err = uc.ipRepo.ListByGroupID(ctx, group.GroupID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to load IPs for group %d: %w", group.GroupID, err)
			}
		}
		result[i] = uc.mapGroupToDTO(group, ips)
	}

	return result, total, nil
}

func (uc *groupUseCase) UpdateCounters(ctx context.Context, groupID int) error {
	return uc.groupRepo.UpdateCounters(ctx, groupID)
}

func (uc *groupUseCase) DeleteGroup(ctx context.Context, groupID int) error {
	group, err := uc.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return err
	}

	ips, err := uc.ipRepo.ListByGroupID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group IPs: %w", err)
	}

	for _, ip := range ips {
		if err := uc.ipRepo.RemoveFromGroup(ctx, ip.ID, groupID); err != nil {
			return fmt.Errorf("failed to remove IP %d from group: %w", ip.ID, err)
		}

		hasOtherGroups, err := uc.ipRepo.IsIPInOtherGroups(ctx, ip.ID, groupID)
		if err != nil {
			return fmt.Errorf("failed to check if IP %d is in other groups: %w", ip.ID, err)
		}

		if !hasOtherGroups {
			if err := uc.historyRepo.DeleteByIPID(ctx, ip.ID); err != nil {
				return fmt.Errorf("failed to delete history for IP %d: %w", ip.ID, err)
			}

			if err := uc.scoreStatRepo.DeleteByIPID(ctx, ip.ID); err != nil {
				return fmt.Errorf("failed to delete stats for IP %d: %w", ip.ID, err)
			}

			if err := uc.ipRepo.Delete(ctx, ip.ID); err != nil {
				return fmt.Errorf("failed to delete IP %d: %w", ip.ID, err)
			}
		}
	}

	if err := uc.groupRepo.Delete(ctx, group.GroupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

func (uc *groupUseCase) mapGroupToDTO(group *domain.Group, ips []*domain.IP) *GroupDTO {
	dto := &GroupDTO{
		ID:            group.ID,
		GroupID:       group.GroupID,
		GroupName:     group.GroupName,
		SpamTrapCount: group.SpamTrapCount,
		IPsCount:      group.IPsCount,
		IPs:           make([]IPDTO, 0),
	}

	if ips != nil {
		dto.IPs = make([]IPDTO, len(ips))
		for i, ip := range ips {
			dto.IPs[i] = IPDTO{
				ID:         ip.ID,
				IP:         ip.IP,
				Score:      ip.Score,
				SpamTrap:   ip.SpamTrap,
				Blocklists: ip.Blocklists,
				Complaints: ip.Complaints,
				UpdatedAt:  ip.UpdatedAt.Unix(),
			}
		}
	}

	return dto
}
