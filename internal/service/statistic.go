package service

import (
	"fmt"
	"log/slog"
	"tracker_cli/internal/repository/api"

	"github.com/spf13/cobra"
)

func StatisticShow(cmd *cobra.Command, args []string) {

	slog.Info("Start Statistic Query")

	resultRecords := api.GetTaskRecords()
	resultRoleRecords := api.GetRoleRecords()

	// TODO USE bubbletea
	// Show Information
	fmt.Println("####")
	fmt.Println("All Days was Done: ")
	for k, v := range resultRecords["all"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("Yesterday was Done: ")
	for k, v := range resultRecords["yesterday"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("For Today was Done: ")
	for k, v := range resultRecords["today"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("Roles Info: ")
	for k, v := range resultRoleRecords {
		fmt.Printf("Roles: %s, Times: %d\n", k, v)
	}
}
