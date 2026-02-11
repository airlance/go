package data

import "time"

type GroupModel struct {
	ID            uint   `gorm:"primaryKey;comment:ID"`
	GroupID       int    `gorm:"uniqueIndex;comment:Group ID"`
	GroupName     string `gorm:"type:varchar(255);comment:Group Name"`
	SpamTrapCount int    `gorm:"default:0;index:idx_group_counts;comment:Spam Trap Count"`
	IPsCount      int    `gorm:"default:0;index:idx_group_counts;comment:IPs Count"`
}

func (GroupModel) TableName() string {
	return "sender_score_groups"
}

type IPModel struct {
	ID         uint      `gorm:"primaryKey;comment:ID"`
	IP         string    `gorm:"type:char(16);uniqueIndex:idx_unique_ip;comment:IP"`
	Score      int       `gorm:"default:0;index:idx_ips_score_trap;comment:Score"`
	SpamTrap   int       `gorm:"default:0;index:idx_ips_score_trap;comment:Spam Trap"`
	Blocklists string    `gorm:"type:varchar(50);comment:Blocklists"`
	Complaints string    `gorm:"type:varchar(50);comment:Complaints"`
	UpdatedAt  time.Time `gorm:"index:idx_ips_updated;comment:Updated"`

	Groups []GroupModel `gorm:"many2many:sender_score_group_ips;joinForeignKey:IPID;joinReferences:GroupID;"`
}

func (IPModel) TableName() string {
	return "sender_score_ips"
}

type GroupIPModel struct {
	IPID    uint `gorm:"primaryKey;index;comment:IP ID"`
	GroupID uint `gorm:"primaryKey;index;comment:Group Internal ID"`

	IP    IPModel    `gorm:"foreignKey:IPID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Group GroupModel `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (GroupIPModel) TableName() string {
	return "sender_score_group_ips"
}

type HistoryModel struct {
	ID       uint      `gorm:"primaryKey;comment:ID"`
	IpsID    uint      `gorm:"not null;index;comment:IPs ID"`
	Score    int       `gorm:"default:0;index:idx_hist_score;comment:Score"`
	SpamTrap int       `gorm:"default:0;index:idx_hist_trap;comment:Spam Trap"`
	Volume   int       `gorm:"default:0;comment:Volume"`
	Time     time.Time `gorm:"type:date;index:idx_hist_time;comment:Date"`

	IPRecord IPModel `gorm:"foreignKey:IpsID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (HistoryModel) TableName() string {
	return "sender_score_histories"
}

type ScoreStatModel struct {
	ID     uint      `gorm:"primaryKey;comment:ID"`
	IpsID  uint      `gorm:"not null;index;comment:IPs ID"`
	Score  int       `gorm:"default:0;index:idx_stats_score;comment:Score"`
	Result int       `gorm:"default:0;index:idx_stats_result;comment:Result"`
	Date   time.Time `gorm:"type:date;index:idx_stats_date;comment:Date"`

	IPRecord IPModel `gorm:"foreignKey:IpsID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ScoreStatModel) TableName() string {
	return "sender_score_score_stats"
}
