package evening

import (
	"errors"
	"fmt"
	"io"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/pkg/notifier"
	"tracker_cli/internal/service/task"
)

// comboSprint is one prepared (not yet started) sprint of the combo chain.
type comboSprint struct {
	role     string
	duration int
	// run blocks until the sprint ends and reports whether it reached its full
	// duration and how many minutes were done.
	run func() (completed bool, doneMinutes int, err error)
}

// Seams swapped by tests so the combo loop runs without a backend or desktop notifications.
var (
	comboSprintCreator = func(name string, duration int) (comboSprint, error) {
		// Combo sprint duration is derived from weekly deficit, not today's daily
		// target, so percentSpecified=false bypasses the daily-target timeLeft cap.
		taskTimer, err := task.CreateTaskTimerWithPercentFlag(name, duration, 100, false)
		if err != nil {
			return comboSprint{}, err
		}
		// RunCombo sends one notification per transition itself.
		taskTimer.SetCompletionNotify(false)

		return comboSprint{
			role:     taskTimer.Role,
			duration: taskTimer.TimeDuration,
			run: func() (bool, int, error) {
				err := taskTimer.Run()
				return taskTimer.Completed(), taskTimer.DoneMinutes(), err
			},
		}, nil
	}
	notifyComboStep     = notifier.NotifyComboStep
	notifyTaskCompleted = notifier.NotifyTaskCompleted
)

// finishedSprint is a sprint that ran to completion and has not been announced yet.
type finishedSprint struct {
	name    string
	role    string
	minutes int
}

// RunCombo runs the candidates as a chain of sprints of comboDuration minutes each.
//
// Exactly one desktop notification is sent per completed sprint: a "combo advance"
// when the next sprint actually starts, or the regular "task completed" when the
// chain ends there. Candidates with no remaining time and sprints stopped early
// produce no notification.
func RunCombo(out io.Writer, candidates []entity.EveningFocusCandidate, comboDuration int) error {
	var pending *finishedSprint
	defer func() {
		if pending != nil {
			notifyTaskCompleted(pending.name, pending.role, pending.minutes)
		}
	}()

	for i, candidate := range candidates {
		sprint, err := comboSprintCreator(candidate.TaskName, comboDuration)
		if err != nil {
			if errors.Is(err, task.ErrTaskCompleted) {
				fmt.Fprintf(out, "Task %s has no remaining time.\n", candidate.TaskName)
				continue
			}
			return err
		}

		if pending != nil {
			notifyComboStep(i+1, len(candidates), candidate.TaskName, sprint.duration)
			pending = nil
		}
		fmt.Fprintf(out, "⚡️ [Combo %d/%d] Starting sprint on '%s' (%d min)...\n", i+1, len(candidates), candidate.TaskName, sprint.duration)

		completed, doneMinutes, err := sprint.run()
		if completed {
			pending = &finishedSprint{name: candidate.TaskName, role: sprint.role, minutes: doneMinutes}
		}
		if err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return nil
			}
			return err
		}

		if completed {
			fmt.Fprintf(out, "✅ Sprint on '%s' completed!\n", candidate.TaskName)
		} else {
			fmt.Fprintf(out, "⏹ Sprint on '%s' stopped early.\n", candidate.TaskName)
		}
	}

	fmt.Fprintln(out, "🎉 Evening Combo Chain completed successfully!")
	return nil
}
