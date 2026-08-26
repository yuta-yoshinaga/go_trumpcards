//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// JulepePlayer フレペのプレイヤー
type JulepePlayer struct {
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

// NewJulepePlayer コンストラクタ
func NewJulepePlayer(isHuman bool) *JulepePlayer {
	return &JulepePlayer{GamePlayer: NewGamePlayer(isHuman), chips: JulepeStartingChips}
}

// ResetGame ゲーム全体をリセットする（チップも初期値に戻す）
func (p *JulepePlayer) ResetGame() {
	p.ResetRound()
	p.chips = JulepeStartingChips
}

// ResetRound 1 ラウンド分の状態だけを初期化する。チップは残す。
func (p *JulepePlayer) ResetRound() {
	resetPlayerRound(p)
	p.inRound = false
	p.decided = false
	p.roundTricks = 0
}

// GetChips 持ちチップを返す
func (p *JulepePlayer) GetChips() int { return p.chips }

// AddChips チップを増減する（支払いは負の値）
func (p *JulepePlayer) AddChips(n int) { p.chips += n }

// SetChips チップを設定する（復元・テスト用）
func (p *JulepePlayer) SetChips(n int) { p.chips = n }

// GetInRound このラウンドに参加しているか
func (p *JulepePlayer) GetInRound() bool { return p.inRound }

// SetInRound 参加状態を設定する
func (p *JulepePlayer) SetInRound(b bool) { p.inRound = b }

// GetDecided play / pass を選び終えたか
func (p *JulepePlayer) GetDecided() bool { return p.decided }

// SetDecided 選択済みフラグを設定する
func (p *JulepePlayer) SetDecided(b bool) { p.decided = b }

// GetRoundTricks このラウンドで取ったトリック数
func (p *JulepePlayer) GetRoundTricks() int { return p.roundTricks }

// AddRoundTrick このラウンドのトリック獲得を 1 つ数える
func (p *JulepePlayer) AddRoundTrick() { p.roundTricks++ }

// SetRoundTricks このラウンドのトリック数を設定する（テスト用）
func (p *JulepePlayer) SetRoundTricks(n int) { p.roundTricks = n }

// julepePlayerJSON is the JSON wire format for JulepePlayer.
type julepePlayerJSON struct {
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
func (p *JulepePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(julepePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Chips:       p.chips,
		InRound:     p.inRound,
		Decided:     p.decided,
		RoundTricks: p.roundTricks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *JulepePlayer) UnmarshalJSON(data []byte) error {
	var j julepePlayerJSON
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
