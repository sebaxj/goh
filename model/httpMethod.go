package model

type methodItem string

var HttpMethods = []methodItem{
  "POST",
  "GET",
  "PUT",
  "DELETE",
}

func (m methodItem) Title() string {
  return string(m)
}

func (m methodItem) Description() string {
  return string("")
}

func (m methodItem) FilterValue() string {
  return string(m)
}
