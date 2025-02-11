package task

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/rest"
	"tracker_cli/internal/service/statistic"
	"tracker_cli/internal/service/telegram"
	"tracker_cli/internal/service/timer"
)

func (t *TaskTimer) Run() error {

	// Set Params
	t.TimeBegin = time.Now()

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

func (t *TaskTimer) Start() {

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

func (t *TaskTimer) Stop() {

	// Set value
	t.TimeEnd = time.Now()
	t.TimeDone = t.TimeEnd.Sub(t.TimeBegin).Minutes()
	api.AddTaskRecord(t.Name, int(t.TimeDone))

	statistic.StatisticTaskShow(t.Name)
	statistic.StatisticFullShow()
	rest.RestShow()
	telegram.TelegramStopSend(t.Name, t.MsgID, int(t.TimeDone), t.TimeEnd.Format("2 January 2006 15:04"))
}
