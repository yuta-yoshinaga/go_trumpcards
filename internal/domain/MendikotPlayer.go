//go:build !js || !wasm || extra

package domain

import "encoding/json"

// MendikotPlayer メンディコットのプレイヤー
type MendikotPlayer struct {
	*GamePlayer
	TrickHolder
	// tens はこのハンドで獲得した 10 の枚数。**勝敗はこれで決まる。**
	tens int
}

// NewMendikotPlayer コンストラクタ
func NewMendikotPlayer(isHuman bool) *MendikotPlayer {
	return &MendikotPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする
func (p *MendikotPlayer) ResetGame() { p.ResetRound() }

// ResetRound 1 ハンド分の状態を初期化する
func (p *MendikotPlayer) ResetRound() {
	resetPlayerRound(p)
	p.tens = 0
}

// GetTens 獲得した 10 の枚数
func (p *MendikotPlayer) GetTens() int { return p.tens }

// AddTens 獲得した 10 を数える
func (p *MendikotPlayer) AddTens(n int) { p.tens += n }

// SetTens 獲得した 10 の枚数を設定する（復元・テスト用）
func (p *MendikotPlayer) SetTens(n int) { p.tens = n }

// mendikotPlayerJSON is the JSON wire format for MendikotPlayer.
type mendikotPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// **10 の枚数は往復させる。** 勝敗そのものなので、抜けると勝ったハンドが
	// 勝ちでなくなる (#4478)。トリック数も 2-2 のときの決着に要る。
	Tens int `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (p *MendikotPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(mendikotPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Tens:        p.tens,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MendikotPlayer) UnmarshalJSON(data []byte) error {
	var j mendikotPlayerJSON
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
	p.tens = j.Tens
	return nil
}
