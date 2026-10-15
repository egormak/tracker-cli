package evening

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/service/task"
)

// fakeSprint describes how one candidate behaves in the combo chain.
type fakeSprint struct {
	createErr error
	completed bool
	runErr    error
}

// fakeCombo swaps the combo seams and returns the recorded notifications and started sprints.
var lastOptions = &chainOptions{}

func fakeCombo(t *testing.T, sprints map[string]fakeSprint) (notifications *[]string, started *[]string) {
	t.Helper()

	origCreator, origStep, origCompleted := comboSprintCreator, notifyComboStep, notifyTaskCompleted
	t.Cleanup(func() {
		comboSprintCreator, notifyComboStep, notifyTaskCompleted = origCreator, origStep, origCompleted
	})

	notifications = &[]string{}
	started = &[]string{}

	comboSprintCreator = func(name string, duration int, opts chainOptions) (comboSprint, error) {
		*lastOptions = opts
		fake, ok := sprints[name]
		if !ok {
			t.Fatalf("unexpected sprint creation for %q", name)
		}
		if fake.createErr != nil {
			return comboSprint{}, fake.createErr
		}
		return comboSprint{
			role:     "work",
			duration: duration,
			run: func() (bool, int, error) {
				*started = append(*started, name)
				return fake.completed, duration, fake.runErr
			},
		}, nil
	}
	notifyComboStep = func(step, total int, nextTask string, durationMinutes int) {
		*notifications = append(*notifications, fmt.Sprintf("step %d/%d %s %d", step, total, nextTask, durationMinutes))
	}
	notifyTaskCompleted = func(taskName, role string, durationMinutes int) {
		*notifications = append(*notifications, fmt.Sprintf("completed %s %s %d", taskName, role, durationMinutes))
	}

	return notifications, started
}

func comboCandidates(names ...string) []entity.EveningFocusCandidate {
	candidates := make([]entity.EveningFocusCandidate, 0, len(names))
	for _, name := range names {
		candidates = append(candidates, entity.EveningFocusCandidate{TaskName: name})
	}
	return candidates
}

func TestRunCombo(t *testing.T) {
	errBackend := errors.New("backend down")

	tests := []struct {
		name              string
		sprints           map[string]fakeSprint
		wantErr           error
		wantStarted       []string
		wantNotifications []string
		wantChainFinished bool
	}{
		{
			name: "all sprints complete: one banner per transition plus final completion",
			sprints: map[string]fakeSprint{
				"a": {completed: true},
				"b": {completed: true},
				"c": {completed: true},
			},
			wantStarted:       []string{"a", "b", "c"},
			wantNotifications: []string{"step 2/3 b 10", "step 3/3 c 10", "completed c work 10"},
			wantChainFinished: true,
		},
		{
			name: "first candidate has no remaining time: no advance before the first real sprint",
			sprints: map[string]fakeSprint{
				"a": {createErr: task.ErrTaskCompleted},
				"b": {completed: true},
				"c": {completed: true},
			},
			wantStarted:       []string{"b", "c"},
			wantNotifications: []string{"step 3/3 c 10", "completed c work 10"},
			wantChainFinished: true,
		},
		{
			name: "candidate without remaining time is never announced",
			sprints: map[string]fakeSprint{
				"a": {completed: true},
				"b": {createErr: task.ErrTaskCompleted},
				"c": {completed: true},
			},
			wantStarted:       []string{"a", "c"},
			wantNotifications: []string{"step 3/3 c 10", "completed c work 10"},
			wantChainFinished: true,
		},
		{
			name: "sprint stopped early: no advance for the next step",
			sprints: map[string]fakeSprint{
				"a": {completed: false},
				"b": {completed: true},
				"c": {completed: false},
			},
			wantStarted:       []string{"a", "b", "c"},
			wantNotifications: []string{"step 3/3 c 10"},
			wantChainFinished: true,
		},
		{
			name: "completed sprint followed only by skipped candidates gets the completion banner",
			sprints: map[string]fakeSprint{
				"a": {completed: true},
				"b": {createErr: task.ErrTaskCompleted},
				"c": {createErr: task.ErrTaskCompleted},
			},
			wantStarted:       []string{"a"},
			wantNotifications: []string{"completed a work 10"},
			wantChainFinished: true,
		},
		{
			name: "abort stops chaining without error",
			sprints: map[string]fakeSprint{
				"a": {completed: true},
				"b": {completed: false, runErr: task.ErrTaskAborted},
				"c": {completed: true},
			},
			wantStarted:       []string{"a", "b"},
			wantNotifications: []string{"step 2/3 b 10"},
		},
		{
			name: "creation error flushes the pending completion and is returned",
			sprints: map[string]fakeSprint{
				"a": {completed: true},
				"b": {createErr: errBackend},
				"c": {completed: true},
			},
			wantErr:           errBackend,
			wantStarted:       []string{"a"},
			wantNotifications: []string{"completed a work 10"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notifications, started := fakeCombo(t, tc.sprints)
			out := &bytes.Buffer{}

			err := RunCombo(out, comboCandidates("a", "b", "c"), 10)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("RunCombo error = %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(*started, tc.wantStarted) {
				t.Errorf("started sprints = %v, want %v", *started, tc.wantStarted)
			}
			if !reflect.DeepEqual(*notifications, tc.wantNotifications) {
				t.Errorf("notifications = %v, want %v", *notifications, tc.wantNotifications)
			}
			if got := strings.Contains(out.String(), "Evening Combo Chain completed"); got != tc.wantChainFinished {
				t.Errorf("chain finished message present = %v, want %v\noutput:\n%s", got, tc.wantChainFinished, out.String())
			}
		})
	}
}

func TestBuildBatch(t *testing.T) {
	cands := comboCandidates("games", "home_task", "telegram", "video", "movies")
	names := func(cs []entity.EveningFocusCandidate) []string {
		out := make([]string, len(cs))
		for i, c := range cs {
			out[i] = c.TaskName
		}
		return out
	}
	cases := []struct {
		anchor string
		n      int
		want   []string
	}{
		{"", 3, []string{"games", "home_task", "telegram"}},
		{"telegram", 3, []string{"telegram", "games", "home_task"}},
		{"Home_Task", 4, []string{"home_task", "games", "telegram", "video"}},
		{"reading", 3, []string{"reading", "games", "home_task"}}, // anchor outside the candidates
	}
	for _, c := range cases {
		if got := names(BuildBatch(cands, c.anchor, c.n)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("BuildBatch(%q, %d) = %v, want %v", c.anchor, c.n, got, c.want)
		}
	}
}

func TestRunBatchSendsOriginAndDebtShare(t *testing.T) {
	_, started := fakeCombo(t, map[string]fakeSprint{"games": {completed: true}, "video": {completed: true}})
	var out bytes.Buffer
	if err := RunBatch(&out, comboCandidates("games", "video"), 15, 25); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*started, []string{"games", "video"}) {
		t.Errorf("started %v", *started)
	}
	if lastOptions.origin != "batch" || lastOptions.debtShare != 25 {
		t.Errorf("expected origin batch and share 25, got %+v", *lastOptions)
	}
	if !strings.Contains(out.String(), "[Батч 1/2]") || !strings.Contains(out.String(), "Батч завершён") {
		t.Errorf("unexpected output:\n%s", out.String())
	}
}
