//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// BaccaratBanquePlayer はバカラ・バンクの席。
//
// 席は 3 つ ── バンカー (人間) と、左右 2 つの子 (タブロー)。**左右は別勘定**
// なので、それぞれ自分の札と賭けを持つ。
type BaccaratBanquePlayer struct {
	*GamePlayer
	// chips は持ちチップ。
	chips int
	// bet はこのクー (1 回の勝負) で張った額。子だけが使う。
	bet int
	// drawn は 3 枚目を引いたか。
	drawn bool
}

// NewBaccaratBanquePlayer constructs a BaccaratBanquePlayer.
func NewBaccaratBanquePlayer(isHuman bool, chips int) *BaccaratBanquePlayer {
	return &BaccaratBanquePlayer{GamePlayer: NewGamePlayer(isHuman), chips: chips}
}

// GetHand は手札を並べて返す。
func (p *BaccaratBanquePlayer) GetHand() []*Card {
	n := p.GetCardsSize()
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

// GetTotal はこの手の合計を返す。
func (p *BaccaratBanquePlayer) GetTotal() int { return BaccaratBanqueTotal(p.GetHand()) }

// GetChips は持ちチップを返す。
func (p *BaccaratBanquePlayer) GetChips() int { return p.chips }

// AddChips はチップを加算する (負で減算)。
func (p *BaccaratBanquePlayer) AddChips(n int) { p.chips += n }

// SetChips は持ちチップを設定する。
func (p *BaccaratBanquePlayer) SetChips(n int) { p.chips = n }

// GetBet はこのクーで張った額を返す。
func (p *BaccaratBanquePlayer) GetBet() int { return p.bet }

// SetBet はこのクーで張った額を設定する。
func (p *BaccaratBanquePlayer) SetBet(n int) { p.bet = n }

// HasDrawn は 3 枚目を引いたかを返す。
func (p *BaccaratBanquePlayer) HasDrawn() bool { return p.drawn }

// SetDrawn は 3 枚目を引いたかを設定する。
func (p *BaccaratBanquePlayer) SetDrawn(v bool) { p.drawn = v }

// ResetCoup は 1 クー分の状態をクリアする。**チップは維持する。**
func (p *BaccaratBanquePlayer) ResetCoup() {
	p.Reset()
	p.bet = 0
	p.drawn = false
	p.SetIsFinished(false)
}

// baccaratBanquePlayerJSON is the JSON wire format for BaccaratBanquePlayer.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** チップが
// 消えると、復元した盤で誰も破産しなくなる。
type baccaratBanquePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Bet        int         `json:"bt"`
	Drawn      bool        `json:"dr"`
}

// MarshalJSON implements json.Marshaler.
func (p *BaccaratBanquePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(baccaratBanquePlayerJSON{
		GamePlayer: p.GamePlayer, Chips: p.chips, Bet: p.bet, Drawn: p.drawn,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BaccaratBanquePlayer) UnmarshalJSON(data []byte) error {
	var j baccaratBanquePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.chips, p.bet, p.drawn = j.Chips, j.Bet, j.Drawn
	return nil
}
