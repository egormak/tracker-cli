// Console Tracker
package main

import (
	"log/slog"
	"os"
	"tracker_cli/cmd/command"

	"github.com/lmittmann/tint"
)

// Run Tracker Manager
func main() {

	// create a new logger
	logger := slog.New(tint.NewHandler(os.Stderr, nil))

	slog.SetDefault(logger)
	slog.Info("Start Application")

	command.Execute()

	// app, err := cli.NewParams()
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	os.Exit(1)
	// }

	// app.RunSystemCommand()
	// app.RunService()
}
