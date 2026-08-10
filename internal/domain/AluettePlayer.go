//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// AluettePlayer アリュエットのプレイヤークラス。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。得点はチーム単位でゲーム本体が管理する。
type AluettePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewAluettePlayer コンストラクタ
func NewAluettePlayer(isHuman bool) *AluettePlayer {
	return &AluettePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *AluettePlayer) ResetRound() {
	resetPlayerRound(p)
}

// aluettePlayerJSON is the JSON wire format for AluettePlayer.
type aluettePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *AluettePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(aluettePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *AluettePlayer) UnmarshalJSON(data []byte) error {
	var j aluettePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	return nil
}
