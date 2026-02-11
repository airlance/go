package usecase

import (
	"context"
	"fmt"
	"time"

	"git.emercury.dev/emercury/senderscore/api/internal/domain"
)

type IPUseCase interface {
	AddIP(ctx context.Context, dto AddIPDTO) (*IPDTO, error)
	AddIPs(ctx context.Context, dtos []AddIPDTO) (*BatchIPResultDTO, error)
	SubmitScore(ctx context.Context, dto SubmitScoreDTO) (*SubmitScoreResultDTO, error)
	GetOldestIP(ctx context.Context) (*IPDTO, error)
}

type ipUseCase struct {
	groupRepo   domain.GroupRepository
	ipRepo      domain.IPRepository
	historyRepo domain.HistoryRepository
}

func NewIPUseCase(
	groupRepo domain.GroupRepository,
	ipRepo domain.IPRepository,
	historyRepo domain.HistoryRepository,
) IPUseCase {
	return &ipUseCase{
		groupRepo:   groupRepo,
		ipRepo:      ipRepo,
		historyRepo: historyRepo,
	}
}

func (uc *ipUseCase) AddIP(ctx context.Context, dto AddIPDTO) (*IPDTO, error) {
	existing, err := uc.ipRepo.GetByIP(ctx, dto.IP)
	if err == nil && existing != nil {
		if err := uc.ipRepo.AddToGroup(ctx, existing.ID, dto.GroupID); err != nil {
			return nil, fmt.Errorf("failed to add IP to group: %w", err)
		}

		if err := uc.groupRepo.UpdateCounters(ctx, dto.GroupID); err != nil {
			return nil, fmt.Errorf("failed to update group counters: %w", err)
		}

		return uc.mapIPToDTO(existing), nil
	}
	if err != nil && err != domain.ErrIPNotFound {
		return nil, fmt.Errorf("failed to check existing IP: %w", err)
	}

	if err := uc.ensureGroupExists(ctx, dto.GroupID, dto.GroupName); err != nil {
		return nil, err
	}

	ip := &domain.IP{
		IP:         dto.IP,
		Score:      0,
		SpamTrap:   0,
		Blocklists: "",
		Complaints: "",
		UpdatedAt:  time.Now(),
	}

	if err := uc.ipRepo.Create(ctx, ip); err != nil {
		return nil, fmt.Errorf("failed to create IP: %w", err)
	}

	if err := uc.ipRepo.AddToGroup(ctx, ip.ID, dto.GroupID); err != nil {
		return nil, fmt.Errorf("failed to add IP to group: %w", err)
	}

	if err := uc.groupRepo.UpdateCounters(ctx, dto.GroupID); err != nil {
		return nil, fmt.Errorf("failed to update group counters: %w", err)
	}

	return uc.mapIPToDTO(ip), nil
}

func (uc *ipUseCase) AddIPs(ctx context.Context, dtos []AddIPDTO) (*BatchIPResultDTO, error) {
	result := &BatchIPResultDTO{}
	groupCache := make(map[int]bool)

	for _, dto := range dtos {
		existing, err := uc.ipRepo.GetByIP(ctx, dto.IP)
		if err == nil && existing != nil {
			if err := uc.ipRepo.AddToGroup(ctx, existing.ID, dto.GroupID); err != nil {
				return nil, fmt.Errorf("failed to add IP %s to group: %w", dto.IP, err)
			}
			result.IPsSkipped++
		} else if err == domain.ErrIPNotFound {
			if !groupCache[dto.GroupID] {
				if err := uc.ensureGroupExists(ctx, dto.GroupID, dto.GroupName); err != nil {
					if err != domain.ErrGroupAlreadyExists {
						return nil, err
					}
				} else {
					result.GroupsCreated++
				}
				groupCache[dto.GroupID] = true
			}

			ip := &domain.IP{
				IP:         dto.IP,
				Score:      0,
				SpamTrap:   0,
				Blocklists: "",
				Complaints: "",
				UpdatedAt:  time.Now(),
			}

			if err := uc.ipRepo.Create(ctx, ip); err != nil {
				return nil, fmt.Errorf("failed to create IP %s: %w", dto.IP, err)
			}

			if err := uc.ipRepo.AddToGroup(ctx, ip.ID, dto.GroupID); err != nil {
				return nil, fmt.Errorf("failed to add IP %s to group: %w", dto.IP, err)
			}

			result.IPsCreated++
		} else {
			return nil, fmt.Errorf("failed to check IP %s: %w", dto.IP, err)
		}

		if err := uc.groupRepo.UpdateCounters(ctx, dto.GroupID); err != nil {
			return nil, fmt.Errorf("failed to update counters for group %d: %w", dto.GroupID, err)
		}
	}

	result.Message = fmt.Sprintf(
		"groups created: %d; ips created: %d; ips skipped: %d",
		result.GroupsCreated,
		result.IPsCreated,
		result.IPsSkipped,
	)

	return result, nil
}

