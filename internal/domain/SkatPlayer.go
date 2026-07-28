//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// SkatPlayer Skat player struct
type SkatPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid        int  // current accepted bid value (0 = passed, -1 = no bid yet)
	isDeclarer bool // declarer of this round
	cardPoints int  // card points won this round (0..120)
	roundsWon  int  // declarer wins this game session
	roundsLost int  // declarer losses this game session
}

// skatPlayerJSON is the JSON wire format for SkatPlayer.
type skatPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
	IsDeclarer       bool              `json:"id"`
	CardPoints       int               `json:"cp"`
	RoundsWon        int               `json:"rw"`
	RoundsLost       int               `json:"rl"`
}

// MarshalJSON implements json.Marshaler.
func (p *SkatPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(skatPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		IsDeclarer:       p.isDeclarer,
		CardPoints:       p.cardPoints,
		RoundsWon:        p.roundsWon,
		RoundsLost:       p.roundsLost,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SkatPlayer) UnmarshalJSON(data []byte) error {
	var j skatPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	p.isDeclarer = j.IsDeclarer
	p.cardPoints = j.CardPoints
	p.roundsWon = j.RoundsWon
	p.roundsLost = j.RoundsLost
	return nil
}

// NewSkatPlayer constructs a Skat player
func NewSkatPlayer(isHuman bool) *SkatPlayer {
	return &SkatPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid returns the player's current bid (-1 = none, 0 = passed)
func (p *SkatPlayer) GetBid() int { return p.bid }

// SetBid sets the player's bid
func (p *SkatPlayer) SetBid(bid int) { p.bid = bid }

// GetIsDeclarer returns whether the player is the declarer
func (p *SkatPlayer) GetIsDeclarer() bool { return p.isDeclarer }

// SetIsDeclarer sets the declarer flag
func (p *SkatPlayer) SetIsDeclarer(v bool) { p.isDeclarer = v }

// GetCardPoints returns card points won this round
func (p *SkatPlayer) GetCardPoints() int { return p.cardPoints }

// SetCardPoints sets card points
func (p *SkatPlayer) SetCardPoints(n int) { p.cardPoints = n }

// GetRoundsWon returns the declarer-win count
func (p *SkatPlayer) GetRoundsWon() int { return p.roundsWon }

// SetRoundsWon sets the declarer-win count
func (p *SkatPlayer) SetRoundsWon(n int) { p.roundsWon = n }

// IncRoundsWon increments declarer-win count
func (p *SkatPlayer) IncRoundsWon() { p.roundsWon++ }

// GetRoundsLost returns declarer-loss count
func (p *SkatPlayer) GetRoundsLost() int { return p.roundsLost }

// SetRoundsLost sets declarer-loss count
func (p *SkatPlayer) SetRoundsLost(n int) { p.roundsLost = n }

// IncRoundsLost increments declarer-loss count
func (p *SkatPlayer) IncRoundsLost() { p.roundsLost++ }

// ResetRound resets per-round player state (bid, declarer flag, points, hand, tricks)
func (p *SkatPlayer) ResetRound() {
	p.bid = -1
	p.isDeclarer = false
	p.cardPoints = 0
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
