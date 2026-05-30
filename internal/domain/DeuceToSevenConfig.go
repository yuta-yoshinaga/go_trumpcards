package domain

// DeuceToSevenCpuCountMin / DeuceToSevenCpuCountMax are the valid CPU opponent
// counts for a 2-7 Triple Draw table.
const (
	DeuceToSevenCpuCountMin = 1
	DeuceToSevenCpuCountMax = 3
)

// DeuceToSevenPlayStyle is the CPU play style for 2-7 Triple Draw.
type DeuceToSevenPlayStyle int

// 2-7 Triple Draw CPU play style constants.
const (
	DeuceToSevenStyleConservative DeuceToSevenPlayStyle = iota
	DeuceToSevenStyleBalanced
	DeuceToSevenStyleAggressive
	DeuceToSevenStyleBluffer
)

// DeuceToSevenPlayStyleNames is the ordered list of display names for
// DeuceToSevenPlayStyle values.
var DeuceToSevenPlayStyleNames = []string{
	"Conservative",
	"Balanced",
	"Aggressive",
	"Bluffer",
}

// DeuceToSevenConfig holds the tunable parameters for a 2-7 Triple Draw table.
type DeuceToSevenConfig struct {
	InitChips    int              // starting chip count per player
	Ante         int              // ante paid by every player each hand
	MinBet       int              // small bet (rounds 1-2); big bet = 2×MinBet (rounds 3-4) for Fixed Limit
	CpuCount     int              // number of CPU opponents (1..3)
	BettingLimit BettingLimitType // fixed / pot / no limit
	CpuMetaAI    bool             // enable session-level learning for CPU
}

// DefaultDeuceToSevenConfig returns the canonical starting config used by CLI,
// Web, and Workers. Fixed Limit is the traditional 2-7 Triple Draw structure
// and is the default.
func DefaultDeuceToSevenConfig() DeuceToSevenConfig {
	return DeuceToSevenConfig{
		InitChips:    1000,
		Ante:         10,
		MinBet:       10,
		CpuCount:     3,
		BettingLimit: BettingLimitFixed,
	}
}

// Validate checks config values against their domain-level constraints. Chip /
// ante / min-bet sanity is enforced at the controller layer via
// cuiutil.ClampIntPtr (see the Poker and Holdem paths).
func (c DeuceToSevenConfig) Validate() error {
	if err := ValidateRange("betting limit", int(c.BettingLimit), int(BettingLimitFixed), int(BettingLimitNoLimit)); err != nil {
		return err
	}
	if err := ValidateRange("CPU player count", c.CpuCount, DeuceToSevenCpuCountMin, DeuceToSevenCpuCountMax); err != nil {
		return err
	}
	return nil
}

// deuceToSevenCpuStyleParams bundles the CPU decision knobs for one play style.
// The "draw size" thresholds are a deuceLowStrength rating (1..4) rather than a
// poker hand rank, mirroring Badugi's BadugiHand.Size scale.
type deuceToSevenCpuStyleParams struct {
	aggressive bool // true = raise-biased, false = call-biased
	bluffRate  int  // percent chance of a bluff on weak hands

	// Early rounds (after deal, after 1st draw).
	earlyBetSize  int // deuceLowStrength >= this → bet/raise for value
	earlyFoldSize int // deuceLowStrength <= this → fold candidate
	earlyCallMult int // Passive only: fold if call amount > MinBet × this

	// Late rounds (after 2nd draw, after 3rd draw).
	lateBetSize  int // strength >= this → value bet/raise
	lateFoldSize int // strength <= this → fold candidate
	lateCallMult int // fold if call amount > MinBet × this

	// Draw decisions.
	drawStandPatSize int // strength >= this → stand pat (draw 0 cards)
	bluffStandPatPct int // weak-hand stand-pat bluff probability
}

// deuceToSevenStyleParamsMap holds the decision knobs per play style. The
// thresholds use deuceLowStrength: 4 = made 8-or-better pat low, 3 = 9-J-high
// low, 2 = queen+ high, 1 = paired/straight/flush.
var deuceToSevenStyleParamsMap = map[DeuceToSevenPlayStyle]deuceToSevenCpuStyleParams{
	DeuceToSevenStyleConservative: {
		aggressive:       false,
		bluffRate:        5,
		earlyBetSize:     4,
		earlyFoldSize:    1,
		earlyCallMult:    2,
		lateBetSize:      4,
		lateFoldSize:     2,
		lateCallMult:     2,
		drawStandPatSize: 4,
		bluffStandPatPct: 0,
	},
	DeuceToSevenStyleBalanced: {
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
	DeuceToSevenStyleAggressive: {
		aggressive:       true,
		bluffRate:        25,
		earlyBetSize:     3,
		earlyFoldSize:    1,
		earlyCallMult:    5,
		lateBetSize:      3,
		lateFoldSize:     2,
		lateCallMult:     4,
		drawStandPatSize: 3,
		bluffStandPatPct: 10,
	},
	DeuceToSevenStyleBluffer: {
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
