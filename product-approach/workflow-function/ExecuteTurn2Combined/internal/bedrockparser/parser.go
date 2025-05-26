package bedrockparser

import (
	"encoding/json"

	"workflow-function/shared/schema"
)

// ParseTurn2Response attempts to decode the LLM response from Turn 2 into
// a structured FinalResults object. If the response is not valid JSON it is
// returned as a generic map for downstream handling.
func ParseTurn2Response(data string) (interface{}, error) {
	var results schema.FinalResults
	if err := json.Unmarshal([]byte(data), &results); err == nil {
		return &results, nil
	}

	var generic map[string]interface{}
	if err := json.Unmarshal([]byte(data), &generic); err != nil {
		return nil, err
	}
	return generic, nil
}
