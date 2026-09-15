package theme

import (
	"fmt"
	"strings"
	"tracker_cli/internal/pkg/restutil"

	"github.com/charmbracelet/lipgloss"
)

// TimeFlow Canvas palette
var (
	ColorWork       = lipgloss.Color("#FF6B4A") // Amber / Coral
	ColorWorkDark   = lipgloss.Color("#7C2D12")
	ColorLearn      = lipgloss.Color("#38BDF8") // Sky Blue
	ColorLearnDark  = lipgloss.Color("#0369A1")
	ColorRest       = lipgloss.Color("#34D399") // Emerald
	ColorRestDark   = lipgloss.Color("#065F46")
	ColorOther      = lipgloss.Color("#A78BFA") // Purple
	ColorBgCard     = lipgloss.Color("#0F172A") // Slate 900
	ColorBorder     = lipgloss.Color("#334155") // Slate 700
	ColorBorderLive = lipgloss.Color("#38BDF8")
	ColorMuted      = lipgloss.Color("#94A3B8") // Slate 400
	ColorText       = lipgloss.Color("#F8FAFC") // Slate 50
	ColorSuccess    = lipgloss.Color("#4ADE80")
	ColorWarning    = lipgloss.Color("#FBBF24")
	ColorDanger     = lipgloss.Color("#F87171")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			MarginBottom(1)

	SubTitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginBottom(1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBgCard).
			Padding(1, 2)

	ActiveCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWork).
			Background(ColorBgCard).
			Padding(1, 2)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorMuted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true).
			Foreground(ColorText).
			Background(lipgloss.Color("#1E293B")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorWork)
)

func RoleColor(role string) lipgloss.Color {
	switch strings.ToLower(role) {
	case "work":
		return ColorWork
	case "learn", "study", "english":
		return ColorLearn
	case "rest":
		return ColorRest
	default:
		return ColorOther
	}
}

func RoleBadge(role string) string {
	c := RoleColor(role)
	badgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(c).
		Background(lipgloss.Color("#1E293B")).
		Padding(0, 1).
		MarginRight(1)
	return badgeStyle.Render(strings.ToUpper(role))
}

func FormatRestTime(units int) string {
	minutes := restutil.MinutesFromUnits(units)
	return fmt.Sprintf("%.1f min", minutes)
}

func RenderProgressBar(width int, ratio float64, fillCol lipgloss.Color) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filledLen := int(float64(width) * ratio)
	emptyLen := width - filledLen

	filled := lipgloss.NewStyle().Foreground(fillCol).Render(strings.Repeat("█", filledLen))
	empty := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("░", emptyLen))
	return filled + empty
}
