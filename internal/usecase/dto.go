package usecase

type CreateGroupDTO struct {
	GroupID   int
	GroupName string
}

type GroupDTO struct {
	ID            uint
	GroupID       int
	GroupName     string
	SpamTrapCount int
	IPsCount      int
	IPs           []IPDTO
}

type IPDTO struct {
	ID         uint
	IP         string
	Score      int
	SpamTrap   int
	Blocklists string
	Complaints string
	UpdatedAt  int64
}

type AddIPDTO struct {
	GroupID    int
	GroupName  string
	IP         string
	Score      int
	SpamTrap   int
	Blocklists string
	Complaints string
}

type HistoryEntryDTO struct {
	Date     string
	Score    int
	Volume   int
	SpamTrap int
}

type SubmitScoreDTO struct {
	IP         string
	Score      int
	SpamTrap   int
	Blocklists string
	Complaints string
	History    []HistoryEntryDTO
}

type SubmitScoreResultDTO struct {
	Success        bool
	Message        string
	IPCreated      bool
	GroupCreated   bool
	HistoryAdded   int
	HistoryUpdated int
}

type BatchIPResultDTO struct {
	GroupsCreated int
	IPsCreated    int
	IPsSkipped    int
	Message       string
}

type PaginationDTO struct {
	Page     int
	PageSize int
}
