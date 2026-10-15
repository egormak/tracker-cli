package notifier

import (
	"bytes"
	"reflect"
	"testing"
)

type startedCommand struct {
	name string
	args []string
}

// fakeNotifier swaps the package seams so tests never spawn osascript/notify-send
// or write a bell to the real stdout. It returns the recorded commands and bell buffer.
func fakeNotifier(t *testing.T, os string) (*[]startedCommand, *bytes.Buffer) {
	t.Helper()

	origStarter, origGOOS, origBell := commandStarter, goos, bellWriter
	t.Cleanup(func() {
		commandStarter, goos, bellWriter = origStarter, origGOOS, origBell
	})

	var commands []startedCommand
	bell := &bytes.Buffer{}

	commandStarter = func(name string, args ...string) error {
		commands = append(commands, startedCommand{name: name, args: args})
		return nil
	}
	goos = os
	bellWriter = bell

	return &commands, bell
}

// darwinScript asserts exactly one osascript command was started and returns its script.
func darwinScript(t *testing.T, commands []startedCommand) string {
	t.Helper()

	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d: %+v", len(commands), commands)
	}
	cmd := commands[0]
	if cmd.name != "osascript" || len(cmd.args) != 2 || cmd.args[0] != "-e" {
		t.Fatalf("expected `osascript -e <script>`, got %s %q", cmd.name, cmd.args)
	}
	return cmd.args[1]
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "Review PRs", want: "Review PRs"},
		{name: "quotes", in: `say "hi"`, want: `say \"hi\"`},
		{name: "backslashes", in: `a\b`, want: `a\\b`},
		{name: "backslash before quote", in: `\"`, want: `\\\"`},
		{name: "apostrophes untouched", in: "it's", want: "it's"},
		{name: "unicode untouched", in: "⚡☕ задача", want: "⚡☕ задача"},
		{name: "empty", in: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeAppleScript(tc.in); got != tc.want {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNotifyWithSound_Darwin(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		message   string
		soundName string
		want      string
	}{
		{
			name:      "custom sound",
			title:     "Title",
			message:   "Message",
			soundName: "Submarine",
			want:      `display notification "Message" with title "Title" sound name "Submarine"`,
		},
		{
			name:      "empty sound falls back to default",
			title:     "Title",
			message:   "Message",
			soundName: "",
			want:      `display notification "Message" with title "Title" sound name "Glass"`,
		},
		{
			name:      "quotes and backslashes are escaped",
			title:     `Title with "quotes"`,
			message:   `Message with "quotes" and \backslashes\`,
			soundName: `Gl"ass`,
			want:      `display notification "Message with \"quotes\" and \\backslashes\\" with title "Title with \"quotes\"" sound name "Gl\"ass"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			commands, bell := fakeNotifier(t, "darwin")

			NotifyWithSound(tc.title, tc.message, tc.soundName)

			if got := darwinScript(t, *commands); got != tc.want {
				t.Errorf("script mismatch\n got: %s\nwant: %s", got, tc.want)
			}
			if bell.String() != "\a" {
				t.Errorf("expected a single terminal bell, got %q", bell.String())
			}
		})
	}
}

func TestNotify_UsesDefaultSound(t *testing.T) {
	commands, _ := fakeNotifier(t, "darwin")

	Notify("Title", "Message")

	want := `display notification "Message" with title "Title" sound name "Glass"`
	if got := darwinScript(t, *commands); got != want {
		t.Errorf("script mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestNotifyWithSound_Linux(t *testing.T) {
	commands, bell := fakeNotifier(t, "linux")

	NotifyWithSound(`Title "quoted"`, "Message", "Submarine")

	// notify-send receives raw arguments: no AppleScript escaping and no sound.
	want := []startedCommand{{name: "notify-send", args: []string{`Title "quoted"`, "Message"}}}
	if !reflect.DeepEqual(*commands, want) {
		t.Errorf("commands = %+v, want %+v", *commands, want)
	}
	if bell.String() != "\a" {
		t.Errorf("expected a single terminal bell, got %q", bell.String())
	}
}

func TestNotifyWithSound_UnsupportedOS(t *testing.T) {
	commands, bell := fakeNotifier(t, "windows")

	NotifyWithSound("Title", "Message", "Glass")

	if len(*commands) != 0 {
		t.Errorf("expected no command on unsupported OS, got %+v", *commands)
	}
	if bell.String() != "\a" {
		t.Errorf("expected a single terminal bell, got %q", bell.String())
	}
}

func TestNotifyTaskCompleted_RoleRouting(t *testing.T) {
	const (
		sprintScript = `display notification "Great job! Sprint 'Reading' (25 min) is finished. Time for a break." with title "⏱ Tracker — Task Completed!" sound name "Glass"`
		breakScript  = `display notification "Break 'Reading' (25 min) is finished. Time to get back into flow!" with title "☕ Tracker — Break Completed" sound name "Ping"`
	)

	tests := []struct {
		name string
		role string
		want string
	}{
		{name: "work role", role: "work", want: sprintScript},
		{name: "learn role", role: "learn", want: sprintScript},
		{name: "empty role", role: "", want: sprintScript},
		{name: "rest role", role: "rest", want: breakScript},
		{name: "rest role mixed case", role: "Rest", want: breakScript},
		{name: "rest role with whitespace", role: " REST\n", want: breakScript},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			commands, _ := fakeNotifier(t, "darwin")

			NotifyTaskCompleted("Reading", tc.role, 25)

			if got := darwinScript(t, *commands); got != tc.want {
				t.Errorf("script mismatch\n got: %s\nwant: %s", got, tc.want)
			}
		})
	}
}

func TestNotifyComboStep(t *testing.T) {
	commands, _ := fakeNotifier(t, "darwin")

	NotifyComboStep(2, 3, `Review "PRs"`, 15)

	want := `display notification "Step 2/3: Starting 'Review \"PRs\"' (15 min). Let's go!" with title "⚡ Tracker — Evening Combo Advance" sound name "Submarine"`
	if got := darwinScript(t, *commands); got != want {
		t.Errorf("script mismatch\n got: %s\nwant: %s", got, want)
	}
}
