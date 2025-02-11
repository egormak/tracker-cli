package task

import "time"

type TaskTimer struct {
	Name         string
	Role         string
	TimeDuration int
	TimeBegin    time.Time
	TimeEnd      time.Time
	TimeDone     float64
	Percent      int
	MsgID        int
}
