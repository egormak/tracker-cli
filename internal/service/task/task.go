package task

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tracker_cli/internal/pkg/day_method"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/rest"
	"tracker_cli/internal/service/role"
	"tracker_cli/internal/service/statistic"
	"tracker_cli/internal/service/task_params"
	"tracker_cli/internal/service/telegram"
	"tracker_cli/internal/service/timer"
)

type TaskManager struct {
	Name         string
	Role         string
	TimeDuration int
	TimeBegin    time.Time
	TimeEnd      time.Time
	TimeDone     float64
	Procent      int
	MsgID        int
}

func TasksNew(taskName string, timer int, procent int) TaskManager {

	if taskName == "" {
		slog.Error("Task Name is not Set")
		os.Exit(1)
	}

	task := TaskManager{
		Name:         taskName,
		TimeDuration: timer,
		Procent:      procent,
		Role:         role.TaskRoleGet(taskName),
	}
	return task
}

func (t *TaskManager) TasksInit() error {

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh,
		syscall.SIGTERM, // terminate: stopped by `kill -9 PID`
		syscall.SIGINT,  // interrupt: stopped by Ctrl + C
	)

	go func() {
		defer func() {
			exitCh <- syscall.SIGTERM // send terminate signal when
			// application stop naturally
		}()
		t.Start() // start the application
	}()

	<-exitCh // blocking until receive exit signal

	t.Stop() // stop the application

	return nil
}

func (t *TaskManager) Start() {

	// Get Task
	t.TimeBegin = time.Now()
	taskParams := task_params.GetTaskParams(t.Name)
	if t.TimeDuration == 0 {
		t.TimeDuration = timer.TimeDurationGet()
	}
	if taskParams != (task_params.TaskParams{}) {
		taskTimeDone := statistic.StatisticTaskGet(t.Name)
		taskTimeLeft := (taskParams.Time*t.Procent)/100 - taskTimeDone
		if taskTimeLeft <= 0 {
			slog.Info("Time for this task was done")
			os.Exit(0)
		}

		if t.TimeDuration > taskTimeLeft {
			t.TimeDuration = taskTimeLeft
		}
	}

	t.MsgID = telegram.TelegramStartSend(t.Name)

	slog.Info("Start Timer")
	for x := 0; x < t.TimeDuration; x++ {
		slog.Info(fmt.Sprintf("Pass: %d, Left: %d", x, t.TimeDuration-x))
		time.Sleep(time.Minute)
	}

	timer.TimeDurationDel(t.TimeDuration)
	procent.ChangeGroupPlanPercent()

	slog.Info("End Timer")
}

func (t *TaskManager) Stop() {

	// Set value
	t.TimeEnd = time.Now()
	t.TimeDone = t.TimeEnd.Sub(t.TimeBegin).Minutes()
	AddTaskRecord(t.Name, int(t.TimeDone))

	statistic.StatisticTaskShow(t.Name)
	if day_method.IsWeekendNow() || t.Role != "rest" {
		statistic.StatisticFullShow()
		rest.RestShow()
	}
	telegram.TelegramStopSend(t.Name, t.MsgID, int(t.TimeDone), t.TimeEnd.Format("2 January 2006 15:04"))
}
