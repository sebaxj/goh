package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func PrettifyJSON(data interface{}) (string, error) {
  var jsonData []byte
  var err error

  // determine the type of the input
  switch v := data.(type) {
  case string:
    jsonData = []byte(v)
  case []byte:
    jsonData = v
  default:
    return "", fmt.Errorf("unsupported  type: %T", data)
  }

  var prettyJSON bytes.Buffer
  if err = json.Indent(&prettyJSON, jsonData, "", " "); err != nil {
    return "", err
  }

  return prettyJSON.String(), nil
}
