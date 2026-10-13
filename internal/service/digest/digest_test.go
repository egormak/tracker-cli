package digest

import (
	"strings"
	"testing"
	"time"
	"tracker_cli/internal/domain/entity"
)

func sampleDigest() entity.WeeklyDigest {
	busiest := entity.WeeklyDigestDay{Day: "monday", Date: "5 October 2026", TotalDone: 150}
	return entity.WeeklyDigest{
		Week:          WeekCurrent,
		WeekStartDate: "5 October 2026",
		WeekEndDate:   "11 October 2026",
		TotalPlanned:  310,
		TotalDone:     320,
		Roles: []entity.WeeklyDigestRole{
			{Role: "work", Planned: 200, Done: 200, SharePercent: 62},
			{Role: "rest", Planned: 70, Done: 120, SharePercent: 37},
		},
		Strict: []entity.WeeklyDigestTask{
			{Name: "work", Role: "work", Planned: 200, Done: 200, PlanPercent: 100, SharePercent: 100},
		},
		Free: []entity.WeeklyDigestTask{
			{Name: "video", Role: "rest", Planned: 10, Done: 90, PlanPercent: 900, SharePercent: 75},
		},
		FreeDone:    120,
		OverPlan:    []entity.WeeklyDigestDelta{{Name: "video", Planned: 10, Done: 90, Delta: 80}},
		Untouched:   []entity.WeeklyDigestDelta{{Name: "games", Planned: 60}},
		Unscheduled: []entity.WeeklyDigestDelta{{Name: "telegram", Done: 30}},
		Days: []entity.WeeklyDigestDay{
			busiest,
			{Day: "tuesday", Date: "6 October 2026", TotalDone: 120},
		},
		BusiestDay: &busiest,
	}
}

func TestRenderPlainText(t *testing.T) {
	out := Render(sampleDigest(), false)

	for _, want := range []string{
		"Итоги недели: 5 October 2026 — 11 October 2026\n",
		"Всего: 5ч 20м из 5ч 10м по плану (103%)\n",
		"\nРоли:\n",
		"  work: 3ч 20м из 3ч 20м · 62% всего времени\n",
		"\nСтрогие задачи:\n  work: 3ч 20м из 3ч 20м (100%)\n",
		"\nСвободное время (2ч):\n  video: 1ч 30м · 75% свободного времени · план 10м\n",
		"  сверх плана — video: +1ч 20м (1ч 30м из 10м)\n",
		"  не начато — games: план 1ч\n",
		"  без расписания — telegram: 30м\n",
		"  monday: 2ч 30м ← самый нагруженный\n",
		"  tuesday: 2ч\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "#") {
		t.Errorf("plain text must not contain markdown headings, got:\n%s", out)
	}
}

func TestRenderMarkdown(t *testing.T) {
	out := Render(sampleDigest(), true)

	for _, want := range []string{
		"# Итоги недели: 5 October 2026 — 11 October 2026\n",
		"\n## Роли\n\n- work: ",
		"\n## Выбор против расписания\n\n- сверх плана — video: ",
		"- monday: 2ч 30м ← самый нагруженный\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected markdown to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRenderEmptyWeek(t *testing.T) {
	out := Render(entity.WeeklyDigest{WeekStartDate: "12 October 2026", WeekEndDate: "18 October 2026"}, false)
	if !strings.Contains(out, "За эту неделю записей нет.") {
		t.Errorf("expected the empty-week notice, got:\n%s", out)
	}
	if strings.Contains(out, "Роли") {
		t.Errorf("empty week must not render sections, got:\n%s", out)
	}
}

func TestRenderSkipsEmptySections(t *testing.T) {
	d := sampleDigest()
	d.OverPlan, d.Untouched, d.Unscheduled, d.Strict = nil, nil, nil, nil
	out := Render(d, false)
	if strings.Contains(out, "Выбор против расписания") || strings.Contains(out, "Строгие задачи") {
		t.Errorf("empty sections must be skipped, got:\n%s", out)
	}
}

func TestWeekBeforeClean(t *testing.T) {
	monday := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	if got := WeekBeforeClean(monday); got != WeekPrevious {
		t.Errorf("Monday: expected %q, got %q", WeekPrevious, got)
	}
	sunday := time.Date(2026, 10, 11, 22, 0, 0, 0, time.UTC)
	if got := WeekBeforeClean(sunday); got != WeekCurrent {
		t.Errorf("Sunday: expected %q, got %q", WeekCurrent, got)
	}
}
