package model

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebaxj/goh/utils"
)

const JSON_ERROR_STRING string = "Error decoding JSON:\n\n"

type state struct {
  options []string // request options
  cursor int // which option our cursor is pointing at

  url textinput.Model // URL
  method string // HTTP method
  body textarea.Model // HTTP request body
  response string // HTTP response

  // additional state
  submenuVisible bool
  submenuType string
  listChoices list.Model
  loading bool
  spinner spinner.Model
  history []string
}

func NewState() *state {
  // text input
  ti := textinput.New()
  ti.Placeholder = "Enter URL..."
  ti.Focus()

  // Convert the method strings to []list.Item using the methodItem type
  items := make([]list.Item, len(HttpMethods))
  for i, m := range HttpMethods {
    items[i] = m
  }

  // method select
  l := list.New(items, list.NewDefaultDelegate(), 25, 20)
  l.Title = "Select HTTP Method"

  // JSON body text area
  jsonTa := textarea.New()
  jsonTa.SetWidth(50)
  jsonTa.SetHeight(5)
  jsonTa.SetValue("{\n}") // add braces by default

  // create spinner model
  s := spinner.New()
  s.Spinner = spinner.Line

  return &state{
    options: []string{"URL", "Method", "Body", "", utils.ButtonStyle.Render("Submit Request")},
    cursor: 0,
    url: ti,
    body: jsonTa,
    submenuVisible: false,
    listChoices: l,
    spinner: s,
  }
}

func (s *state) Init() tea.Cmd {
  return textinput.Blink
}

func (s *state) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {

  /*
   * Key Press
   */
  case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c":
      // quit the program
      return s, tea.Quit

    case "up", "k":
      if !s.submenuVisible {
        if s.cursor > 0 {
          if len(s.options) - 1 == s.cursor {
            s.cursor -= 2
          } else {
            s.cursor--
          }
        }
      }

    case "down", "j":
      if !s.submenuVisible {
        if s.cursor < len(s.options) - 1 {
          if len(s.options) - 3 == s.cursor {
            s.cursor += 2
          } else {
            s.cursor++
          }
        }
      }

    case "enter":
      if s.submenuVisible {
        // we are in a submenu

        switch s.submenuType {
        case "url":
          s.options[0] = fmt.Sprintf("Enter URL:\t%s", s.url.Value())
        case "method":
          i, ok := s.listChoices.SelectedItem().(methodItem)
          if ok {
            s.method = string(i)
            s.options[1] = fmt.Sprintf("Select HTTP Method:\t%s", s.method)
          }
        case "body":
          body := s.body.Value()
          if "{\n\n}" != body {
            prettyBody, err := utils.PrettifyJSON(body)
            if nil == err {
              s.body.SetValue(prettyBody)
            } else {
              s.body.SetValue(JSON_ERROR_STRING + body)
            }
            s.options[2] = fmt.Sprintf("Add JSON Body:\n\n%s", s.body.Value())
          } else {
            s.options[2] = fmt.Sprintf("Add JSON Body:")
          }
        }

        s.submenuVisible = false

      } else {
        // we are in the main menu, open submenu based on selection

        switch s.cursor {
        case 0:
          s.submenuType = "url"
          s.submenuVisible = true
        case 1:
          s.submenuType = "method"
          s.submenuVisible = true
        case 2:
          s.submenuType = "body"

          // check for error message
          body := s.body.Value()
          if strings.HasPrefix(body, JSON_ERROR_STRING) {
            pos := strings.Index(body, "{")
            if -1 != pos {
              s.body.SetValue(body[pos:])
            } else {
              s.body.SetValue("{\n}")
            }
          }

          s.submenuVisible = true
          s.body.Focus()
          s.body.CursorUp()
        case len(s.options) - 2:
          // do nothing since this is an empty line to separate the Submit button
        case len(s.options) - 1:
          if !s.loading {
            s.loading = true
            return s, tea.Batch(s.spinner.Tick, s.makeHttpRequest())
          }
        } // switch c.cursor
      }
    } // switch msg.String()

    // update submenus if visible
    if s.submenuVisible {
      switch s.submenuType {
      case "url":
        var cmd tea.Cmd
        s.url, cmd = s.url.Update(msg)
        return s, cmd
      case "method":
        var cmd tea.Cmd
        s.listChoices, cmd = s.listChoices.Update(msg)
        return s, cmd
      case "body":
        var cmd tea.Cmd
        s.body, cmd = s.body.Update(msg)
        return s, cmd
      }
    }

  case spinner.TickMsg:
    var cmd tea.Cmd
    s.spinner, cmd = s.spinner.Update(msg)
    return s, cmd

  case string:
    // handle the HTTP response
    s.loading = false
    s.response = msg

  }

  return s, nil
}

