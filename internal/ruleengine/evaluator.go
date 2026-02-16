package ruleengine

import (
	"bytes"
	"encoding/json"
	"github.com/diegoholiveira/jsonlogic/v3"
)

// Evaluate checks if a user's metrics match a segment's JSON rule.
func Evaluate(rule json.RawMessage, data interface{}) (bool, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	var result bytes.Buffer
	err = jsonlogic.Apply(bytes.NewReader(rule), bytes.NewReader(dataBytes), &result)
	if err != nil {
		return false, err
	}

	// The library returns "true" or "false" as a string in the buffer
	return result.String() == "true", nil
}