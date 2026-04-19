package domain

import "encoding/json"

// BadugiPlayer is the betting + card-holding seat for a single Badugi player.
// The raw 4-card hand lives in the embedded Player; the best evaluated
// subset is cached as bestHand after each EvalHand call and reused for
// showdown comparison.
type BadugiPlayer struct {
	Player
	ChipHolder
	bettingPlayerBase
	isHuman      bool
	playStyle    BadugiPlayStyle
	bestHand     BadugiHand // most recent eval result (Size, sorted subset)
	drawCount    int        // cards exchanged in the most recent draw
	totalDrawCnt int        // cumulative draws across all 3 draw rounds (for opponent reads)
}

// NewBadugiPlayer constructs a player with an empty hand and the given style.
func NewBadugiPlayer(isHuman bool, style BadugiPlayStyle) *BadugiPlayer {
	return &BadugiPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// ExchangeCard replaces the card at idx with the supplied card. Out-of-range
// indices are silently ignored so callers can iterate with the draw mask
// without defensive bounds checks at every site.
func (bp *BadugiPlayer) ExchangeCard(idx int, card *Card) {
	if 0 <= idx && idx < len(bp.cards) {
		bp.cards[idx] = card
	}
}

// EvalHand runs the Badugi evaluator and caches the result. The returned
// int is BadugiHand.Size (1..4), stored in handRank for compatibility
// with the shared BettingPlayer interface used by pot distribution.
func (bp *BadugiPlayer) EvalHand() int {
	bp.bestHand = evalBadugiHand(bp.cards)
	bp.handRank = bp.bestHand.Size
	return bp.handRank
}

// GetBestHand returns the most recent cached evaluation.
func (bp *BadugiPlayer) GetBestHand() BadugiHand { return bp.bestHand }

// GetHandName returns the display name for the current handRank (e.g.
// "Badugi", "3-card", "2-card", "1-card"). Returns "Unknown" if no
// evaluation has been run or the rank is out of range.
func (bp *BadugiPlayer) GetHandName() string {
	if 1 <= bp.handRank && bp.handRank < len(BadugiHandNames) {
		return BadugiHandNames[bp.handRank]
	}
	return "Unknown"
}

// GetIsHuman reports whether this seat is the human player.
func (bp *BadugiPlayer) GetIsHuman() bool { return bp.isHuman }

// GetPlayStyle returns the CPU play style (meaningful only when !isHuman).
func (bp *BadugiPlayer) GetPlayStyle() BadugiPlayStyle { return bp.playStyle }

// GetPlayStyleName returns the human-readable play style name.
func (bp *BadugiPlayer) GetPlayStyleName() string {
	return playStyleName(int(bp.playStyle), BadugiPlayStyleNames)
}

// GetDrawCount returns the number of cards exchanged in the most recent draw
// (0 = stand pat, 4 = drew all new cards).
func (bp *BadugiPlayer) GetDrawCount() int { return bp.drawCount }

// SetDrawCount records how many cards were exchanged in the most recent draw.
func (bp *BadugiPlayer) SetDrawCount(n int) { bp.drawCount = n }

// GetTotalDrawCount returns the cumulative number of exchanges across all
// three draw rounds in the current hand. Used by CPU opponent-reads.
func (bp *BadugiPlayer) GetTotalDrawCount() int { return bp.totalDrawCnt }

// AddToTotalDrawCount increases the cumulative draw counter by n.
func (bp *BadugiPlayer) AddToTotalDrawCount(n int) { bp.totalDrawCnt += n }

// ResetDrawCounters clears the per-hand draw counters. Called at hand start.
func (bp *BadugiPlayer) ResetDrawCounters() {
	bp.drawCount = 0
	bp.totalDrawCnt = 0
}

// GetComparisonCards returns the cards used for showdown comparison: the
// cached best-subset from the most recent EvalHand. Satisfies the
// BettingPlayer interface consumed by FindPotWinnersBadugi.
func (bp *BadugiPlayer) GetComparisonCards() []*Card {
	out := make([]*Card, len(bp.bestHand.Cards))
	copy(out, bp.bestHand.Cards)
	return out
}

// badugiPlayerJSON is the persisted wire format. Embedded structs are
// pointer-serialized so their custom JSON codecs run.
type badugiPlayerJSON struct {
	Player            *Player            `json:"p"`
	ChipHolder        *ChipHolder        `json:"ch"`
	BettingPlayerBase *bettingPlayerBase `json:"bp"`
	IsHuman           bool               `json:"ih"`
	PlayStyle         BadugiPlayStyle    `json:"ps"`
	DrawCount         int                `json:"dc"`
	TotalDrawCnt      int                `json:"tdc"`
}

// MarshalJSON implements json.Marshaler. bestHand is intentionally omitted
// because it is a pure function of Player.cards and is recomputed by
// EvalHand after reload.
func (bp *BadugiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(badugiPlayerJSON{
		Player:            &bp.Player,
		ChipHolder:        &bp.ChipHolder,
		BettingPlayerBase: &bp.bettingPlayerBase,
		IsHuman:           bp.isHuman,
		PlayStyle:         bp.playStyle,
		DrawCount:         bp.drawCount,
		TotalDrawCnt:      bp.totalDrawCnt,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (bp *BadugiPlayer) UnmarshalJSON(data []byte) error {
	var j badugiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		bp.Player = *j.Player
	}
	if j.ChipHolder != nil {
		bp.ChipHolder = *j.ChipHolder
	}
	if j.BettingPlayerBase != nil {
		bp.bettingPlayerBase = *j.BettingPlayerBase
	}
	bp.isHuman = j.IsHuman
	bp.playStyle = j.PlayStyle
	bp.drawCount = j.DrawCount
	bp.totalDrawCnt = j.TotalDrawCnt
	return nil
}
