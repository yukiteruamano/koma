package style

import "github.com/charmbracelet/lipgloss"

func New() lipgloss.Style {
	return lipgloss.NewStyle()
}

func NewColored(foreground, background lipgloss.Color) lipgloss.Style {
	return New().Foreground(foreground).Background(background)
}

func Fg(color lipgloss.Color) func(string) string {
	return func(s string) string { return NewColored(color, "").Render(s) }
}

func Truncate(max int) func(string) string {
	return func(s string) string { return New().Width(max).Render(s) }
}

func Faint(s string) string  { return New().Faint(true).Render(s) }
func Bold(s string) string   { return New().Bold(true).Render(s) }
func Italic(s string) string { return New().Italic(true).Render(s) }
