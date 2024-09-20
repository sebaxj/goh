package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sebaxj/goh/model"
)

func main() {
  p := tea.NewProgram(model.NewState())
  if _, err := p.Run(); err != nil {
    fmt.Printf("ERROR: %v", err)
    os.Exit(1)
  }
}
