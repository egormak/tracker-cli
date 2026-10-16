package evening

import (
	"errors"
	"fmt"
	"io"
	"strings"

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

// chainOptions describes how a chain's sprints are started and labelled.
type chainOptions struct {
	origin    string // "combo" or "batch", sent to the server on start
	debtShare int    // 0–100: % of each sprint that pays the task's past debt
	label     string // shown in the progress lines
	// firstDuration, when > 0, replaces the sprint length for the first task (the batch anchor).
	firstDuration int
}

// Seams swapped by tests so the combo loop runs without a backend or desktop notifications.
var (
	comboSprintCreator = func(name string, duration int, opts chainOptions) (comboSprint, error) {
		// Combo sprint duration is derived from weekly deficit, not today's daily
		// target, so percentSpecified=false bypasses the daily-target timeLeft cap.
		taskTimer, err := task.CreateTaskTimerWithPercentFlag(name, duration, 100, false)
		if err != nil {
			return comboSprint{}, err
		}
		// RunCombo sends one notification per transition itself.
		taskTimer.SetCompletionNotify(false)
		taskTimer.Origin = opts.origin
		taskTimer.DebtShare = opts.debtShare

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
	return runChain(out, candidates, comboDuration, chainOptions{origin: "combo", label: "Combo"})
}

// RunBatch runs an evening batch: the anchor for anchorMinutes (0 = sprintMinutes), then the
// next candidates as sprints of sprintMinutes each; every sprint sends debtShare % of its
// minutes to its task's past debt.
func RunBatch(out io.Writer, tasks []entity.EveningFocusCandidate, sprintMinutes, anchorMinutes, debtShare int) error {
	return runChain(out, tasks, sprintMinutes, chainOptions{
		origin:        "batch",
		debtShare:     debtShare,
		label:         "Батч",
		firstDuration: anchorMinutes,
	})
}

// BuildBatch returns the batch's tasks: the anchor first (the first candidate when anchor
// is empty), then the next candidates in the server's order, n tasks at most.
func BuildBatch(candidates []entity.EveningFocusCandidate, anchor string, n int) []entity.EveningFocusCandidate {
	if n <= 0 {
		return nil
	}
	var out []entity.EveningFocusCandidate
	if anchor != "" {
		first := entity.EveningFocusCandidate{TaskName: anchor}
		for _, c := range candidates {
			if strings.EqualFold(c.TaskName, anchor) {
				first = c
				break
			}
		}
		out = append(out, first)
	}
	for _, c := range candidates {
		if len(out) >= n {
			break
		}
		if anchor != "" && strings.EqualFold(c.TaskName, anchor) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// BatchCapacity is the largest batch BuildBatch can fill: the anchor (when given) plus
// every candidate other than the anchor.
func BatchCapacity(candidates []entity.EveningFocusCandidate, anchor string) int {
	if anchor == "" {
		return len(candidates)
	}
	n := 1
	for _, c := range candidates {
		if !strings.EqualFold(c.TaskName, anchor) {
			n++
		}
	}
	return n
}

func runChain(out io.Writer, candidates []entity.EveningFocusCandidate, comboDuration int, opts chainOptions) error {
	var pending *finishedSprint
	defer func() {
		if pending != nil {
			notifyTaskCompleted(pending.name, pending.role, pending.minutes)
		}
	}()

	for i, candidate := range candidates {
		duration := comboDuration
		if i == 0 && opts.firstDuration > 0 {
			duration = opts.firstDuration
		}
		sprint, err := comboSprintCreator(candidate.TaskName, duration, opts)
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
		fmt.Fprintf(out, "⚡️ [%s %d/%d] Starting sprint on '%s' (%d min)...\n", opts.label, i+1, len(candidates), candidate.TaskName, sprint.duration)

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

	if opts.origin == "batch" {
		fmt.Fprintln(out, "🎉 Батч завершён!")
	} else {
		fmt.Fprintln(out, "🎉 Evening Combo Chain completed successfully!")
	}
	return nil
}
