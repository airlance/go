package http

type CreateGroupRequest struct {
	GroupID   int    `json:"group_id" binding:"required"`
	GroupName string `json:"group_name" binding:"required"`
}

type AddIPRequest struct {
	GroupID   int    `json:"group_id" binding:"required"`
	GroupName string `json:"group_name"`
	IP        string `json:"ip"`
}

type AddIPsRequest struct {
	IPs []AddIPRequest `json:"ips" binding:"required,min=1,dive"`
}

type HistoryEntry struct {
	Date     string `json:"date" binding:"required"`
	Score    int    `json:"score" binding:"required"`
	Volume   int    `json:"volume" binding:"required"`
	SpamTrap int    `json:"spam_trap"`
}

type SubmitScoreRequest struct {
	IP         string         `json:"ip" binding:"required"`
	Score      int            `json:"score" binding:"required"`
	SpamTrap   int            `json:"spam_trap"`
	Blocklists string         `json:"blocklists"`
	Complaints string         `json:"complaints"`
	History    []HistoryEntry `json:"history" binding:"required,min=1,dive"`
}

type GroupResponse struct {
	ID            uint         `json:"id"`
	GroupID       int          `json:"group_id"`
	GroupName     string       `json:"group_name"`
	SpamTrapCount int          `json:"spam_trap_count"`
	IpsCount      int          `json:"ips_count"`
	IPs           []IPResponse `json:"ips,omitempty"`
}

type IPResponse struct {
	ID         uint   `json:"id"`
	IP         string `json:"ip"`
	Score      int    `json:"score"`
	SpamTrap   int    `json:"spam_trap"`
	Blocklists string `json:"blocklists,omitempty"`
	Complaints string `json:"complaints,omitempty"`
	UpdatedAt  int64  `json:"updated_at"`
}

type SubmitScoreResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	IPCreated      bool   `json:"ip_created"`
	GroupCreated   bool   `json:"group_created"`
	HistoryAdded   int    `json:"history_added"`
	HistoryUpdated int    `json:"history_updated"`
}

type AddIPsResponse struct {
	GroupsCreated int    `json:"groups_created"`
	IPsCreated    int    `json:"ips_created"`
	IPsSkipped    int    `json:"ips_skipped"`
	Message       string `json:"message"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}
