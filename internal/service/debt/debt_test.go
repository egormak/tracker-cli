package debt

import (
	"strings"
	"testing"

	"tracker_cli/internal/domain/entity"
)

func TestLine(t *testing.T) {
	if got := Line(entity.DebtSummary{}); got != "Долгов нет ✅" {
		t.Errorf("empty debt: got %q", got)
	}
	d := entity.DebtSummary{TotalMinutes: 126, OldestDay: "tuesday", Items: make([]entity.DebtItem, 4)}
	if got, want := Line(d), "Долг: 126м · 4 позиции · старейший: Вт"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := Line(entity.DebtSummary{TotalMinutes: 5, Items: make([]entity.DebtItem, 5)}); !strings.Contains(got, "5 позиций") {
		t.Errorf("plural: got %q", got)
	}
}

func TestWeekStripAndRender(t *testing.T) {
	d := entity.DebtSummary{
		TotalMinutes: 36,
		OldestDay:    "wednesday",
		Items:        []entity.DebtItem{{TaskName: "home_task", Day: "wednesday", Remaining: 36}},
		Days: []entity.DebtDay{
			{Day: "monday", Status: "closed"}, {Day: "tuesday", Status: "partial"}, {Day: "wednesday", Status: "open"},
			{Day: "thursday", Status: "today"}, {Day: "friday", Status: "future"}, {Day: "saturday", Status: "none"},
		},
	}
	if got, want := WeekStrip(d), "Пн ✅  Вт ◐  Ср ○  Чт ●  Пт ·  Сб —"; got != want {
		t.Errorf("strip: got %q, want %q", got, want)
	}
	out := Render(d)
	for _, want := range []string{"Долг: 36м", "Пн ✅", "Ср  home_task", "36м"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
}
