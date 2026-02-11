package data

import (
	"git.emercury.dev/emercury/senderscore/api/internal/domain"
)

func toGroupDomain(model *GroupModel) *domain.Group {
	if model == nil {
		return nil
	}
	return &domain.Group{
		ID:            model.ID,
		GroupID:       model.GroupID,
		GroupName:     model.GroupName,
		SpamTrapCount: model.SpamTrapCount,
		IPsCount:      model.IPsCount,
	}
}

func toGroupModel(entity *domain.Group) *GroupModel {
	if entity == nil {
		return nil
	}
	return &GroupModel{
		ID:            entity.ID,
		GroupID:       entity.GroupID,
		GroupName:     entity.GroupName,
		SpamTrapCount: entity.SpamTrapCount,
		IPsCount:      entity.IPsCount,
	}
}

func toIPDomain(model *IPModel) *domain.IP {
	if model == nil {
		return nil
	}
	return &domain.IP{
		ID:         model.ID,
		IP:         model.IP,
		Score:      model.Score,
		SpamTrap:   model.SpamTrap,
		Blocklists: model.Blocklists,
		Complaints: model.Complaints,
		UpdatedAt:  model.UpdatedAt,
		GroupIDs:   []int{}, // Будет заполнено в репозитории
	}
}

func toIPModel(entity *domain.IP) *IPModel {
	if entity == nil {
		return nil
	}
	return &IPModel{
		ID:         entity.ID,
		IP:         entity.IP,
		Score:      entity.Score,
		SpamTrap:   entity.SpamTrap,
		Blocklists: entity.Blocklists,
		Complaints: entity.Complaints,
		UpdatedAt:  entity.UpdatedAt,
	}
}

func toHistoryDomain(model *HistoryModel) *domain.History {
	if model == nil {
		return nil
	}
	return &domain.History{
		ID:       model.ID,
		IPsID:    model.IpsID,
		Score:    model.Score,
		SpamTrap: model.SpamTrap,
		Volume:   model.Volume,
		Time:     model.Time,
	}
}

func toHistoryModel(entity *domain.History) *HistoryModel {
	if entity == nil {
		return nil
	}
	return &HistoryModel{
		ID:       entity.ID,
		IpsID:    entity.IPsID,
		Score:    entity.Score,
		SpamTrap: entity.SpamTrap,
		Volume:   entity.Volume,
		Time:     entity.Time,
	}
}

func toScoreStatDomain(model *ScoreStatModel) *domain.ScoreStat {
	if model == nil {
		return nil
	}
	return &domain.ScoreStat{
		ID:     model.ID,
		IPsID:  model.IpsID,
		Score:  model.Score,
		Result: model.Result,
		Date:   model.Date,
	}
}

func toScoreStatModel(entity *domain.ScoreStat) *ScoreStatModel {
	if entity == nil {
		return nil
	}
	return &ScoreStatModel{
		ID:     entity.ID,
		IpsID:  entity.IPsID,
		Score:  entity.Score,
		Result: entity.Result,
		Date:   entity.Date,
	}
}
