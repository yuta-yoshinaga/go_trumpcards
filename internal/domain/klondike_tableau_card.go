//go:build !js || !wasm || solo || extra4

package domain

// KlondikeTableauCard タブロー上のカード。Yukon 系ゲームでも共有する。
type KlondikeTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}
