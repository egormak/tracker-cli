package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
)

func GetTaskParams(taskName string) entity.TaskParams {
	var result entity.TaskParams

	responceBody, err := sendRequest("GET", fmt.Sprintf("%s?task_name=%s", "/api/v1/task/params", taskName), nil)
	if err != nil {
		slog.Debug("get task params: task not found or error", "task", taskName, "error", err)
		return entity.TaskParams{}
	}
	defer responceBody.Close()

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode task params response", "error", err)
		return entity.TaskParams{}
	}

	return result
}

func AddTaskRecord(taskName string, timeDone int, sourceDay string) entity.Answer {

	taskRecord := entity.TaskRecorcRequest{
		TaskName:  taskName,
		TimeDone:  timeDone,
		SourceDay: sourceDay,
	}
	var result entity.Answer

	json_data, err := json.Marshal(&taskRecord)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		os.Exit(1)
	}

	responceBody, err := sendRequest("POST", "/api/v1/taskrecord", bytes.NewBuffer(json_data))

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result

}

func GetTaskList() ([]entity.TaskList, error) {
	state, err := GetDashboardState()
	if err != nil {
		return nil, fmt.Errorf("get task list: %w", err)
	}
	return state.TodayTasks, nil
}
