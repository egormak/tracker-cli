package entity

import "time"

type TaskManager struct {
	Name         string
	Role         string
	TimeDuration int
	TimeBegin    time.Time
	TimeEnd      time.Time
	TimeDone     float64
	Percent      int
	MsgID        int
}

type TaskPercent struct {
	Name    string `json:"task_name"`
	Percent int    `json:"percent"`
}

type TaskParams struct {
	Name     string `json:"name"`
	Time     int    `json:"time"`
	Priority int    `json:"priority"`
}

type TaskRecorcRequest struct {
	TaskName string `json:"task_name"`
	TimeDone int    `json:"time_done"`
}
