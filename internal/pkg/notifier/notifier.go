package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Notify sends a desktop notification on supported OS (macOS, Linux) and sounds terminal bell
func Notify(title, message string) {
	// 1. Terminal audio bell
	fmt.Print("\a")

	// 2. Desktop notification
	switch runtime.GOOS {
	case "darwin":
		// Escape quotes for AppleScript
		safeTitle := strings.ReplaceAll(title, `"`, `\"`)
		safeMessage := strings.ReplaceAll(message, `"`, `\"`)
		script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, safeMessage, safeTitle)
		_ = exec.Command("osascript", "-e", script).Start()

	case "linux":
		// notify-send if installed
		_ = exec.Command("notify-send", title, message).Start()
	}
}

// NotifyTaskCompleted alerts the user that a sprint / task is complete
func NotifyTaskCompleted(taskName, role string, durationMinutes int) {
	title := "⏱ Tracker — Task Completed!"
	message := fmt.Sprintf("Great job! Sprint '%s' (%d min) is finished. Time for a break.", taskName, durationMinutes)
	Notify(title, message)
}

// NotifyRestBalance alerts the user about rest bank balance
func NotifyRestBalance(restMinutes float64) {
	title := "☕ Tracker — Rest Reminder"
	message := fmt.Sprintf("You have %.1f minutes of rest banked. Time to stretch or relax!", restMinutes)
	Notify(title, message)
}
