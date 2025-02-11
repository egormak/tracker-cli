package task

import (
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"

	"github.com/spf13/cobra"
)

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

	taskApp := NewTaskProcess(taskName, taskTime, taskPercent)
	taskApp.Prepare()
}

// NewTaskProcess initializes a new TaskInfo object with the provided parameters
func NewTaskProcess(name string, timeDuration, percent int) *TaskTimer {
	// Fetch task parameters and completed task time
	taskParams := api.GetTaskParams(name)
	taskDone := api.StatisticTaskGet(name)

	// Determine the appropriate time duration for the task
	if timeDuration == 0 {
		slog.Info("Time duration is 0, get time from task params")
		timeDurationApi := api.TimeDurationGet()
		if taskParams == (entity.TaskParams{}) {
			timeDuration = timeDurationApi
		} else {
			timeTaskParamsLeft := (taskParams.TimePlanDuration*percent)/100 - taskDone
			if timeTaskParamsLeft >= timeDurationApi {
				timeDuration = timeDurationApi
			} else {
				timeDuration = timeTaskParamsLeft
			}
		}
	} else {
		if taskParams != (entity.TaskParams{}) {
			timeTaskParamsLeft := (taskParams.TimePlanDuration*percent)/100 - taskDone
			if timeTaskParamsLeft <= timeDuration {
				timeDuration = timeTaskParamsLeft
			}
		}
	}

	// Exit if the calculated time duration is non-positive
	if timeDuration <= 0 {
		slog.Info("Time for this task was done")
		os.Exit(0)
	}

	// Return a new TaskInfo object
	return &TaskTimer{
		Name:         name,
		Role:         api.TaskRoleGet(name),
		TimeDuration: timeDuration,
	}
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

// func (t *TaskInfo) Run() error {

// 	// Set Params
// 	t.TimeBegin = time.Now()

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

// func (t *TaskInfo) Start() {

// 	// Get Task
// 	taskParams := task_params.GetTaskParams(t.Name)

// 	if taskParams != (task_params.TaskParams{}) {
// 		taskTimeDone := statistic.StatisticTaskGet(t.Name)
// 		taskTimeLeft := (taskParams.Time*t.Percent)/100 - taskTimeDone
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

// func (t *TaskInfo) Stop() {

// 	// Set value
// 	t.TimeEnd = time.Now()
// 	t.TimeDone = t.TimeEnd.Sub(t.TimeBegin).Minutes()
// 	api.AddTaskRecord(t.Name, int(t.TimeDone))

// 	statistic.StatisticTaskShow(t.Name)
// 	statistic.StatisticFullShow()
// 	rest.RestShow()
// 	telegram.TelegramStopSend(t.Name, t.MsgID, int(t.TimeDone), t.TimeEnd.Format("2 January 2006 15:04"))
// }
