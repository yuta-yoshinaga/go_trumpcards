//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// BalootPlayer バルートのプレイヤー
type BalootPlayer struct {
	*GamePlayer
	TrickHolder
	// hasBaloot は Hokom で切り札の K+Q を持っているか（Baloot、20点）。
	hasBaloot bool
	// balootRevealed は Baloot の有無を対戦相手に見せてよいか。
	//
	// **配られた瞬間に相手の手の内が割れるのはこのゲームの体験を壊す** (#5750)。
	// 実際に切り札の K か Q を出した時点（またはラウンド終了）で初めて開く。
	balootRevealed bool
	// declared はモード宣言を済ませたか。
	declared bool
}

// NewBalootPlayer コンストラクタ
func NewBalootPlayer(isHuman bool) *BalootPlayer {
	return &BalootPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする
func (p *BalootPlayer) ResetGame() { p.ResetRound() }

// ResetRound 1 ラウンド分の状態を初期化する
func (p *BalootPlayer) ResetRound() {
	resetPlayerRound(p)
	p.hasBaloot = false
	p.balootRevealed = false
	p.declared = false
}

// GetHasBaloot Baloot（切り札の K+Q）を持っているか
func (p *BalootPlayer) GetHasBaloot() bool { return p.hasBaloot }

// SetHasBaloot Baloot の有無を設定する
func (p *BalootPlayer) SetHasBaloot(b bool) { p.hasBaloot = b }

// GetBalootRevealed Baloot の有無を対戦相手に見せてよいか
func (p *BalootPlayer) GetBalootRevealed() bool { return p.balootRevealed }

// SetBalootRevealed Baloot の開示状態を設定する
func (p *BalootPlayer) SetBalootRevealed(b bool) { p.balootRevealed = b }

// GetDeclared モード宣言を済ませたか
func (p *BalootPlayer) GetDeclared() bool { return p.declared }

// SetDeclared 宣言済みフラグを設定する
func (p *BalootPlayer) SetDeclared(b bool) { p.declared = b }

// balootPlayerJSON is the JSON wire format for BalootPlayer.
type balootPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// Baloot と宣言状態は往復させる。Worker はリクエストごとに KV から
	// 作り直すので、抜けるとボーナスが消えたり宣言がやり直しになる (#4478)。
	HasBaloot bool `json:"hb"`
	// BalootRevealed は Baloot を対戦相手に見せてよいか。
	BalootRevealed bool `json:"br"`
	Declared       bool `json:"dc"`
}

// MarshalJSON implements json.Marshaler.
func (p *BalootPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(balootPlayerJSON{
		GamePlayer:     p.GamePlayer,
		TrickHolder:    &p.TrickHolder,
		HasBaloot:      p.hasBaloot,
		BalootRevealed: p.balootRevealed,
		Declared:       p.declared,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BalootPlayer) UnmarshalJSON(data []byte) error {
	var j balootPlayerJSON
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
	p.hasBaloot = j.HasBaloot
	p.balootRevealed = j.BalootRevealed
	p.declared = j.Declared
	return nil
}
