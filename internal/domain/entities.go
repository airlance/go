package domain

import "time"

type Group struct {
	ID            uint
	GroupID       int
	GroupName     string
	SpamTrapCount int
	IPsCount      int
}

type IP struct {
	ID         uint
	IP         string
	Score      int
	SpamTrap   int
	Blocklists string
	Complaints string
	UpdatedAt  time.Time
	GroupIDs   []int
}

type History struct {
	ID       uint
	IPsID    uint
	Score    int
	SpamTrap int
	Volume   int
	Time     time.Time
}

type ScoreStat struct {
	ID     uint
	IPsID  uint
	Score  int
	Result int
	Date   time.Time
}
