//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// QuadrillePlayer カドリール (Quadrille) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。点はゲーム本体が管理する。
type QuadrillePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewQuadrillePlayer コンストラクタ
func NewQuadrillePlayer(isHuman bool) *QuadrillePlayer {
	return &QuadrillePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *QuadrillePlayer) ResetRound() {
	resetPlayerRound(p)
}

// quadrillePlayerJSON is the JSON wire format for QuadrillePlayer.
type quadrillePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *QuadrillePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(quadrillePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *QuadrillePlayer) UnmarshalJSON(data []byte) error {
	var j quadrillePlayerJSON
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
