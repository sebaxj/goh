package utils

import "github.com/charmbracelet/lipgloss"

var (
  PromptStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD770"))
  AnswerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#878787"))
  ButtonStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B6465F"))
  ResponseStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8FF")).Background(lipgloss.Color("#008080")).Padding(1, 1)
  ErrorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8FF")).Background(lipgloss.Color("#EC555E")).Padding(1, 1)
)
