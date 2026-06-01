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
	Name      string `json:"task_name"`
	Percent   int    `json:"percent"`
	TimeLeft  int    `json:"time_left"`
	SourceDay string `json:"source_day,omitempty"` // Optional: day this task is from (for rollover tasks)
}

type TaskParams struct {
	Name     string `json:"name"`
	Time     int    `json:"time"`
	Priority int    `json:"priority"`
}

type TaskRecorcRequest struct {
	TaskName  string `json:"task_name"`
	TimeDone  int    `json:"time_done"`
	SourceDay string `json:"source_day,omitempty"` // Optional: day to record against (for rollover tasks)
}

type TaskList struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	TimeDuration int    `json:"time_duration"`
	TimeDone     int    `json:"time_done"`
	Priority     int    `json:"priority"`
}

type RunningTask struct {
	ID                string    `json:"id"`
	TaskName          string    `json:"task_name"`
	Role              string    `json:"role"`
	StartTime         time.Time `json:"start_time"`
	Accumulated       int       `json:"accumulated"` // accumulated minutes
	IsRunning         bool      `json:"is_running"`
	TargetDuration    int       `json:"target_duration"`
	SourceDay         string    `json:"source_day"`
	TelegramMessageID int       `json:"telegram_message_id"`
}

type TaskRecord struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	TimeDuration int    `json:"time_duration"`
	Date         string `json:"date"`
	SourceDay    string `json:"source_day"`
}
