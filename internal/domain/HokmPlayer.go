//go:build !js || !wasm || classic

package domain

import "encoding/json"

// HokmPlayer ホクムのプレイヤー
type HokmPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewHokmPlayer コンストラクタ
func NewHokmPlayer(isHuman bool) *HokmPlayer {
	return &HokmPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする
func (p *HokmPlayer) ResetGame() { p.ResetRound() }

// ResetRound 1 ハンド分の状態を初期化する
func (p *HokmPlayer) ResetRound() { resetPlayerRound(p) }

// hokmPlayerJSON is the JSON wire format for HokmPlayer.
type hokmPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	// **獲得トリック数は往復させる。** 7 トリック先取の判定にも Kot の判定にも
	// 使うので、抜けると勝ったハンドが勝ちでなくなる (#4478)。
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *HokmPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(hokmPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *HokmPlayer) UnmarshalJSON(data []byte) error {
	var j hokmPlayerJSON
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
