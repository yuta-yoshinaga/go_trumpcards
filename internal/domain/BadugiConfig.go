//go:build !js || !wasm || casino

package domain

// BadugiCpuCountMin / BadugiCpuCountMax are the valid CPU opponent counts.
const (
	BadugiCpuCountMin = 1
	BadugiCpuCountMax = 3
)

// BadugiPlayStyle is the CPU play style for Badugi.
type BadugiPlayStyle int

// Badugi CPU play style constants.
const (
	BadugiStyleConservative BadugiPlayStyle = iota
	BadugiStyleBalanced
	BadugiStyleAggressive
	BadugiStyleBluffer
)

// BadugiPlayStyleNames is the ordered list of display names for
// BadugiPlayStyle values.
var BadugiPlayStyleNames = []string{
	"Conservative",
	"Balanced",
	"Aggressive",
	"Bluffer",
}

// BadugiConfig holds the tunable parameters for a Badugi table.
type BadugiConfig struct {
	InitChips    int              // starting chip count per player
	Ante         int              // ante paid by every player each hand
	MinBet       int              // small bet (rounds 1-2); big bet = 2×MinBet (rounds 3-4) for Fixed Limit
	CpuCount     int              // number of CPU opponents (1..3)
	BettingLimit BettingLimitType // fixed / pot / no limit
	CpuMetaAI    bool             // enable session-level learning for CPU
}

// DefaultBadugiConfig returns the canonical starting config used by CLI, Web,
// and Workers. Fixed Limit is the traditional Badugi variant and is the
// default.
func DefaultBadugiConfig() BadugiConfig {
	return BadugiConfig{
		InitChips:    1000,
		Ante:         10,
		MinBet:       10,
		CpuCount:     3,
		BettingLimit: BettingLimitFixed,
	}
}

// Validate checks config values against their domain-level constraints.
// Chip / ante / min-bet sanity is enforced at the controller layer via
// cuiutil.ClampIntPtr (see the Poker and Holdem paths).
func (c BadugiConfig) Validate() error {
	if err := ValidateRange("betting limit", int(c.BettingLimit), int(BettingLimitFixed), int(BettingLimitNoLimit)); err != nil {
		return err
	}
	if err := ValidateRange("CPU player count", c.CpuCount, BadugiCpuCountMin, BadugiCpuCountMax); err != nil {
		return err
	}
	return nil
}

// badugiCpuStyleParams bundles the CPU decision knobs for one play style.
// Field semantics mirror Poker's style params but are tuned for Badugi's
// 4-round / 3-draw structure: the "draw size" threshold is a BadugiHand.Size
// (1..4) rather than a Poker hand rank.
type badugiCpuStyleParams struct {
	aggressive bool // true = raise-biased, false = call-biased
	bluffRate  int  // percent chance of a bluff on weak hands

	// Early rounds (after deal, after 1st draw).
	earlyBetSize  int // BadugiHand.Size >= this → bet/raise for value
	earlyFoldSize int // BadugiHand.Size <= this → fold candidate
	earlyCallMult int // Passive only: fold if call amount > MinBet × this

	// Late rounds (after 2nd draw, after 3rd draw).
	lateBetSize  int // Size >= this → value bet/raise
	lateFoldSize int // Size <= this → fold candidate
	lateCallMult int // fold if call amount > MinBet × this

	// Draw decisions.
	drawStandPatSize int // Size >= this → stand pat (draw 0 cards)
	bluffStandPatPct int // weak-hand stand-pat bluff probability
}

// badugiStyleParamsMap holds the decision knobs per play style. The thresholds
// use BadugiHand.Size: 4 = perfect Badugi, 3 = three-card, 2 = two-card, 1 = paired/suited mess.
var badugiStyleParamsMap = map[BadugiPlayStyle]badugiCpuStyleParams{
	BadugiStyleConservative: {
		aggressive:       false,
		bluffRate:        5,
		earlyBetSize:     3,
		earlyFoldSize:    1,
		earlyCallMult:    2,
		lateBetSize:      4,
		lateFoldSize:     2,
		lateCallMult:     2,
		drawStandPatSize: 4,
		bluffStandPatPct: 0,
	},
	BadugiStyleBalanced: {
		aggressive:       false,
		bluffRate:        15,
		earlyBetSize:     3,
		earlyFoldSize:    1,
		earlyCallMult:    3,
		lateBetSize:      3,
		lateFoldSize:     2,
		lateCallMult:     3,
		drawStandPatSize: 4,
		bluffStandPatPct: 5,
	},
	BadugiStyleAggressive: {
		aggressive:       true,
		bluffRate:        25,
		earlyBetSize:     2,
		earlyFoldSize:    1,
		earlyCallMult:    5,
		lateBetSize:      3,
		lateFoldSize:     2,
		lateCallMult:     4,
		drawStandPatSize: 3,
		bluffStandPatPct: 10,
	},
	BadugiStyleBluffer: {
		aggressive:       true,
		bluffRate:        40,
		earlyBetSize:     2,
		earlyFoldSize:    1,
		earlyCallMult:    4,
		lateBetSize:      2,
		lateFoldSize:     1,
		lateCallMult:     4,
		drawStandPatSize: 3,
		bluffStandPatPct: 20,
	},
}
