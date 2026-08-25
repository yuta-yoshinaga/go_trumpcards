//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// SutdaPlayer はソッタのプレイヤー。
type SutdaPlayer struct {
	*GamePlayer
	// chips は所持チップ。
	chips int
	// bet はこのハンドで場に出したチップ。
	bet int
	// folded はこのハンドで降りたか。
	folded bool
	// revealed はショーダウンで手札を開いたか。
	revealed bool
}

// NewSutdaPlayer constructs a SutdaPlayer.
func NewSutdaPlayer(isHuman bool, chips int) *SutdaPlayer {
	return &SutdaPlayer{GamePlayer: NewGamePlayer(isHuman), chips: chips}
}

// GetChips は所持チップを返す。
func (p *SutdaPlayer) GetChips() int { return p.chips }

// SetChips は所持チップを設定する。
func (p *SutdaPlayer) SetChips(n int) { p.chips = n }

// AddChips はチップを加算する (負で減算)。
func (p *SutdaPlayer) AddChips(n int) { p.chips += n }

// GetBet はこのハンドで場に出したチップを返す。
func (p *SutdaPlayer) GetBet() int { return p.bet }

// AddBet は場に出したチップを加算する。
func (p *SutdaPlayer) AddBet(n int) { p.bet += n }

// ResetBet は場に出したチップを 0 に戻す。
func (p *SutdaPlayer) ResetBet() { p.bet = 0 }

// IsFolded は降りたかを返す。
func (p *SutdaPlayer) IsFolded() bool { return p.folded }

// SetFolded は降りたかを設定する。
func (p *SutdaPlayer) SetFolded(v bool) { p.folded = v }

// IsRevealed は手札を開いたかを返す。
func (p *SutdaPlayer) IsRevealed() bool { return p.revealed }

// SetRevealed は手札を開いたかを設定する。
func (p *SutdaPlayer) SetRevealed(v bool) { p.revealed = v }

// ResetHand は 1 ハンド分の状態をクリアする。チップは維持する。
func (p *SutdaPlayer) ResetHand() {
	p.Reset()
	p.bet = 0
	p.folded = false
	p.revealed = false
	p.SetIsFinished(false)
}

// sutdaPlayerJSON is the JSON wire format for SutdaPlayer.
type sutdaPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Bet        int         `json:"bt"`
	Folded     bool        `json:"fd"`
	Revealed   bool        `json:"rv"`
}

// MarshalJSON implements json.Marshaler.
func (p *SutdaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sutdaPlayerJSON{
		GamePlayer: p.GamePlayer,
		Chips:      p.chips,
		Bet:        p.bet,
		Folded:     p.folded,
		Revealed:   p.revealed,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SutdaPlayer) UnmarshalJSON(data []byte) error {
	var j sutdaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.chips = j.Chips
	p.bet = j.Bet
	p.folded = j.Folded
	p.revealed = j.Revealed
	return nil
}
