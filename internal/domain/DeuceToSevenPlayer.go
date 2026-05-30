package domain

import "encoding/json"

// DeuceToSevenPlayer is the betting + card-holding seat for a single 2-7 Triple
// Draw player. The raw 5-card hand lives in the embedded Player; the poker
// category (handRank) and the lowball strength rating are cached after each
// EvalHand call for showdown display and CPU decisions respectively.
type DeuceToSevenPlayer struct {
	Player
	ChipHolder
	bettingPlayerBase
	isHuman      bool
	playStyle    DeuceToSevenPlayStyle
	strength     int // cached deuceLowStrength (1..4) from the most recent EvalHand
	drawCount    int // cards exchanged in the most recent draw
	totalDrawCnt int // cumulative draws across all 3 draw rounds (for opponent reads)
}

// NewDeuceToSevenPlayer constructs a player with an empty hand and the given style.
func NewDeuceToSevenPlayer(isHuman bool, style DeuceToSevenPlayStyle) *DeuceToSevenPlayer {
	return &DeuceToSevenPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// ExchangeCard replaces the card at idx with the supplied card. Out-of-range
// indices are silently ignored so callers can iterate with the draw mask
// without defensive bounds checks at every site.
func (dp *DeuceToSevenPlayer) ExchangeCard(idx int, card *Card) {
	if 0 <= idx && idx < len(dp.cards) {
		dp.cards[idx] = card
	}
}

// EvalHand runs the 2-7 evaluator and caches the result. handRank stores the
// standard poker category (PokerHandHighCard … PokerHandRoyalFlush) for display
// and strength stores the 1..4 lowball rating used by the CPU. Returns handRank.
func (dp *DeuceToSevenPlayer) EvalHand() int {
	dp.handRank = evalDeuceToSevenHand(dp.cards)
	dp.strength = deuceLowStrength(dp.cards)
	return dp.handRank
}

// GetStrength returns the cached lowball strength rating (1..4) from the most
// recent EvalHand. Higher is a stronger low.
func (dp *DeuceToSevenPlayer) GetStrength() int { return dp.strength }

// GetHandName returns the display name for the cached poker category (e.g.
// "High Card", "One Pair"). Returns "Unknown" if no evaluation has run or the
// rank is out of range. A 2-7 low is "read" as an ordinary poker hand, so the
// shared PokerHandNames slice is reused.
func (dp *DeuceToSevenPlayer) GetHandName() string {
	if 0 <= dp.handRank && dp.handRank < len(PokerHandNames) {
		return PokerHandNames[dp.handRank]
	}
	return "Unknown"
}

// GetIsHuman reports whether this seat is the human player.
func (dp *DeuceToSevenPlayer) GetIsHuman() bool { return dp.isHuman }

// GetPlayStyle returns the CPU play style (meaningful only when !isHuman).
func (dp *DeuceToSevenPlayer) GetPlayStyle() DeuceToSevenPlayStyle { return dp.playStyle }

// GetPlayStyleName returns the human-readable play style name.
func (dp *DeuceToSevenPlayer) GetPlayStyleName() string {
	return playStyleName(int(dp.playStyle), DeuceToSevenPlayStyleNames)
}

// GetDrawCount returns the number of cards exchanged in the most recent draw
// (0 = stand pat, 5 = drew an entirely new hand).
func (dp *DeuceToSevenPlayer) GetDrawCount() int { return dp.drawCount }

// SetDrawCount records how many cards were exchanged in the most recent draw.
func (dp *DeuceToSevenPlayer) SetDrawCount(n int) { dp.drawCount = n }

// GetTotalDrawCount returns the cumulative number of exchanges across all three
// draw rounds in the current hand. Used by CPU opponent-reads.
func (dp *DeuceToSevenPlayer) GetTotalDrawCount() int { return dp.totalDrawCnt }

// AddToTotalDrawCount increases the cumulative draw counter by n.
func (dp *DeuceToSevenPlayer) AddToTotalDrawCount(n int) { dp.totalDrawCnt += n }

// ResetDrawCounters clears the per-hand draw counters. Called at hand start.
func (dp *DeuceToSevenPlayer) ResetDrawCounters() {
	dp.drawCount = 0
	dp.totalDrawCnt = 0
}

// GetComparisonCards returns a copy of the player's full 5-card hand, used for
// showdown comparison. Satisfies the BettingPlayer interface consumed by
// FindPotWinnersDeuceToSeven.
func (dp *DeuceToSevenPlayer) GetComparisonCards() []*Card {
	out := make([]*Card, len(dp.cards))
	copy(out, dp.cards)
	return out
}

// deuceToSevenPlayerJSON is the persisted wire format. Embedded structs are
// pointer-serialized so their custom JSON codecs run.
type deuceToSevenPlayerJSON struct {
	Player            *Player               `json:"p"`
	ChipHolder        *ChipHolder           `json:"ch"`
	BettingPlayerBase *bettingPlayerBase    `json:"bp"`
	IsHuman           bool                  `json:"ih"`
	PlayStyle         DeuceToSevenPlayStyle `json:"ps"`
	DrawCount         int                   `json:"dc"`
	TotalDrawCnt      int                   `json:"tdc"`
}

// MarshalJSON implements json.Marshaler. strength is intentionally omitted
// because it is a pure function of Player.cards and is recomputed by EvalHand
// after reload.
func (dp *DeuceToSevenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(deuceToSevenPlayerJSON{
		Player:            &dp.Player,
		ChipHolder:        &dp.ChipHolder,
		BettingPlayerBase: &dp.bettingPlayerBase,
		IsHuman:           dp.isHuman,
		PlayStyle:         dp.playStyle,
		DrawCount:         dp.drawCount,
		TotalDrawCnt:      dp.totalDrawCnt,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (dp *DeuceToSevenPlayer) UnmarshalJSON(data []byte) error {
	var j deuceToSevenPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		dp.Player = *j.Player
	}
	if j.ChipHolder != nil {
		dp.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		dp.bettingPlayerBase = *j.BettingPlayerBase
	}
	dp.isHuman = j.IsHuman
	dp.playStyle = j.PlayStyle
	dp.drawCount = j.DrawCount
	dp.totalDrawCnt = j.TotalDrawCnt
	return nil
}
