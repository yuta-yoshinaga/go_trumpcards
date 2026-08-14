//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// RamsPlayer ラムスのプレイヤー
type RamsPlayer struct {
	*GamePlayer
	TrickHolder
	// chips は持ちチップ。**多いほど良い。**
	chips int
	// inRound はこのラウンドに参加しているか (play を選んだか)。
	inRound bool
	// decided は play / pass を選び終えたか。
	decided bool
	// roundTricks はこのラウンドで取ったトリック数。
	roundTricks int
}

// NewRamsPlayer コンストラクタ
func NewRamsPlayer(isHuman bool) *RamsPlayer {
	return &RamsPlayer{GamePlayer: NewGamePlayer(isHuman), chips: RamsStartingChips}
}

// ResetGame ゲーム全体をリセットする（チップも初期値に戻す）
func (p *RamsPlayer) ResetGame() {
	p.ResetRound()
	p.chips = RamsStartingChips
}

// ResetRound 1 ラウンド分の状態だけを初期化する。チップは残す。
func (p *RamsPlayer) ResetRound() {
	resetPlayerRound(p)
	p.inRound = false
	p.decided = false
	p.roundTricks = 0
}

// GetChips 持ちチップを返す
func (p *RamsPlayer) GetChips() int { return p.chips }

// AddChips チップを増減する（支払いは負の値）
func (p *RamsPlayer) AddChips(n int) { p.chips += n }

// SetChips チップを設定する（復元・テスト用）
func (p *RamsPlayer) SetChips(n int) { p.chips = n }

// GetInRound このラウンドに参加しているか
func (p *RamsPlayer) GetInRound() bool { return p.inRound }

// SetInRound 参加状態を設定する
func (p *RamsPlayer) SetInRound(b bool) { p.inRound = b }

// GetDecided play / pass を選び終えたか
func (p *RamsPlayer) GetDecided() bool { return p.decided }

// SetDecided 選択済みフラグを設定する
func (p *RamsPlayer) SetDecided(b bool) { p.decided = b }

// GetRoundTricks このラウンドで取ったトリック数
func (p *RamsPlayer) GetRoundTricks() int { return p.roundTricks }

// AddRoundTrick このラウンドのトリック獲得を 1 つ数える
func (p *RamsPlayer) AddRoundTrick() { p.roundTricks++ }

// SetRoundTricks このラウンドのトリック数を設定する（テスト用）
func (p *RamsPlayer) SetRoundTricks(n int) { p.roundTricks = n }

// ramsPlayerJSON is the JSON wire format for RamsPlayer.
type ramsPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// チップと参加状態は必ず往復させる。Worker はリクエストごとに KV から
	// 作り直すので、抜けると降りたはずの人が参加者に戻る (#4478)。
	Chips       int  `json:"ch"`
	InRound     bool `json:"ir"`
	Decided     bool `json:"dc"`
	RoundTricks int  `json:"rt"`
}

// MarshalJSON implements json.Marshaler.
func (p *RamsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ramsPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Chips:       p.chips,
		InRound:     p.inRound,
		Decided:     p.decided,
		RoundTricks: p.roundTricks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RamsPlayer) UnmarshalJSON(data []byte) error {
	var j ramsPlayerJSON
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
	p.chips = j.Chips
	p.inRound = j.InRound
	p.decided = j.Decided
	p.roundTricks = j.RoundTricks
	return nil
}
