package task

import (
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"

	"github.com/spf13/cobra"
)

// ErrTaskCompleted indicates that no time remains for the task
var ErrTaskCompleted = fmt.Errorf("no time remaining for this task")

func TaskRun(cmd *cobra.Command, args []string) {
	fmt.Println("Run Task")
	taskName, err := cmd.Flags().GetString("name")
	if err != nil {
		slog.Error("failed to get task name", "error", err)
	}
	taskTime, err := cmd.Flags().GetInt("time")
	if err != nil {
		slog.Error("failed to get task time", "error", err)
	}
	taskPercent, err := cmd.Flags().GetInt("percent")
	if err != nil {
		slog.Error("failed to get task percent", "error", err)
	}

	taskApp := CreateTaskTimer(taskName, taskTime, taskPercent)
	taskApp.Run()
}

// CreateTaskTimer initializes a new TaskTimer object with the provided parameters
func CreateTaskTimer(name string, requestedDuration, percent int) *TaskTimer {
	taskParams := api.GetTaskParams(name)
	taskDone := api.StatisticTaskGet(name)

	duration, err := calculateDuration(taskParams, requestedDuration, percent, taskDone)
	if err != nil {
		slog.Error("calculating duration", "error", err)
		os.Exit(1)
	}

	// Return a new TaskTimer object
	return &TaskTimer{
		Name:         name,
		Role:         api.TaskRoleGet(name),
		TimeDuration: duration,
	}
}

// calculateDuration determines the appropriate time duration for the task
func calculateDuration(params entity.TaskParams, requested, percent, done int) (int, error) {
	if requested == 0 {
		return calculateDefaultDuration(params, percent, done)
	}
	return calculateRequestedDuration(params, requested, percent, done)

}

// calculateDefaultDuration handles the case when no specific duration is requested
func calculateDefaultDuration(params entity.TaskParams, percent, done int) (int, error) {
	apiDuration := api.TimeDurationGet()
	slog.Info("using default duration from API", "duration", apiDuration)

	if params == (entity.TaskParams{}) {
		return apiDuration, nil
	}

	timeLeft := calculateTimeLeft(params.Time, percent, done)
	if timeLeft <= 0 {
		return 0, ErrTaskCompleted
	}

	if timeLeft >= apiDuration {
		return apiDuration, nil
	}
	return timeLeft, nil
}

// calculateRequestedDuration handles the case when a specific duration is requested
func calculateRequestedDuration(params entity.TaskParams, requested, percent, done int) (int, error) {
	if params == (entity.TaskParams{}) {
		return requested, nil
	}
	fmt.Println("Time Duration: ", params.Time)
	timeLeft := calculateTimeLeft(params.Time, percent, done)
	if timeLeft <= 0 {
		return 0, ErrTaskCompleted
	}

	if timeLeft < requested {
		return timeLeft, nil
	}
	return requested, nil
}

// calculateTimeLeft calculates remaining time based on plan duration, percentage and time already spent
func calculateTimeLeft(planDuration, percent, done int) int {
	return (planDuration*percent)/100 - int(done)
}

func (t *TaskTimer) Prepare() {

	fmt.Println("Prepare Task")
	fmt.Println("Name: ", t)

	// Get Task
	// taskParamsTime := api.GetTaskParamsTime(t.Name)

	// if taskParamsTime != 0 {

	// }
	// taskParams := task_params.GetTaskParams(t.Name)

	// if taskParams != (task_params.TaskParams{}) {
	// 	taskTimeDone := statistic.StatisticTaskGet(t.Name)
	// 	taskTimeLeft := (taskParams.Time*t.Percent)/100 - taskTimeDone
	// 	if taskTimeLeft <= 0 {
	// 		slog.Info("Time for this task was done")
	// 		os.Exit(0)
	// 	}

	// 	if t.TimeDuration > taskTimeLeft {
	// 		t.TimeDuration = taskTimeLeft
	// 	}
	// }

}

// type TaskManager struct {
// 	Name         string
// 	Role         string
// 	TimeDuration int
// 	TimeBegin    time.Time
// 	TimeEnd      time.Time
// 	TimeDone     float64
// 	Procent      int
// 	MsgID        int
// }

// func TasksNew(taskName string, timer int, procent int) TaskManager {

// 	if taskName == "" {
// 		slog.Error("Task Name is not Set")
// 		os.Exit(1)
// 	}

// 	task := TaskManager{
// 		Name:         taskName,
// 		TimeDuration: timer,
// 		Procent:      procent,
// 		Role:         role.TaskRoleGet(taskName),
// 	}
// 	return task
// }

// func (t *TaskManager) TasksInit() error {

// 	exitCh := make(chan os.Signal)
// 	signal.Notify(exitCh,
// 		syscall.SIGTERM, // terminate: stopped by `kill -9 PID`
// 		syscall.SIGINT,  // interrupt: stopped by Ctrl + C
// 	)

// 	go func() {
// 		defer func() {
// 			exitCh <- syscall.SIGTERM // send terminate signal when
// 			// application stop naturally
// 		}()
// 		t.Start() // start the application
// 	}()

// 	<-exitCh // blocking until receive exit signal

// 	t.Stop() // stop the application

// 	return nil
// }

// func (t *TaskManager) Start() {

// 	// Get Task
// 	t.TimeBegin = time.Now()
// 	taskParams := task_params.GetTaskParams(t.Name)
// 	if t.TimeDuration == 0 {
// 		t.TimeDuration = timer.TimeDurationGet()
// 	}
// 	if taskParams != (task_params.TaskParams{}) {
// 		taskTimeDone := statistic.StatisticTaskGet(t.Name)
// 		taskTimeLeft := (taskParams.Time*t.Procent)/100 - taskTimeDone
// 		if taskTimeLeft <= 0 {
// 			slog.Info("Time for this task was done")
// 			os.Exit(0)
// 		}

// 		if t.TimeDuration > taskTimeLeft {
// 			t.TimeDuration = taskTimeLeft
// 		}
// 	}

// 	t.MsgID = telegram.TelegramStartSend(t.Name)

// 	slog.Info("Start Timer")
// 	for x := 0; x < t.TimeDuration; x++ {
// 		slog.Info(fmt.Sprintf("Pass: %d, Left: %d", x, t.TimeDuration-x))
// 		time.Sleep(time.Minute)
// 	}

// 	timer.TimeDurationDel(t.TimeDuration)
// 	procent.ChangeGroupPlanPercent()

// 	slog.Info("End Timer")
// }

// func (t *TaskManager) Stop() {

// 	// Set value
// 	t.TimeEnd = time.Now()
// 	t.TimeDone = t.TimeEnd.Sub(t.TimeBegin).Minutes()
// 	AddTaskRecord(t.Name, int(t.TimeDone))

// 	statistic.StatisticTaskShow(t.Name)
// 	statistic.StatisticFullShow()
// 	rest.RestShow()
// 	telegram.TelegramStopSend(t.Name, t.MsgID, int(t.TimeDone), t.TimeEnd.Format("2 January 2006 15:04"))
// }

///////////////////////////////////////
