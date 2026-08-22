//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// RamschPlayer Ramsch player struct
type RamschPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	cardPoints int // card points won this round (0..120)
	roundsWon  int // declarer wins this game session
	roundsLost int // declarer losses this game session
}

// ramschPlayerJSON is the JSON wire format for RamschPlayer.
type ramschPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	CardPoints       int               `json:"cp"`
	RoundsWon        int               `json:"rw"`
	RoundsLost       int               `json:"rl"`
}

// MarshalJSON implements json.Marshaler.
func (p *RamschPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ramschPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		CardPoints:       p.cardPoints,
		RoundsWon:        p.roundsWon,
		RoundsLost:       p.roundsLost,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RamschPlayer) UnmarshalJSON(data []byte) error {
	var j ramschPlayerJSON
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
	p.cardPoints = j.CardPoints
	p.roundsWon = j.RoundsWon
	p.roundsLost = j.RoundsLost
	return nil
}

// NewRamschPlayer constructs a Ramsch player
func NewRamschPlayer(isHuman bool) *RamschPlayer {
	return &RamschPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetCardPoints returns card points won this round
func (p *RamschPlayer) GetCardPoints() int { return p.cardPoints }

// SetCardPoints sets card points
func (p *RamschPlayer) SetCardPoints(n int) { p.cardPoints = n }

// GetRoundsWon returns the declarer-win count
func (p *RamschPlayer) GetRoundsWon() int { return p.roundsWon }

// SetRoundsWon sets the declarer-win count
func (p *RamschPlayer) SetRoundsWon(n int) { p.roundsWon = n }

// IncRoundsWon increments declarer-win count
func (p *RamschPlayer) IncRoundsWon() { p.roundsWon++ }

// GetRoundsLost returns declarer-loss count
func (p *RamschPlayer) GetRoundsLost() int { return p.roundsLost }

// SetRoundsLost sets declarer-loss count
func (p *RamschPlayer) SetRoundsLost(n int) { p.roundsLost = n }

// IncRoundsLost increments the round-loss count (took the most card points)
func (p *RamschPlayer) IncRoundsLost() { p.roundsLost++ }

// ResetRound resets per-round player state (points, hand, tricks)
func (p *RamschPlayer) ResetRound() {
	p.cardPoints = 0
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
