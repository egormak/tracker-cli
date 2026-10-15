// Package debt renders this week's debt (the previous days' unfinished schedule).
package debt

import (
	"fmt"
	"strings"

	"tracker_cli/internal/domain/entity"
)

var dayShort = map[string]string{
	"monday": "Пн", "tuesday": "Вт", "wednesday": "Ср", "thursday": "Чт",
	"friday": "Пт", "saturday": "Сб", "sunday": "Вс",
}

// DayShort returns the Russian two-letter name of a lowercase weekday.
func DayShort(day string) string {
	if s, ok := dayShort[strings.ToLower(day)]; ok {
		return s
	}
	return day
}

// Line is the one-line summary, e.g. "Долг: 126м · 4 позиции · старейший: Вт".
func Line(d entity.DebtSummary) string {
	if d.TotalMinutes <= 0 {
		return "Долгов нет ✅"
	}
	line := fmt.Sprintf("Долг: %dм · %d %s", d.TotalMinutes, len(d.Items), positions(len(d.Items)))
	if d.OldestDay != "" {
		line += " · старейший: " + DayShort(d.OldestDay)
	}
	return line
}

func positions(n int) string {
	switch {
	case n%10 == 1 && n%100 != 11:
		return "позиция"
	case n%10 >= 2 && n%10 <= 4 && (n%100 < 12 || n%100 > 14):
		return "позиции"
	default:
		return "позиций"
	}
}

var statusMark = map[string]string{
	"closed": "✅", "partial": "◐", "open": "○", "today": "●", "future": "·", "none": "—",
}

// WeekStrip renders the week, e.g. "Пн ✅  Вт ◐  Ср ○  Чт ●  Пт ·  Сб ·  Вс ·".
func WeekStrip(d entity.DebtSummary) string {
	cells := make([]string, 0, len(d.Days))
	for _, day := range d.Days {
		mark := statusMark[day.Status]
		if mark == "" {
			mark = "?"
		}
		cells = append(cells, DayShort(day.Day)+" "+mark)
	}
	return strings.Join(cells, "  ")
}

// Render is the full view for `tracker debt`: the line, the week strip and the items.
func Render(d entity.DebtSummary) string {
	var b strings.Builder
	b.WriteString(Line(d) + "\n")
	if len(d.Days) > 0 {
		b.WriteString(WeekStrip(d) + "\n")
	}
	for _, it := range d.Items {
		b.WriteString(fmt.Sprintf("  %s  %-14s %3dм\n", DayShort(it.Day), it.TaskName, it.Remaining))
	}
	if d.TodayRemaining > 0 {
		b.WriteString(fmt.Sprintf("Сегодня осталось: %dм\n", d.TodayRemaining))
	}
	return b.String()
}
