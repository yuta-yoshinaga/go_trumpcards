//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ZwanzigerrufenPlayer ツヴァンツィガールーフェン (Zwanzigerrufen) のプレイヤー。
// 手札 (GamePlayer) と獲得トリック (TrickHolder) を保持する。得点はゲーム本体が管理する。
type ZwanzigerrufenPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewZwanzigerrufenPlayer コンストラクタ
func NewZwanzigerrufenPlayer(isHuman bool) *ZwanzigerrufenPlayer {
	return &ZwanzigerrufenPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *ZwanzigerrufenPlayer) ResetRound() {
	resetPlayerRound(p)
}

// zwanzigerrufenPlayerJSON is the JSON wire format for ZwanzigerrufenPlayer.
type zwanzigerrufenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *ZwanzigerrufenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(zwanzigerrufenPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ZwanzigerrufenPlayer) UnmarshalJSON(data []byte) error {
	var j zwanzigerrufenPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.TrickHolder, j.TrickHolder)
	return nil
}
