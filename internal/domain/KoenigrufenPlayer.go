//go:build !js || !wasm || extra

package domain

import "encoding/json"

// KoenigrufenPlayer ケーニッヒルーフェン (Königrufen) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。得点はゲーム本体が管理する。
type KoenigrufenPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewKoenigrufenPlayer コンストラクタ
func NewKoenigrufenPlayer(isHuman bool) *KoenigrufenPlayer {
	return &KoenigrufenPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *KoenigrufenPlayer) ResetRound() {
	resetPlayerRound(p)
}

// koenigrufenPlayerJSON is the JSON wire format for KoenigrufenPlayer.
type koenigrufenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *KoenigrufenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(koenigrufenPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KoenigrufenPlayer) UnmarshalJSON(data []byte) error {
	var j koenigrufenPlayerJSON
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
