package utils

import (
	"encoding/json"
	"io"
)

func ParseJSONBody(body io.Reader, obj interface{}) (bool, string) {
	data, err := io.ReadAll(body)
	if err != nil {
		return false, "Failed to read request body"
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return false, "Invalid JSON format"
	}

	return true, ""
}
