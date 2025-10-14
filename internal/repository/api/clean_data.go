package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tracker_cli/internal/domain/entity"
)

func CleanData() (entity.Answer, error) {

	var result entity.Answer

	responceBody, err := sendRequest(http.MethodGet, "/api/v1/records/clean", nil)

	if err != nil {
		return result, fmt.Errorf("api-clean-data error: %w", err)
	}

	defer responceBody.Close()

	if err := json.NewDecoder(responceBody).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil

}
