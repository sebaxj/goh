package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
  options []string // request options
  cursor int // which option our cursor is pointing at

  url textinput.Model
  method string
  response string

  submenuVisible bool
  submenuType string
  listChoices list.Model
  loading bool
  spinner spinner.Model
}

type methodItem string

func (m methodItem) Title() string { return string(m) }
func (m methodItem) Description() string { return "" }
func (m methodItem) FilterValue() string { return string(m) }

var methods = []methodItem{"POST", "GET", "PUT", "DELETE"}

// styling
var (
  promptStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD770"))
  answerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
  buttonStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B6465F"))
  responseStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#008080")).Padding(1, 1)
)

func main() {
  p := tea.NewProgram(initModel())
  if _, err := p.Run(); err != nil {
    fmt.Printf("ERROR: %v", err)
    os.Exit(1)
  }
}

func initModel() model {
  // text input
  ti := textinput.New()
  ti.Placeholder = "Enter URL..."
  ti.Focus()

  // Convert the method strings to []list.Item using the methodItem type
  items := make([]list.Item, len(methods))
  for i, m := range methods {
    items[i] = m
  }

  // method select
  l := list.New(items, list.NewDefaultDelegate(), 25, 20)
  l.Title = "Select HTTP Method"

  // create spinner model
  s := spinner.New()
  s.Spinner = spinner.Line

  return model{
    options: []string{"URL", "Method", "", buttonStyle.Render("Submit Request")},
    cursor: 0,
    url: ti,
    submenuVisible: false,
    listChoices: l,
    spinner: s,
  }
}

func (m model) Init() tea.Cmd {
  return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
  case tea.KeyMsg:
    // this is a key press
    switch msg.String() {
    case "ctrl+c":
      // quit the program
      return m, tea.Quit

    case "up", "k":
      if !m.submenuVisible {
        if m.cursor > 0 {
          if len(m.options) - 1 == m.cursor {
            m.cursor -= 2
          } else {
            m.cursor--
          }
        }
      }

    case "down", "j":
      if !m.submenuVisible {
        if m.cursor < len(m.options) - 1 {
          if len(m.options) - 3 == m.cursor {
            m.cursor += 2
          } else {
            m.cursor++
          }
        }
      }

    case "enter":
      // handle selection from the main menu
      if m.submenuVisible {
        switch m.submenuType {
        case "url":
          m.options[0] = fmt.Sprintf("Enter URL:\t%s", m.url.Value())
        case "method":
          i, ok := m.listChoices.SelectedItem().(methodItem)
          if ok {
            m.method = string(i)
            m.options[1] = fmt.Sprintf("Select HTTP Method:\t%s", m.method)
          }
        }
        m.submenuVisible = false
      } else {
        // open submenu based on selection
        switch m.cursor {
        case 0:
          m.submenuType = "url"
          m.submenuVisible = true
        case 1:
          m.submenuType = "method"
          m.submenuVisible = true
        case 2:
          // do nothing since this is an empty line to separate the Submit button
        case 3:
          if !m.loading {
            m.loading = true
            return m, tea.Batch(m.spinner.Tick, m.makeHttpRequest())
          }
        }
      }
    }

    // update submenus if visible
    if m.submenuVisible {
      switch m.submenuType {
      case "url":
        var cmd tea.Cmd
        m.url, cmd = m.url.Update(msg)
        return m, cmd
      case "method":
        var cmd tea.Cmd
        m.listChoices, cmd = m.listChoices.Update(msg)
        return m, cmd
      }
    }

  case spinner.TickMsg:
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd

  case string:
    // handle the HTTP response
    m.loading = false
    m.response = msg
  }

  return m, nil
}

func (m model) View() string {
  if m.loading {
    return fmt.Sprintf("%s Submitting request...\n\n%s", m.spinner.View(), m.response)
  }

  if m.submenuVisible {
    switch m.submenuType {
    case "url":
      return fmt.Sprintf("%s \n\n%s", promptStyle.Render("URL:"), m.url.View())
    case "method":
      return m.listChoices.View()
    }
  }

  var b strings.Builder
  for i, option := range m.options {
    if i == m.cursor {
      fmt.Fprintf(&b, "→ %s\n", promptStyle.Render(option))
    } else {
      if i < 2 && strings.Contains(m.options[i], "\t") {
        fmt.Fprintf(&b, "  %s\n", answerStyle.Render(option))
      } else {
        fmt.Fprintf(&b, "  %s\n", option)
      }
    }
  }

  // display the HTTP response
  if m.response != "" {
    b.WriteString("\n" + responseStyle.Render(m.response) + "\n")
  }

  return b.String() + "\nPress CTRL-c to quit."
}

func (m model) makeHttpRequest() tea.Cmd {
  url := m.url.Value()
  method := m.method
  if "" == url {
    return func() tea.Msg {
      return "Error: URL is missing"
    }
  } else if "" == method {
    return func() tea.Msg {
      return "Error: Method is missing"
    }
  }

  // perform HTTP request
  return func() tea.Msg {
    client := &http.Client{}
    req, err := http.NewRequest(method, url, nil)
    if err != nil {
      return "Error creating request"
    }

    res, err := client.Do(req)
    if err != nil {
      return "Error executing request"
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
      return "Error reading response body"
    }

    // pretty print the JSON response
    var prettyJSON bytes.Buffer
    if err := json.Indent(&prettyJSON, body, "", " "); err != nil {
      return "Error formatting JSON response"
    }

    return prettyJSON.String()
  }
}
