//go:build !js || !wasm || solo

package domain

import "encoding/json"

// MusPlayer ムスのプレイヤークラス。手札（GamePlayer）のみを保持する。点（アマ）は
// チーム単位で Mus 本体が管理するため、プレイヤー個別のスコアホルダーは持たない。
type MusPlayer struct {
	*GamePlayer
}

// NewMusPlayer コンストラクタ
func NewMusPlayer(isHuman bool) *MusPlayer {
	return &MusPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（手札・終了状態を初期化）
func (p *MusPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// musPlayerJSON is the JSON wire format for MusPlayer.
type musPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *MusPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(musPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MusPlayer) UnmarshalJSON(data []byte) error {
	var j musPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	return nil
}