func (uc *ipUseCase) SubmitScore(ctx context.Context, dto SubmitScoreDTO) (*SubmitScoreResultDTO, error) {
	result := &SubmitScoreResultDTO{Success: true}

	ip, err := uc.ipRepo.GetByIP(ctx, dto.IP)
	if err == domain.ErrIPNotFound {
		ip = &domain.IP{
			IP:         dto.IP,
			Score:      dto.Score,
			SpamTrap:   dto.SpamTrap,
			Blocklists: dto.Blocklists,
			Complaints: dto.Complaints,
			UpdatedAt:  time.Now(),
		}
		if err := uc.ipRepo.Create(ctx, ip); err != nil {
			return nil, fmt.Errorf("failed to create IP: %w", err)
		}
		result.IPCreated = true
	} else if err != nil {
		return nil, fmt.Errorf("failed to check IP: %w", err)
	} else {
		ip.Score = dto.Score
		ip.SpamTrap = dto.SpamTrap
		ip.Blocklists = dto.Blocklists
		ip.Complaints = dto.Complaints
		ip.UpdatedAt = time.Now()

		if err := uc.ipRepo.Update(ctx, ip); err != nil {
			return nil, fmt.Errorf("failed to update IP: %w", err)
		}
	}

	groupIDs, err := uc.groupRepo.GetGroupIDsByIP(ctx, ip.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group IDs: %w", err)
	}

	for _, groupID := range groupIDs {
		if err := uc.groupRepo.UpdateCounters(ctx, groupID); err != nil {
			return nil, fmt.Errorf("failed to update counters for group %d: %w", groupID, err)
		}
	}

	for _, histEntry := range dto.History {
		date, err := time.Parse("02.01.2006", histEntry.Date)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", domain.ErrInvalidDateFormat, histEntry.Date)
		}

		existing, err := uc.historyRepo.GetByIPAndDate(ctx, ip.ID, histEntry.Date)
		if err == domain.ErrGroupNotFound {
			history := &domain.History{
				IPsID:    ip.ID,
				Score:    histEntry.Score,
				SpamTrap: histEntry.SpamTrap,
				Volume:   histEntry.Volume,
				Time:     date,
			}
			if err := uc.historyRepo.Create(ctx, history); err != nil {
				return nil, fmt.Errorf("failed to create history: %w", err)
			}
			result.HistoryAdded++
		} else if err != nil {
			return nil, fmt.Errorf("failed to check history: %w", err)
		} else {
			existing.Score = histEntry.Score
			existing.SpamTrap = histEntry.SpamTrap
			existing.Volume = histEntry.Volume

			if err := uc.historyRepo.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to update history: %w", err)
			}
			result.HistoryUpdated++
		}
	}

	result.Message = fmt.Sprintf(
		"Successfully processed. IP created: %t, History added: %d, History updated: %d",
		result.IPCreated,
		result.HistoryAdded,
		result.HistoryUpdated,
	)

	return result, nil
}

func (uc *ipUseCase) GetOldestIP(ctx context.Context) (*IPDTO, error) {
	ip, err := uc.ipRepo.GetOldestIP(ctx)
	if err != nil {
		return nil, err
	}
	return uc.mapIPToDTO(ip), nil
}

func (uc *ipUseCase) ensureGroupExists(ctx context.Context, groupID int, groupName string) error {
	_, err := uc.groupRepo.GetByGroupID(ctx, groupID)
	if err == nil {
		return domain.ErrGroupAlreadyExists
	}
	if err != domain.ErrGroupNotFound {
		return fmt.Errorf("failed to check group: %w", err)
	}

	group := &domain.Group{
		GroupID:       groupID,
		GroupName:     groupName,
		SpamTrapCount: 0,
		IPsCount:      0,
	}

	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

func (uc *ipUseCase) mapIPToDTO(ip *domain.IP) *IPDTO {
	return &IPDTO{
		ID:         ip.ID,
		IP:         ip.IP,
		Score:      ip.Score,
		SpamTrap:   ip.SpamTrap,
		Blocklists: ip.Blocklists,
		Complaints: ip.Complaints,
		UpdatedAt:  ip.UpdatedAt.Unix(),
	}
}
