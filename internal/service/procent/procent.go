package procent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tracker_cli/config"
)

func ProcentSets(procentsStr string, roleName string) {

	var procents []int
	type body struct {
		RoleName string `json:"role_name"`
		Procents []int  `json:"procents"`
	}

	for _, v := range strings.Split(procentsStr, ",") {
		procent, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal(err)
		}
		procents = append(procents, procent)
	}

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	values := body{RoleName: roleName, Procents: procents}
	// struct{}{"procents": procents, "role_name": roleName}
	json_data, err := json.Marshal(&values)
	if err != nil {
		log.Fatal(err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/procents"), bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Errorf("request error, status code: %d", resp.StatusCode))
	}

}

func ChangeGroupPlanPercent() {

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/task/plan-percent/change"), nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Errorf("request error, status code: %d", resp.StatusCode))
	}
}
