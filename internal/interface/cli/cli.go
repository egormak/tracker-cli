package cli

import (
	"flag"
	"log/slog"
	"os"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/role"
	"tracker_cli/internal/service/task_params"
	"tracker_cli/internal/service/timer"
)

type ParamsDataStruct struct {
	menu          *bool
	plan          *bool
	priority      *int
	percent       *int
	percentPlan   *bool
	percentsSet   *string
	roleRecheck   *bool
	taskConfig    *bool
	taskName      *string
	taskNameAdd   *string
	taskNameDel   *string
	taskNameRole  *string
	taskRecordAdd *bool
	timer         *int
	timerListSet  *int
	timerRecheck  *bool
}

func NewParams() (*ParamsDataStruct, error) {

	p := &ParamsDataStruct{}
	// Main Command
	// General
	p.percentsSet = flag.String("percents_set", "", "Set Percents from List")
	p.roleRecheck = flag.Bool("recheck", false, "Recheck role statistics")
	p.taskConfig = flag.Bool("config", false, "Set Config Option")
	p.taskRecordAdd = flag.Bool("taskrecordadd", false, "Add task record")

	// Logic for Run TaskManager
	p.menu = flag.Bool("menu", false, "Run Menu")
	p.percentPlan = flag.Bool("percent_plan", false, "Run Percent Plan")
	p.plan = flag.Bool("plan", false, "Run Task from Plan on Day")

	// Manage Task List
	p.taskNameAdd = flag.String("taskadd", "", "Add available Task")
	p.taskNameDel = flag.String("taskdel", "", "Remove available Task")

	// Timer Logic
	p.timerListSet = flag.Int("timerlistset", 0, "Create Timer List with Count")
	p.timerRecheck = flag.Bool("timerrecheck", false, "Timer List Recheck")

	// Sub Command
	p.priority = flag.Int("priority", 0, "Set Task Priority")
	p.percent = flag.Int("percent", 100, "Set Percent Plan Task")
	p.taskName = flag.String("task", "", "Set Task Name")
	p.taskNameRole = flag.String("taskrole", "", "Task Role")
	p.timer = flag.Int("time", 0, "Set minutes for Timer")

	flag.Parse()

	return p, nil

}

func (p *ParamsDataStruct) RunSystemCommand() {

	// ProcentSets
	if *p.percentsSet != "" {
		procent.ProcentSets(*p.percentsSet, *p.taskName)
		os.Exit(0)
	}

	// Recheck all Task Statistics
	if *p.roleRecheck {
		role.RoleRecheck()
		os.Exit(0)
	}

	// Set New TimeList
	if *p.timerListSet != 0 {
		timer.TimeListSet(*p.timerListSet)
		os.Exit(0)
	}

	// Time Recheck
	if *p.timerRecheck {
		timer.TimerRecheck()
		os.Exit(0)
	}
	// Configure Task
	if *p.taskConfig {
		if *p.taskName == "" {
			slog.Info("Set Global Time")
			timer.SetGlobalTime(*p.timer)
		} else {
			slog.Info("Set Params for Tasks")
			task_params.SetTaskParams(*p.taskName, *p.timer, *p.priority)
		}
		os.Exit(0)
	}

}

// func (p *ParamsDataStruct) RunService() {
// 	// Run Service

// 	if *p.percentPlan {
// 		service.PlanPercent()
// 	}

// 	if *p.plan {
// 		taskPlan := task.GetTaskDay(*p.percent)
// 		if taskPlan.Name == "" {
// 			slog.Info("Task Name can't Get, maybe all plan was done.")
// 			os.Exit(0)
// 		}
// 		slog.Info(fmt.Sprintf("\033[33mTaskName:\033[32m %s\033[0m", taskPlan.Name))
// 		time.Sleep(time.Second * 15)
// 		taskPlan.TasksInit()
// 	}

// 	if *p.taskName != "" {
// 		taskName := *p.taskName
// 		taskConfig := task.TasksNew(taskName, *p.timer, *p.percent)
// 		taskConfig.TasksInit()
// 	}

// if *p.menu {
// 	taskName = menu.RunMenu()
// }

// // Init Service
// taskD = TasksNew(taskName, *p.timer, *p.timerRandom, *p.procent)
// taskD.TasksInit()
// }

// func (p *ParamsDataStruct) RunService() {

// 	// Add TaskName
// 	if *p.taskNameAdd != "" {
// 		if *p.taskNameRole == "" {
// 			log.Fatal("Task Role is not Set")
// 		}
// 		log.Info("Add Task Name")
// 		db_controller.TaskNameAdd(*p.taskNameAdd, *p.taskNameRole)
// 		return
// 	}

// 	// Remove TaskName
// 	if *p.taskNameDel != "" {
// 		db_controller.TaskNameRemove(*p.taskNameDel)
// 		return
// 	}

// 	// Added Record for Task
// 	if *p.taskRecordAdd {
// 		task.AddTaskRecord(*p.taskName, *p.timer)
// 		return
// 	}

// }
