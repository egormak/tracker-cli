package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"tracker_cli/internal/domain/entity"
)

// GetWeeklyDigest fetches the weekly digest; week is "current" or "previous".
func GetWeeklyDigest(week string) (entity.WeeklyDigest, error) {
	path := "/api/v1/stats/weekly/digest"
	if week != "" {
		path += "?" + url.Values{"week": {week}}.Encode()
	}

	body, err := sendRequest("GET", path, nil)
	if err != nil {
		return entity.WeeklyDigest{}, err
	}
	defer body.Close()

	var digest entity.WeeklyDigest
	if err := json.NewDecoder(body).Decode(&digest); err != nil {
		return entity.WeeklyDigest{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return digest, nil
}
