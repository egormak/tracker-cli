// Package digest renders the weekly digest for the terminal or as markdown.
package digest

import (
	"fmt"
	"strings"
	"time"
	"tracker_cli/internal/domain/entity"
)

const (
	WeekCurrent  = "current"
	WeekPrevious = "previous"
)

// WeekBeforeClean picks the week `tracker clean` is about to wipe: on Monday the
// finished week is already the previous one.
func WeekBeforeClean(now time.Time) string {
	if now.Weekday() == time.Monday {
		return WeekPrevious
	}
	return WeekCurrent
}

func minutes(m int) string {
	if m < 60 {
		return fmt.Sprintf("%dм", m)
	}
	if m%60 == 0 {
		return fmt.Sprintf("%dч", m/60)
	}
	return fmt.Sprintf("%dч %dм", m/60, m%60)
}

func percent(part, total int) int {
	if total <= 0 {
		return 0
	}
	return part * 100 / total
}

// Render formats the digest. With markdown it emits headings and bullet lists,
// otherwise an indented plain-text layout.
func Render(d entity.WeeklyDigest, markdown bool) string {
	var b strings.Builder

	title := func(text string) {
		if markdown {
			fmt.Fprintf(&b, "# %s\n\n", text)
		} else {
			fmt.Fprintf(&b, "%s\n", text)
		}
	}
	section := func(text string) {
		if markdown {
			fmt.Fprintf(&b, "\n## %s\n\n", text)
		} else {
			fmt.Fprintf(&b, "\n%s:\n", text)
		}
	}
	item := func(format string, args ...any) {
		prefix := "  "
		if markdown {
			prefix = "- "
		}
		fmt.Fprintf(&b, prefix+format+"\n", args...)
	}

	title(fmt.Sprintf("Итоги недели: %s — %s", d.WeekStartDate, d.WeekEndDate))
	if d.Source == "snapshot" {
		fmt.Fprintf(&b, "Снимок перед очисткой (%s).\n", snapshotTime(d.SnapshotTakenAt))
	}
	if d.TotalDone == 0 {
		b.WriteString("За эту неделю записей нет.\n")
		usageSection(d, section, item)
		return b.String()
	}
	fmt.Fprintf(&b, "Всего: %s из %s по плану (%d%%)\n",
		minutes(d.TotalDone), minutes(d.TotalPlanned), percent(d.TotalDone, d.TotalPlanned))

	section("Роли")
	for _, r := range d.Roles {
		item("%s: %s из %s · %d%% всего времени", r.Role, minutes(r.Done), minutes(r.Planned), r.SharePercent)
	}

	if len(d.Strict) > 0 {
		section("Строгие задачи")
		for _, t := range d.Strict {
			item("%s: %s из %s (%d%%)", t.Name, minutes(t.Done), minutes(t.Planned), t.PlanPercent)
		}
	}

	if len(d.Free) > 0 {
		section(fmt.Sprintf("Свободное время (%s)", minutes(d.FreeDone)))
		for _, t := range d.Free {
			item("%s: %s · %d%% свободного времени · план %s", t.Name, minutes(t.Done), t.SharePercent, minutes(t.Planned))
		}
	}

	if len(d.OverPlan)+len(d.Untouched)+len(d.Unscheduled) > 0 {
		section("Выбор против расписания")
		for _, t := range d.OverPlan {
			item("сверх плана — %s: +%s (%s из %s)", t.Name, minutes(t.Delta), minutes(t.Done), minutes(t.Planned))
		}
		for _, t := range d.Untouched {
			item("не начато — %s: план %s", t.Name, minutes(t.Planned))
		}
		for _, t := range d.Unscheduled {
			item("без расписания — %s: %s", t.Name, minutes(t.Done))
		}
	}

	section("Дни")
	for _, day := range d.Days {
		mark := ""
		if d.BusiestDay != nil && day.Day == d.BusiestDay.Day {
			mark = " ← самый нагруженный"
		}
		item("%s: %s%s", day.Day, minutes(day.TotalDone), mark)
	}

	if len(d.Clients) > 0 {
		section("Клиенты")
		for _, c := range d.Clients {
			item("%s: %s · %d%% времени · записей %d", clientLabel(c.Client), minutes(c.Minutes), c.SharePercent, c.Records)
		}
	}
	usageSection(d, section, item)

	return b.String()
}

// usageSection lists the week's choice counters with the per-client split.
func usageSection(d entity.WeeklyDigest, section func(string), item func(string, ...any)) {
	if len(d.Usage) == 0 {
		return
	}
	section("Использование функций")
	for _, u := range d.Usage {
		parts := make([]string, 0, len(u.Clients))
		for _, c := range u.Clients {
			parts = append(parts, fmt.Sprintf("%s %d", clientLabel(c.Client), c.Count))
		}
		item("%s: %d (%s)", usageLabel(u.Event), u.Count, strings.Join(parts, ", "))
	}
}

var usageLabels = map[string]string{
	"start_manual":        "запуск вручную",
	"start_break_variety": "перерыв — отстающая задача",
	"start_break_mood":    "перерыв — по настроению",
	"start_evening":       "запуск из вечернего режима",
	"start_combo":         "шаг комбо",
	"evening_skip":        "пропуск в вечернем режиме",
	"ramp_reset":          "сброс разгона",
	"timer_extend":        "продление таймера",
}

// usageLabel falls back to the raw event name for events a newer server may add.
func usageLabel(event string) string {
	if label, ok := usageLabels[event]; ok {
		return label
	}
	return event
}

func clientLabel(client string) string {
	if client == "unknown" {
		return "без метки"
	}
	return client
}

func snapshotTime(takenAt string) string {
	t, err := time.Parse(time.RFC3339, takenAt)
	if err != nil {
		return takenAt
	}
	return t.Local().Format("2 January 2006 15:04")
}
