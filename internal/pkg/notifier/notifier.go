package notifier

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	defaultSound = "Glass"
	roleRest     = "rest"
)

// Seams swapped by tests so no real process is spawned and no bell is written.
var (
	commandStarter = func(name string, args ...string) error {
		return exec.Command(name, args...).Start()
	}
	goos                 = runtime.GOOS
	bellWriter io.Writer = os.Stdout
)

func escapeAppleScript(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}

// NotifyWithSound sends a desktop notification on supported OS (macOS, Linux) and sounds terminal bell.
// soundName is honoured on macOS only (empty falls back to the default); Linux uses plain notify-send.
func NotifyWithSound(title, message, soundName string) {
	// 1. Terminal audio bell
	fmt.Fprint(bellWriter, "\a")

	// 2. Desktop notification
	switch goos {
	case "darwin":
		if soundName == "" {
			soundName = defaultSound
		}
		safeTitle := escapeAppleScript(title)
		safeMessage := escapeAppleScript(message)
		safeSound := escapeAppleScript(soundName)
		script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "%s"`, safeMessage, safeTitle, safeSound)
		_ = commandStarter("osascript", "-e", script)

	case "linux":
		// notify-send if installed
		_ = commandStarter("notify-send", title, message)
	}
}

// Notify sends a desktop notification on supported OS (macOS, Linux) and sounds terminal bell
func Notify(title, message string) {
	NotifyWithSound(title, message, defaultSound)
}

// NotifyTaskCompleted alerts the user that a sprint / task is complete
func NotifyTaskCompleted(taskName, role string, durationMinutes int) {
	if strings.EqualFold(strings.TrimSpace(role), roleRest) {
		NotifyBreakCompleted(taskName, durationMinutes)
		return
	}
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

// NotifyComboStep alerts the user that the evening combo chain is advancing to the next step
func NotifyComboStep(step, total int, nextTask string, durationMinutes int) {
	title := "⚡ Tracker — Evening Combo Advance"
	message := fmt.Sprintf("Step %d/%d: Starting '%s' (%d min). Let's go!", step, total, nextTask, durationMinutes)
	NotifyWithSound(title, message, "Submarine")
}

// NotifyBreakCompleted alerts the user that a rest-role task (a break) has finished
func NotifyBreakCompleted(taskName string, durationMinutes int) {
	title := "☕ Tracker — Break Completed"
	message := fmt.Sprintf("Break '%s' (%d min) is finished. Time to get back into flow!", taskName, durationMinutes)
	NotifyWithSound(title, message, "Ping")
}
