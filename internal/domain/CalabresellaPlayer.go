//go:build !js || !wasm || extra

package domain

import "encoding/json"

// CalabresellaPlayer カラブレセッラ (Calabresella) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。点はゲーム本体が管理する。
type CalabresellaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewCalabresellaPlayer コンストラクタ
func NewCalabresellaPlayer(isHuman bool) *CalabresellaPlayer {
	return &CalabresellaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *CalabresellaPlayer) ResetRound() {
	resetPlayerRound(p)
}

// calabresellaPlayerJSON is the JSON wire format for CalabresellaPlayer.
type calabresellaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *CalabresellaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(calabresellaPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CalabresellaPlayer) UnmarshalJSON(data []byte) error {
	var j calabresellaPlayerJSON
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
