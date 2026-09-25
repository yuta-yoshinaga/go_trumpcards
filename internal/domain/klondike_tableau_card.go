//go:build !js || !wasm || solo || extra4

package domain

// KlondikeTableauCard is a face-up or face-down card in a Klondike-family tableau.
type KlondikeTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}
