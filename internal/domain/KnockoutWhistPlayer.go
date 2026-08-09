//go:build !js || !wasm || classic

package domain

import "encoding/json"

// KnockoutWhistPlayer ノックアウト・ホイストのプレイヤークラス。手札（GamePlayer）と
// 獲得トリック（TrickHolder）に加え、脱落状態・Dogbone 残数・現ラウンドのトリック数を
// 保持する。
type KnockoutWhistPlayer struct {
	*GamePlayer
	TrickHolder
	eliminated  bool
	dogbones    int
	roundTricks int
}

// NewKnockoutWhistPlayer コンストラクタ
func NewKnockoutWhistPlayer(isHuman bool) *KnockoutWhistPlayer {
	return &KnockoutWhistPlayer{GamePlayer: NewGamePlayer(isHuman), dogbones: KnockoutWhistStartingDogbones}
}

// ResetRound ラウンドをリセット（トリック・手札・ラウンドトリック数を初期化）。
// 脱落状態と Dogbone 残数はラウンドをまたいで保持する。
func (p *KnockoutWhistPlayer) ResetRound() {
	resetPlayerRound(p)
	p.roundTricks = 0
}

// GetEliminated 脱落しているか。
func (p *KnockoutWhistPlayer) GetEliminated() bool { return p.eliminated }

// SetEliminated 脱落状態を設定する。
func (p *KnockoutWhistPlayer) SetEliminated(v bool) { p.eliminated = v }

// GetDogbones Dogbone 残数を返す。
func (p *KnockoutWhistPlayer) GetDogbones() int { return p.dogbones }

// SetDogbones Dogbone 残数を設定する。
func (p *KnockoutWhistPlayer) SetDogbones(n int) { p.dogbones = n }

// GetRoundTricks 現ラウンドの獲得トリック数を返す。
func (p *KnockoutWhistPlayer) GetRoundTricks() int { return p.roundTricks }

// IncRoundTricks 現ラウンドの獲得トリック数を 1 増やす。
func (p *KnockoutWhistPlayer) IncRoundTricks() { p.roundTricks++ }

// knockoutWhistPlayerJSON is the JSON wire format for KnockoutWhistPlayer.
type knockoutWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Eliminated  bool         `json:"el"`
	Dogbones    int          `json:"db"`
	RoundTricks int          `json:"rt"`
}

// MarshalJSON implements json.Marshaler.
func (p *KnockoutWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(knockoutWhistPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Eliminated:  p.eliminated,
		Dogbones:    p.dogbones,
		RoundTricks: p.roundTricks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KnockoutWhistPlayer) UnmarshalJSON(data []byte) error {
	var j knockoutWhistPlayerJSON
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
	p.eliminated = j.Eliminated
	p.dogbones = j.Dogbones
	p.roundTricks = j.RoundTricks
	return nil
}
