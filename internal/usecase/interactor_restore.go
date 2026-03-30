package usecase

import "encoding/json"

// restoreGame deserialises JSON into a game domain struct.
func restoreGame[G any](data []byte) (*G, error) {
	var g G
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}
