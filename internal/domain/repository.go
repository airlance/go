package domain

import "context"

type GroupRepository interface {
	Create(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id uint) (*Group, error)
	GetByGroupID(ctx context.Context, groupID int) (*Group, error)
	List(ctx context.Context, offset, limit int) ([]*Group, int64, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, groupID int) error
	UpdateCounters(ctx context.Context, groupID int) error
	GetGroupIDsByIP(ctx context.Context, ipID uint) ([]int, error)
}

type IPRepository interface {
	Create(ctx context.Context, ip *IP) error
	GetByID(ctx context.Context, id uint) (*IP, error)
	GetByIP(ctx context.Context, ipAddress string) (*IP, error)
	GetOldestIP(ctx context.Context) (*IP, error)
	ListByGroupID(ctx context.Context, groupID int) ([]*IP, error)
	Update(ctx context.Context, ip *IP) error
	Delete(ctx context.Context, id uint) error
	AddToGroup(ctx context.Context, ipID uint, groupID int) error
	RemoveFromGroup(ctx context.Context, ipID uint, groupID int) error
	IsIPInOtherGroups(ctx context.Context, ipID uint, excludeGroupID int) (bool, error)
}

type HistoryRepository interface {
	Create(ctx context.Context, history *History) error
	GetByIPAndDate(ctx context.Context, ipID uint, date string) (*History, error)
	ListByIPID(ctx context.Context, ipID uint) ([]*History, error)
	Update(ctx context.Context, history *History) error
	DeleteByIPID(ctx context.Context, ipID uint) error
}

type ScoreStatRepository interface {
	Create(ctx context.Context, stat *ScoreStat) error
	ListByIPID(ctx context.Context, ipID uint) ([]*ScoreStat, error)
	DeleteByIPID(ctx context.Context, ipID uint) error
}
