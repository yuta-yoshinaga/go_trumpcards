package presenter

import "encoding/json"

// jsonMarshal is a variable wrapping json.Marshal to allow replacement in tests.
var jsonMarshal = json.Marshal

// marshalOrError marshals v to JSON, returning internalServerErrorJSON on failure.
func marshalOrError(v any) string {
	res, err := jsonMarshal(v)
	if err != nil {
		return internalServerErrorJSON()
	}
	return string(res)
}
