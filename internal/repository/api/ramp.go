package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"tracker_cli/internal/domain/entity"
)

// GetRampStatus retrieves the current warm-up ramp status from the server.
func GetRampStatus() (entity.RampStatus, error) {
	body, err := sendRequest("GET", "/api/v1/ramp/status", nil)
	if err != nil {
		return entity.RampStatus{}, fmt.Errorf("get ramp status: %w", err)
	}
	defer body.Close()

	var status entity.RampStatus
	if err := json.NewDecoder(body).Decode(&status); err != nil {
		return entity.RampStatus{}, fmt.Errorf("decode ramp status: %w", err)
	}

	return status, nil
}

// ResetRamp resets the warm-up ramp step to 1.
func ResetRamp() (entity.RampStatus, error) {
	body, err := sendRequest("POST", "/api/v1/ramp/reset", nil)
	if err != nil {
		return entity.RampStatus{}, fmt.Errorf("reset ramp: %w", err)
	}
	defer body.Close()

	var status entity.RampStatus
	if err := json.NewDecoder(body).Decode(&status); err != nil {
		return entity.RampStatus{}, fmt.Errorf("decode ramp reset response: %w", err)
	}

	return status, nil
}

// AdvanceRamp manually increments the warm-up ramp step by 1.
func AdvanceRamp() (entity.RampStatus, error) {
	body, err := sendRequest("POST", "/api/v1/ramp/advance", nil)
	if err != nil {
		return entity.RampStatus{}, fmt.Errorf("advance ramp: %w", err)
	}
	defer body.Close()

	var status entity.RampStatus
	if err := json.NewDecoder(body).Decode(&status); err != nil {
		return entity.RampStatus{}, fmt.Errorf("decode ramp advance response: %w", err)
	}

	return status, nil
}

// GetRampConfig retrieves the current ramp configuration.
func GetRampConfig() (entity.RampConfig, error) {
	body, err := sendRequest("GET", "/api/v1/ramp/config", nil)
	if err != nil {
		return entity.RampConfig{}, fmt.Errorf("get ramp config: %w", err)
	}
	defer body.Close()

	var cfg entity.RampConfig
	if err := json.NewDecoder(body).Decode(&cfg); err != nil {
		return entity.RampConfig{}, fmt.Errorf("decode ramp config: %w", err)
	}

	return cfg, nil
}

// UpdateRampConfig updates the warm-up ramp configuration rules.
func UpdateRampConfig(cfg entity.RampConfig) (entity.RampStatus, error) {
	payload, err := json.Marshal(cfg)
	if err != nil {
		return entity.RampStatus{}, fmt.Errorf("marshal ramp config: %w", err)
	}

	body, err := sendRequest("PUT", "/api/v1/ramp/config", bytes.NewBuffer(payload))
	if err != nil {
		return entity.RampStatus{}, fmt.Errorf("update ramp config: %w", err)
	}
	defer body.Close()

	var status entity.RampStatus
	if err := json.NewDecoder(body).Decode(&status); err != nil {
		return entity.RampStatus{}, fmt.Errorf("decode update ramp config response: %w", err)
	}

	return status, nil
}
