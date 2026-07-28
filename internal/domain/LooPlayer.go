//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// LooPlayer はルー (Loo) のプレイヤー。基底の GamePlayer (手札) に加えて、現ディールで
// 獲得したトリック (TrickHolder)、参加 (play) したか、全ディール通算の累計チップを持つ。
type LooPlayer struct {
	*GamePlayer
	*TrickHolder
	playing bool // 現ディールで参加 (play) しているか
	chips   int  // 全ディール通算の累計チップ
}

// NewLooPlayer constructs a LooPlayer.
func NewLooPlayer(isHuman bool) *LooPlayer {
	return &LooPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
		playing:     false,
		chips:       0,
	}
}

// GetPlaying は現ディールで参加しているかを返す。
func (p *LooPlayer) GetPlaying() bool { return p.playing }

// SetPlaying は参加フラグを設定する。
func (p *LooPlayer) SetPlaying(v bool) { p.playing = v }

// GetChips は累計チップを返す (全ディール通算)。
func (p *LooPlayer) GetChips() int { return p.chips }

// AddChips はチップを加算する (負の値で減算)。
func (p *LooPlayer) AddChips(n int) { p.chips += n }

// ResetChips は累計チップを 0 に戻す (新規ゲーム開始時)。
func (p *LooPlayer) ResetChips() { p.chips = 0 }

// ResetDeal は 1 ディール分の状態 (手札・トリック・参加フラグ) をクリアする。
// 累計チップは維持する。
func (p *LooPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
	p.playing = false
}

// looPlayerJSON is the JSON wire format for LooPlayer.
type looPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Playing     bool         `json:"pl"`
	Chips       int          `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *LooPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(looPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: p.TrickHolder,
		Playing:     p.playing,
		Chips:       p.chips,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *LooPlayer) UnmarshalJSON(data []byte) error {
	var j looPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = j.TrickHolder
	} else {
		p.TrickHolder = &TrickHolder{}
	}
	p.playing = j.Playing
	p.chips = j.Chips
	return nil
}
