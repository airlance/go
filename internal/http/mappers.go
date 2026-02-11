package http

import (
	"git.emercury.dev/emercury/senderscore/api/internal/usecase"
)

func toCreateGroupDTO(req CreateGroupRequest) usecase.CreateGroupDTO {
	return usecase.CreateGroupDTO{
		GroupID:   req.GroupID,
		GroupName: req.GroupName,
	}
}

func toAddIPDTO(req AddIPRequest) usecase.AddIPDTO {
	return usecase.AddIPDTO{
		GroupID:    req.GroupID,
		GroupName:  req.GroupName,
		IP:         req.IP,
		Score:      req.Score,
		SpamTrap:   req.SpamTrap,
		Blocklists: req.Blocklists,
		Complaints: req.Complaints,
	}
}

func toAddIPDTOs(req AddIPsRequest) []usecase.AddIPDTO {
	dtos := make([]usecase.AddIPDTO, len(req.IPs))
	for i, ip := range req.IPs {
		dtos[i] = toAddIPDTO(ip)
	}
	return dtos
}

func toHistoryEntryDTOs(entries []HistoryEntry) []usecase.HistoryEntryDTO {
	dtos := make([]usecase.HistoryEntryDTO, len(entries))
	for i, entry := range entries {
		dtos[i] = usecase.HistoryEntryDTO{
			Date:     entry.Date,
			Score:    entry.Score,
			Volume:   entry.Volume,
			SpamTrap: entry.SpamTrap,
		}
	}
	return dtos
}

func toSubmitScoreDTO(req SubmitScoreRequest) usecase.SubmitScoreDTO {
	return usecase.SubmitScoreDTO{
		IP:         req.IP,
		Score:      req.Score,
		SpamTrap:   req.SpamTrap,
		Blocklists: req.Blocklists,
		Complaints: req.Complaints,
		History:    toHistoryEntryDTOs(req.History),
	}
}

func toGroupResponse(dto *usecase.GroupDTO) GroupResponse {
	ips := make([]IPResponse, len(dto.IPs))
	for i, ip := range dto.IPs {
		ips[i] = IPResponse{
			ID:         ip.ID,
			IP:         ip.IP,
			Score:      ip.Score,
			SpamTrap:   ip.SpamTrap,
			Blocklists: ip.Blocklists,
			Complaints: ip.Complaints,
			UpdatedAt:  ip.UpdatedAt,
		}
	}

	return GroupResponse{
		ID:            dto.ID,
		GroupID:       dto.GroupID,
		GroupName:     dto.GroupName,
		SpamTrapCount: dto.SpamTrapCount,
		IpsCount:      dto.IPsCount,
		IPs:           ips,
	}
}

func toGroupResponses(dtos []*usecase.GroupDTO) []GroupResponse {
	responses := make([]GroupResponse, len(dtos))
	for i, dto := range dtos {
		responses[i] = toGroupResponse(dto)
	}
	return responses
}

func toIPResponse(dto *usecase.IPDTO) IPResponse {
	return IPResponse{
		ID:         dto.ID,
		IP:         dto.IP,
		Score:      dto.Score,
		SpamTrap:   dto.SpamTrap,
		Blocklists: dto.Blocklists,
		Complaints: dto.Complaints,
		UpdatedAt:  dto.UpdatedAt,
	}
}

func toSubmitScoreResponse(dto *usecase.SubmitScoreResultDTO) SubmitScoreResponse {
	return SubmitScoreResponse{
		Success:        dto.Success,
		Message:        dto.Message,
		IPCreated:      dto.IPCreated,
		GroupCreated:   dto.GroupCreated,
		HistoryAdded:   dto.HistoryAdded,
		HistoryUpdated: dto.HistoryUpdated,
	}
}

func toAddIPsResponse(dto *usecase.BatchIPResultDTO) AddIPsResponse {
	return AddIPsResponse{
		GroupsCreated: dto.GroupsCreated,
		IPsCreated:    dto.IPsCreated,
		IPsSkipped:    dto.IPsSkipped,
		Message:       dto.Message,
	}
}
