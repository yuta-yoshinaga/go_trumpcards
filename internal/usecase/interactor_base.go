package usecase

import "encoding/json"

// GameBase provides common functionality for all game interactors.
// Embed this struct to inherit the Snapshot method.
type GameBase[G any] struct {
	Game G
}

// Snapshot serialises the game state to JSON for KV persistence.
func (b *GameBase[G]) Snapshot() ([]byte, error) {
	return json.Marshal(b.Game)
}

// restoreAndBuild deserialises JSON into a domain struct and passes it to a
// builder function that constructs the interactor.
func restoreAndBuild[D any, I any](data []byte, build func(*D) *I) (*I, error) {
	g, err := restoreGame[D](data)
	if err != nil {
		return nil, err
	}
	return build(g), nil
}
