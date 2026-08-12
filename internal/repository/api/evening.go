package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"tracker_cli/internal/domain/entity"
)

func GetEveningFocus(category string, sprintTime int) (entity.EveningFocusResponse, error) {
	v := url.Values{}
	if category != "" {
		v.Set("category", category)
	}
	if sprintTime > 0 {
		v.Set("time", fmt.Sprintf("%d", sprintTime))
	}

	path := "/api/v1/mode/evening-focus"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}

	body, err := sendRequest("GET", path, nil)
	if err != nil {
		return entity.EveningFocusResponse{}, err
	}
	defer body.Close()

	var apiResp entity.EveningFocusAPIResponse
	if err := json.NewDecoder(body).Decode(&apiResp); err != nil {
		return entity.EveningFocusResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.Data, nil
}

func SkipEveningTask(taskName string, category string, sprintTime int) (entity.EveningFocusResponse, error) {
	v := url.Values{}
	if category != "" {
		v.Set("category", category)
	}
	if sprintTime > 0 {
		v.Set("time", fmt.Sprintf("%d", sprintTime))
	}

	path := "/api/v1/mode/evening-focus/skip"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}

	reqPayload := map[string]string{"task_name": taskName}
	jsonBytes, _ := json.Marshal(reqPayload)

	body, err := sendRequest("POST", path, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return entity.EveningFocusResponse{}, err
	}
	defer body.Close()

	var apiResp entity.EveningFocusAPIResponse
	if err := json.NewDecoder(body).Decode(&apiResp); err != nil {
		return entity.EveningFocusResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.Data, nil
}
