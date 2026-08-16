package color

import "github.com/charmbracelet/lipgloss"

var (
	Red    = New("1")
	Green  = New("2")
	Yellow = New("3")
	Blue   = New("4")
	Purple = New("5")
	Cyan   = New("6")
)

var (
	HiRed    = New("9")
	HiPurple = New("13")
)

func New(color string) lipgloss.Color {
	return lipgloss.Color(color)
}