func (s *state) View() string {
  // we are executing a HTTP request
  if s.loading {
    return fmt.Sprintf("\n%s Submitting request...", s.spinner.View())
  }

  // we are in a submenu
  if s.submenuVisible {
    switch s.submenuType {
    case "url":
      return fmt.Sprintf("%s\n\n%s", utils.PromptStyle.Render("URL:"), s.url.View())
    case "method":
      return s.listChoices.View()
    case "body":
      return fmt.Sprintf("%s\n\n%s", utils.PromptStyle.Render("JSON Body:"), s.body.View())
    }
  }

  // print out the main view
  var b strings.Builder

  // history
  if len(s.history) > 0 {
    for _, req := range s.history {
      b.WriteString(req + "\n\n")
    }
  }

  // options
  for i, option := range s.options {
    if i == s.cursor {
      fmt.Fprintf(&b, "→ %s\n", utils.PromptStyle.Render(option))
    } else if 2 == i {
      body := s.body.Value()
      if strings.HasPrefix(body, JSON_ERROR_STRING) {
        fmt.Fprintf(&b, "  %s\n\n%s", utils.AnswerStyle.Render("Body"), utils.ErrorStyle.Render(body))
      } else {
        fmt.Fprintf(&b, "  %s\n", utils.AnswerStyle.Render(option))
      }
    } else {
      fmt.Fprintf(&b, "  %s\n", utils.AnswerStyle.Render(option))
    }
  }

  // response
  if s.response != "" {
    b.WriteString("\n" + utils.ErrorStyle.Render(s.response) + "\n")
  }

  return b.String() + "\nPress CTRL-c to quit."
}

func (s *state) makeHttpRequest() tea.Cmd {
  // parse input
  url := s.url.Value()
  method := s.method
  if "" == url {
    return func() tea.Msg {
      return "Error: URL is missing"
    }
  } else if "" == method {
    return func() tea.Msg {
      return "Error: Method is missing"
    }
  }

  // prepare the request body if provided
  var body *bytes.Reader = nil
  if "GET" != s.method {
    mBody := s.body.Value()
    if "" != mBody && "{\n}" != mBody {
      body = bytes.NewReader([]byte(mBody))
    }
  }

  // perform HTTP request
  return func() tea.Msg {
    client := &http.Client{}

    var req *http.Request
    var err error
    if nil == body {
      req, err = http.NewRequest(method, url, nil)
    } else {
      req, err = http.NewRequest(method, url, body)
    }

    if err != nil {
      return fmt.Sprintf("Error creating request: %v", err)
    }

    res, err := client.Do(req)
    if err != nil {
      return fmt.Sprintf("Error executing request: %v", err)
    }
    defer res.Body.Close()

    resBody, err := io.ReadAll(res.Body)
    if err != nil {
      return fmt.Sprintf("Error reading response body: %v", err)
    }

    // pretty print the JSON response
    prettyResBody, err := utils.PrettifyJSON(resBody)
    if err != nil {
      prettyResBody = string(resBody)
    }

    s.history = append(s.history, fmt.Sprintf("URL:\t%s\nMethod:\t%s\nBody:\n%s\nResponse:\n%s",
      utils.AnswerStyle.Render(url),
      utils.AnswerStyle.Render(method),
      utils.AnswerStyle.Render(s.body.Value()),
      utils.ResponseStyle.Render(prettyResBody)),
    )

    return ""
  }
}
