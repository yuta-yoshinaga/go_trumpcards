package controller_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func boolPtr(v bool) *bool { return &v }

// toConfigInput is the constraint satisfied by every *WebInput type whose
// ToConfig() returns Cfg, allowing jsonToConfig to be reused across games.
type toConfigInput[Cfg any] interface {
	ToConfig() (Cfg, error)
}

// jsonToConfig unmarshals raw into a fresh In value and runs ToConfig on it.
// Used to assert that the live JSON wire shape still survives mixin embedding.
func jsonToConfig[In toConfigInput[Cfg], Cfg any](t *testing.T, raw []byte) (Cfg, error) {
	t.Helper()
	var in In
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return in.ToConfig()
}

// ---------------------------------------------------------------------------
// BlackJackWebInput.HasConfigParams
// ---------------------------------------------------------------------------

func TestBlackJackWebInput_HasConfigParams_False(t *testing.T) {
	p := controller.BlackJackWebInput{}
	assert.False(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_DealerHitsSoft17(t *testing.T) {
	p := controller.BlackJackWebInput{DealerHitsSoft17: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_CpuPlayerCount(t *testing.T) {
	p := controller.BlackJackWebInput{CpuPlayerCount: intPtr(2)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_CountingEnabled(t *testing.T) {
	p := controller.BlackJackWebInput{CountingEnabled: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_DoubleAfterSplit(t *testing.T) {
	p := controller.BlackJackWebInput{DoubleAfterSplit: boolPtr(false)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_CountingSystem(t *testing.T) {
	p := controller.BlackJackWebInput{CountingSystem: intPtr(1)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_DeckPenetration(t *testing.T) {
	p := controller.BlackJackWebInput{DeckPenetration: intPtr(75)}
	assert.True(t, p.HasConfigParams())
}

func TestBlackJackWebInput_HasConfigParams_SurrenderRule(t *testing.T) {
	p := controller.BlackJackWebInput{SurrenderRule: intPtr(1)}
	assert.True(t, p.HasConfigParams())
}

// ---------------------------------------------------------------------------
// BlackJackWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestBlackJackWebInput_ToConfig_NilPointers(t *testing.T) {
	p := controller.BlackJackWebInput{}
	cfg := p.ToConfig()
	assert.False(t, cfg.DealerHitsSoft17)
	assert.Equal(t, 0, cfg.CpuPlayerCount)
	assert.False(t, cfg.CountingEnabled)
	// derefDefault with true as default
	assert.True(t, cfg.DoubleAfterSplit)
	assert.Equal(t, 0, cfg.CountingSystem)
	assert.Equal(t, 0, cfg.DeckPenetration)
	assert.Equal(t, 0, cfg.SurrenderRule)
}

func TestBlackJackWebInput_ToConfig_AllSet(t *testing.T) {
	p := controller.BlackJackWebInput{
		DealerHitsSoft17: boolPtr(true),
		CpuPlayerCount:   intPtr(3),
		CountingEnabled:  boolPtr(true),
		DoubleAfterSplit: boolPtr(false),
		CountingSystem:   intPtr(2),
		DeckPenetration:  intPtr(80),
		SurrenderRule:    intPtr(1),
	}
	cfg := p.ToConfig()
	assert.True(t, cfg.DealerHitsSoft17)
	assert.Equal(t, 3, cfg.CpuPlayerCount)
	assert.True(t, cfg.CountingEnabled)
	assert.False(t, cfg.DoubleAfterSplit)
	assert.Equal(t, 2, cfg.CountingSystem)
	assert.Equal(t, 80, cfg.DeckPenetration)
	assert.Equal(t, 1, cfg.SurrenderRule)
}

// ---------------------------------------------------------------------------
// SevensWebInput.HasConfigParams
// ---------------------------------------------------------------------------

func TestSevensWebInput_HasConfigParams_False(t *testing.T) {
	p := controller.SevensWebInput{}
	assert.False(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_TunnelEnabled(t *testing.T) {
	p := controller.SevensWebInput{TunnelEnabled: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_TunnelSkipWidth(t *testing.T) {
	p := controller.SevensWebInput{TunnelSkipWidth: intPtr(2)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_JokerCount(t *testing.T) {
	p := controller.SevensWebInput{JokerCount: intPtr(1)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_CpuStrategy(t *testing.T) {
	p := controller.SevensWebInput{CpuStrategy: intPtr(1)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_MaxPasses(t *testing.T) {
	p := controller.SevensWebInput{MaxPasses: intPtr(3)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_NoJokerFinish(t *testing.T) {
	p := controller.SevensWebInput{NoJokerFinish: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_JokerReclaim(t *testing.T) {
	p := controller.SevensWebInput{JokerReclaim: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_EndStop(t *testing.T) {
	p := controller.SevensWebInput{EndStop: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

func TestSevensWebInput_HasConfigParams_JokerConsecutiveBanned(t *testing.T) {
	p := controller.SevensWebInput{JokerConsecutiveBanned: boolPtr(true)}
	assert.True(t, p.HasConfigParams())
}

// ---------------------------------------------------------------------------
// SevensWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestSevensWebInput_ToConfig_NilPointers(t *testing.T) {
	p := controller.SevensWebInput{}
	cfg := p.ToConfig()
	assert.False(t, cfg.TunnelEnabled)
	assert.Equal(t, 0, cfg.TunnelSkipWidth)
	assert.Equal(t, 0, cfg.JokerCount)
	assert.Equal(t, 0, cfg.CpuStrategy)
	// derefDefault with domain.SevensMaxPasses as default
	assert.Equal(t, domain.SevensMaxPasses, cfg.MaxPasses)
	assert.False(t, cfg.NoJokerFinish)
	assert.False(t, cfg.JokerReclaimEnabled)
	assert.False(t, cfg.EndStopEnabled)
	assert.False(t, cfg.JokerConsecutiveBanned)
}

func TestSevensWebInput_ToConfig_AllSet(t *testing.T) {
	p := controller.SevensWebInput{
		TunnelEnabled:          boolPtr(true),
		TunnelSkipWidth:        intPtr(3),
		JokerCount:             intPtr(1),
		CpuStrategy:            intPtr(2),
		MaxPasses:              intPtr(5),
		NoJokerFinish:          boolPtr(true),
		JokerReclaim:           boolPtr(true),
		EndStop:                boolPtr(true),
		JokerConsecutiveBanned: boolPtr(true),
	}
	cfg := p.ToConfig()
	assert.True(t, cfg.TunnelEnabled)
	assert.Equal(t, 3, cfg.TunnelSkipWidth)
	assert.Equal(t, 1, cfg.JokerCount)
	assert.Equal(t, 2, cfg.CpuStrategy)
	assert.Equal(t, 5, cfg.MaxPasses)
	assert.True(t, cfg.NoJokerFinish)
	assert.True(t, cfg.JokerReclaimEnabled)
	assert.True(t, cfg.EndStopEnabled)
	assert.True(t, cfg.JokerConsecutiveBanned)
}

// ---------------------------------------------------------------------------
// PokerWebInput.ToConfig - bounds clamping branches
// ---------------------------------------------------------------------------

func TestPokerWebInput_ToConfig_NilPointers(t *testing.T) {
	p := controller.PokerWebInput{}
	cfg := p.ToConfig()
	def := domain.DefaultPokerConfig()
	assert.Equal(t, def.CpuCount, cfg.CpuCount)
	assert.Equal(t, def.JokerCount, cfg.JokerCount)
	assert.Equal(t, def.BettingLimit, cfg.BettingLimit)
	assert.Equal(t, def.IsLowball, cfg.IsLowball)
}

func TestPokerWebInput_ToConfig_CpuCountBelowMin(t *testing.T) {
	p := controller.PokerWebInput{CpuCount: intPtr(0)}
	cfg := p.ToConfig()
	assert.Equal(t, 3, cfg.CpuCount) // out-of-range → default
}

func TestPokerWebInput_ToConfig_CpuCountAboveMax(t *testing.T) {
	p := controller.PokerWebInput{CpuCount: intPtr(10)}
	cfg := p.ToConfig()
	assert.Equal(t, 3, cfg.CpuCount) // out-of-range → default
}

func TestPokerWebInput_ToConfig_CpuCountInRange(t *testing.T) {
	p := controller.PokerWebInput{CpuCount: intPtr(2)}
	cfg := p.ToConfig()
	assert.Equal(t, 2, cfg.CpuCount)
}

func TestPokerWebInput_ToConfig_JokerCountBelowMin(t *testing.T) {
	p := controller.PokerWebInput{JokerCount: intPtr(-1)}
	cfg := p.ToConfig()
	assert.Equal(t, 0, cfg.JokerCount)
}

func TestPokerWebInput_ToConfig_JokerCountAboveMax(t *testing.T) {
	p := controller.PokerWebInput{JokerCount: intPtr(5)}
	cfg := p.ToConfig()
	assert.Equal(t, 0, cfg.JokerCount) // out-of-range → default
}

func TestPokerWebInput_ToConfig_JokerCountInRange(t *testing.T) {
	p := controller.PokerWebInput{JokerCount: intPtr(1)}
	cfg := p.ToConfig()
	assert.Equal(t, 1, cfg.JokerCount)
}

func TestPokerWebInput_ToConfig_BettingLimitBelowMin(t *testing.T) {
	p := controller.PokerWebInput{BettingLimit: intPtr(-1)}
	cfg := p.ToConfig()
	assert.Equal(t, domain.BettingLimitType(0), cfg.BettingLimit)
}

func TestPokerWebInput_ToConfig_BettingLimitAboveMax(t *testing.T) {
	p := controller.PokerWebInput{BettingLimit: intPtr(5)}
	cfg := p.ToConfig()
	assert.Equal(t, domain.BettingLimitType(0), cfg.BettingLimit) // out-of-range → default
}

func TestPokerWebInput_ToConfig_BettingLimitInRange(t *testing.T) {
	p := controller.PokerWebInput{BettingLimit: intPtr(1)}
	cfg := p.ToConfig()
	assert.Equal(t, domain.BettingLimitType(1), cfg.BettingLimit)
}

func TestPokerWebInput_ToConfig_IsLowball(t *testing.T) {
	p := controller.PokerWebInput{IsLowball: boolPtr(true)}
	cfg := p.ToConfig()
	assert.True(t, cfg.IsLowball)
}

// ---------------------------------------------------------------------------
// HoldemWebInput.ToConfig - branches
// ---------------------------------------------------------------------------

func TestHoldemWebInput_ToConfig_Defaults(t *testing.T) {
	p := controller.HoldemWebInput{}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	def := domain.DefaultHoldemConfig()
	assert.Equal(t, def.SmallBlind, cfg.SmallBlind)
	assert.Equal(t, def.BigBlind, cfg.BigBlind)
}

func TestHoldemWebInput_ToConfig_BothBlindsSet(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{SmallBlind: intPtr(10), BigBlind: intPtr(20)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 10, cfg.SmallBlind)
	assert.Equal(t, 20, cfg.BigBlind)
}

// Only SmallBlind provided and >= BigBlind default: auto-adjust BigBlind
func TestHoldemWebInput_ToConfig_OnlySmallBlindGeDefault(t *testing.T) {
	def := domain.DefaultHoldemConfig()
	// Set sb >= def.BigBlind to trigger auto-adjust
	sb := def.BigBlind
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{SmallBlind: intPtr(sb)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, sb, cfg.SmallBlind)
	assert.Equal(t, sb*2, cfg.BigBlind)
}

// Only SmallBlind provided but sb < default BigBlind: no auto-adjust, still sb < bb
func TestHoldemWebInput_ToConfig_OnlySmallBlindLtDefault(t *testing.T) {
	// default SmallBlind=5, BigBlind=10; setting sb=7 < 10, no auto-adjust; 7<10 is valid
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{SmallBlind: intPtr(7)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 7, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
}

// Only BigBlind provided, bb > 1: auto-adjust SmallBlind
func TestHoldemWebInput_ToConfig_OnlyBigBlindGt1(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BigBlind: intPtr(20)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 10, cfg.SmallBlind)
	assert.Equal(t, 20, cfg.BigBlind)
}

// Only BigBlind=2: auto sb = bb/2 = 1
func TestHoldemWebInput_ToConfig_OnlyBigBlind2(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BigBlind: intPtr(2)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.SmallBlind)
	assert.Equal(t, 2, cfg.BigBlind)
}

// Only BigBlind=1: no auto-adjust (bb > 1 is false), but sb(5) >= bb(1) → error
func TestHoldemWebInput_ToConfig_OnlyBigBlind1Error(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BigBlind: intPtr(1)}}
	_, err := p.ToConfig()
	assert.Error(t, err)
}

// sb >= bb error
func TestHoldemWebInput_ToConfig_SmallBlindGeqBigBlind(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{SmallBlind: intPtr(10), BigBlind: intPtr(10)}}
	_, err := p.ToConfig()
	assert.Error(t, err)
}

func TestHoldemWebInput_ToConfig_TournamentMode(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{TournamentMode: boolPtr(true)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.True(t, cfg.TournamentMode)
}

func TestHoldemWebInput_ToConfig_BlindLevelHands_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BlindLevelHands: intPtr(5)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 5, cfg.BlindLevelHands)
}

func TestHoldemWebInput_ToConfig_BlindLevelHands_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BlindLevelHands: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	// 0 < 1, so not applied - default stays
	assert.Equal(t, domain.DefaultHoldemConfig().BlindLevelHands, cfg.BlindLevelHands)
}

func TestHoldemWebInput_ToConfig_BlindMultiplier_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BlindMultiplier: intPtr(150)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 150, cfg.BlindMultiplier)
}

func TestHoldemWebInput_ToConfig_BlindMultiplier_TooLow(t *testing.T) {
	p := controller.HoldemWebInput{PokerBlindsInput: controller.PokerBlindsInput{BlindMultiplier: intPtr(100)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	// 100 < 101, so not applied - default stays
	assert.Equal(t, domain.DefaultHoldemConfig().BlindMultiplier, cfg.BlindMultiplier)
}

func TestHoldemWebInput_ToConfig_BettingLimit_BelowMin(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{BettingLimit: intPtr(-1)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.BettingLimitType(0), cfg.BettingLimit)
}

func TestHoldemWebInput_ToConfig_BettingLimit_AboveMax(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{BettingLimit: intPtr(5)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.BettingLimitType(2), cfg.BettingLimit)
}

func TestHoldemWebInput_ToConfig_BettingLimit_InRange(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{BettingLimit: intPtr(1)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.BettingLimitType(1), cfg.BettingLimit)
}

func TestHoldemWebInput_ToConfig_TableSize_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{TableSize: intPtr(6)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 6, cfg.TableSize)
}

func TestHoldemWebInput_ToConfig_TableSize_Invalid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{TableSize: intPtr(5)}}
	_, err := p.ToConfig()
	assert.Error(t, err)
}

func TestHoldemWebInput_ToConfig_RebuyEnabled(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyEnabled: boolPtr(true)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.True(t, cfg.RebuyEnabled)
}

func TestHoldemWebInput_ToConfig_RebuyMaxCount_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyMaxCount: intPtr(5)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 5, cfg.RebuyMaxCount)
}

func TestHoldemWebInput_ToConfig_RebuyMaxCount_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyMaxCount: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultHoldemConfig().RebuyMaxCount, cfg.RebuyMaxCount)
}

func TestHoldemWebInput_ToConfig_RebuyChips_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyChips: intPtr(500)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 500, cfg.RebuyChips)
}

func TestHoldemWebInput_ToConfig_RebuyChips_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyChips: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultHoldemConfig().RebuyChips, cfg.RebuyChips)
}

func TestHoldemWebInput_ToConfig_RebuyPeriodHands_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyPeriodHands: intPtr(10)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 10, cfg.RebuyPeriodHands)
}

func TestHoldemWebInput_ToConfig_RebuyPeriodHands_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{RebuyPeriodHands: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultHoldemConfig().RebuyPeriodHands, cfg.RebuyPeriodHands)
}

func TestHoldemWebInput_ToConfig_AddonEnabled(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{AddonEnabled: boolPtr(true)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.True(t, cfg.AddonEnabled)
}

func TestHoldemWebInput_ToConfig_AddonChips_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{AddonChips: intPtr(2000)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 2000, cfg.AddonChips)
}

func TestHoldemWebInput_ToConfig_AddonChips_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{AddonChips: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultHoldemConfig().AddonChips, cfg.AddonChips)
}

func TestHoldemWebInput_ToConfig_AddonAfterHand_Valid(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{AddonAfterHand: intPtr(15)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, 15, cfg.AddonAfterHand)
}

func TestHoldemWebInput_ToConfig_AddonAfterHand_Zero(t *testing.T) {
	p := controller.HoldemWebInput{PokerCommonInput: controller.PokerCommonInput{AddonAfterHand: intPtr(0)}}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultHoldemConfig().AddonAfterHand, cfg.AddonAfterHand)
}

// ---------------------------------------------------------------------------
// JSON wire compatibility: the embedded PokerCommonInput / PokerBlindsInput
// mixins must keep deserializing the same flat JSON shape the frontend already
// sends. These tests pin that contract so future mixin edits cannot silently
// break the live API.
// ---------------------------------------------------------------------------

func TestHoldemWebInput_FlatJSON_PreservesAllFields(t *testing.T) {
	raw := []byte(`{
		"command": "reset",
		"smallBlind": 25,
		"bigBlind": 50,
		"blindLevelHands": 8,
		"blindMultiplier": 150,
		"tournamentMode": true,
		"bettingLimit": 2,
		"tableSize": 6,
		"rebuyEnabled": true,
		"rebuyMaxCount": 4,
		"rebuyChips": 1500,
		"rebuyPeriodHands": 25,
		"addonEnabled": true,
		"addonChips": 2500,
		"addonAfterHand": 30
	}`)
	cfg, err := jsonToConfig[controller.HoldemWebInput, domain.HoldemConfig](t, raw)
	assert.NoError(t, err)
	assert.Equal(t, 25, cfg.SmallBlind)
	assert.Equal(t, 50, cfg.BigBlind)
	assert.Equal(t, 8, cfg.BlindLevelHands)
	assert.Equal(t, 150, cfg.BlindMultiplier)
	assert.True(t, cfg.TournamentMode)
	assert.Equal(t, domain.BettingLimitType(2), cfg.BettingLimit)
	assert.Equal(t, 6, cfg.TableSize)
	assert.True(t, cfg.RebuyEnabled)
	assert.Equal(t, 4, cfg.RebuyMaxCount)
	assert.Equal(t, 1500, cfg.RebuyChips)
	assert.Equal(t, 25, cfg.RebuyPeriodHands)
	assert.True(t, cfg.AddonEnabled)
	assert.Equal(t, 2500, cfg.AddonChips)
	assert.Equal(t, 30, cfg.AddonAfterHand)
}

func TestPineappleWebInput_FlatJSON_PreservesAllFields(t *testing.T) {
	raw := []byte(`{
		"command": "reset",
		"smallBlind": 15,
		"bigBlind": 30,
		"tournamentMode": true,
		"tableSize": 9,
		"rebuyEnabled": true,
		"rebuyChips": 800,
		"addonEnabled": true,
		"addonChips": 1200
	}`)
	cfg, err := jsonToConfig[controller.PineappleWebInput, domain.PineappleConfig](t, raw)
	assert.NoError(t, err)
	assert.Equal(t, 15, cfg.SmallBlind)
	assert.Equal(t, 30, cfg.BigBlind)
	assert.True(t, cfg.TournamentMode)
	assert.Equal(t, 9, cfg.TableSize)
	assert.True(t, cfg.RebuyEnabled)
	assert.Equal(t, 800, cfg.RebuyChips)
	assert.True(t, cfg.AddonEnabled)
	assert.Equal(t, 1200, cfg.AddonChips)
}

func TestSevenCardStudWebInput_FlatJSON_PreservesAllFields(t *testing.T) {
	raw := []byte(`{
		"command": "reset",
		"ante": 2,
		"bringIn": 4,
		"smallBet": 10,
		"bigBet": 20,
		"tournamentMode": true,
		"anteLevelHands": 5,
		"anteMultiplier": 200,
		"bettingLimit": 1,
		"tableSize": 4,
		"rebuyEnabled": true,
		"rebuyMaxCount": 2,
		"rebuyChips": 500,
		"rebuyPeriodHands": 10,
		"addonEnabled": true,
		"addonChips": 800,
		"addonAfterHand": 15
	}`)
	cfg, err := jsonToConfig[controller.SevenCardStudWebInput, domain.SevenCardStudConfig](t, raw)
	assert.NoError(t, err)
	assert.Equal(t, 2, cfg.Ante)
	assert.Equal(t, 4, cfg.BringIn)
	assert.Equal(t, 10, cfg.SmallBet)
	assert.Equal(t, 20, cfg.BigBet)
	assert.True(t, cfg.TournamentMode)
	assert.Equal(t, 5, cfg.AnteLevelHands)
	assert.Equal(t, 200, cfg.AnteMultiplier)
	assert.Equal(t, domain.BettingLimitType(1), cfg.BettingLimit)
	assert.Equal(t, 4, cfg.TableSize)
	assert.True(t, cfg.RebuyEnabled)
	assert.Equal(t, 2, cfg.RebuyMaxCount)
	assert.Equal(t, 500, cfg.RebuyChips)
	assert.Equal(t, 10, cfg.RebuyPeriodHands)
	assert.True(t, cfg.AddonEnabled)
	assert.Equal(t, 800, cfg.AddonChips)
	assert.Equal(t, 15, cfg.AddonAfterHand)
}

// ---------------------------------------------------------------------------
// OldMaidWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestOldMaidWebInput_ToConfig_ValidMode0(t *testing.T) {
	p := controller.OldMaidWebInput{Mode: 0}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.OldMaidMode(0), cfg.Mode)
}

func TestOldMaidWebInput_ToConfig_ValidMode1(t *testing.T) {
	p := controller.OldMaidWebInput{Mode: int(domain.OldMaidModeJijiNuki)}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.Equal(t, domain.OldMaidModeJijiNuki, cfg.Mode)
}

func TestOldMaidWebInput_ToConfig_InvalidModeNegative(t *testing.T) {
	p := controller.OldMaidWebInput{Mode: -1}
	_, err := p.ToConfig()
	assert.Error(t, err)
}

func TestOldMaidWebInput_ToConfig_InvalidModeAboveMax(t *testing.T) {
	p := controller.OldMaidWebInput{Mode: int(domain.OldMaidModeJijiNuki) + 1}
	_, err := p.ToConfig()
	assert.Error(t, err)
}

func TestOldMaidWebInput_ToConfig_CpuFlags(t *testing.T) {
	p := controller.OldMaidWebInput{
		Mode:                 0,
		CpuPlacementStrategy: true,
		CpuMemoryAI:          true,
		CpuHesitationEnabled: true,
		CpuMetaAI:            true,
	}
	cfg, err := p.ToConfig()
	assert.NoError(t, err)
	assert.True(t, cfg.CpuPlacementStrategy)
	assert.True(t, cfg.CpuMemoryAI)
	assert.True(t, cfg.CpuHesitationEnabled)
	assert.True(t, cfg.CpuMetaAI)
}

// ---------------------------------------------------------------------------
// HeartsWebConfig.ToConfig and HeartsWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestHeartsWebConfig_ToConfig_Defaults(t *testing.T) {
	c := &controller.HeartsWebConfig{}
	cfg := c.ToConfig()
	def := domain.DefaultHeartsConfig()
	assert.Equal(t, def.CpuDifficulty, cfg.CpuDifficulty)
	assert.Equal(t, def.PointLimit, cfg.PointLimit)
	assert.False(t, cfg.OmnibusJD)
}

func TestHeartsWebConfig_ToConfig_ValidCpuDifficulty(t *testing.T) {
	c := &controller.HeartsWebConfig{CpuDifficulty: intPtr(int(domain.HeartsCpuDifficultyHard))}
	cfg := c.ToConfig()
	assert.Equal(t, domain.HeartsCpuDifficultyHard, cfg.CpuDifficulty)
}

func TestHeartsWebConfig_ToConfig_CpuDifficultyBelowRange(t *testing.T) {
	c := &controller.HeartsWebConfig{CpuDifficulty: intPtr(-1)}
	cfg := c.ToConfig()
	// out of range: not applied; default stays
	assert.Equal(t, domain.DefaultHeartsConfig().CpuDifficulty, cfg.CpuDifficulty)
}

func TestHeartsWebConfig_ToConfig_CpuDifficultyAboveRange(t *testing.T) {
	c := &controller.HeartsWebConfig{CpuDifficulty: intPtr(100)}
	cfg := c.ToConfig()
	assert.Equal(t, domain.DefaultHeartsConfig().CpuDifficulty, cfg.CpuDifficulty)
}

func TestHeartsWebConfig_ToConfig_ValidPointLimit(t *testing.T) {
	c := &controller.HeartsWebConfig{PointLimit: intPtr(50)}
	cfg := c.ToConfig()
	assert.Equal(t, 50, cfg.PointLimit)
}

func TestHeartsWebConfig_ToConfig_PointLimitZero(t *testing.T) {
	c := &controller.HeartsWebConfig{PointLimit: intPtr(0)}
	cfg := c.ToConfig()
	// 0 < 1: not applied
	assert.Equal(t, domain.DefaultHeartsConfig().PointLimit, cfg.PointLimit)
}

func TestHeartsWebConfig_ToConfig_PointLimitAboveMax(t *testing.T) {
	c := &controller.HeartsWebConfig{PointLimit: intPtr(1001)}
	cfg := c.ToConfig()
	// > 1000: not applied
	assert.Equal(t, domain.DefaultHeartsConfig().PointLimit, cfg.PointLimit)
}

func TestHeartsWebInput_ToConfig_NilConfig(t *testing.T) {
	p := controller.HeartsWebInput{Config: nil}
	cfg := p.ToConfig()
	def := domain.DefaultHeartsConfig()
	assert.Equal(t, def.CpuDifficulty, cfg.CpuDifficulty)
	assert.Equal(t, def.PointLimit, cfg.PointLimit)
}

func TestHeartsWebInput_ToConfig_WithConfig(t *testing.T) {
	p := controller.HeartsWebInput{
		Config: &controller.HeartsWebConfig{PointLimit: intPtr(75)},
	}
	cfg := p.ToConfig()
	assert.Equal(t, 75, cfg.PointLimit)
}

func TestHeartsWebConfig_ToConfig_OmnibusJDTrue(t *testing.T) {
	c := &controller.HeartsWebConfig{OmnibusJD: boolPtr(true)}
	cfg := c.ToConfig()
	assert.True(t, cfg.OmnibusJD)
}

func TestHeartsWebConfig_ToConfig_OmnibusJDFalse(t *testing.T) {
	c := &controller.HeartsWebConfig{OmnibusJD: boolPtr(false)}
	cfg := c.ToConfig()
	assert.False(t, cfg.OmnibusJD)
}

func TestHeartsWebConfig_ToConfig_OmnibusJDNil(t *testing.T) {
	c := &controller.HeartsWebConfig{}
	cfg := c.ToConfig()
	assert.False(t, cfg.OmnibusJD) // nil → default false
}

// ---------------------------------------------------------------------------
// MemoryWebConfig.ToConfig and MemoryWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestMemoryWebConfig_ToConfig_Defaults(t *testing.T) {
	c := &controller.MemoryWebConfig{}
	cfg := c.ToConfig()
	def := domain.DefaultMemoryConfig()
	assert.Equal(t, def.CpuDifficulty, cfg.CpuDifficulty)
}

func TestMemoryWebConfig_ToConfig_ValidCpuDifficulty(t *testing.T) {
	c := &controller.MemoryWebConfig{CpuDifficulty: intPtr(int(domain.MemoryCpuDifficultyHard))}
	cfg := c.ToConfig()
	assert.Equal(t, domain.MemoryCpuDifficultyHard, cfg.CpuDifficulty)
}

func TestMemoryWebConfig_ToConfig_CpuDifficultyBelowRange(t *testing.T) {
	c := &controller.MemoryWebConfig{CpuDifficulty: intPtr(-1)}
	cfg := c.ToConfig()
	assert.Equal(t, domain.DefaultMemoryConfig().CpuDifficulty, cfg.CpuDifficulty)
}

func TestMemoryWebConfig_ToConfig_CpuDifficultyAboveRange(t *testing.T) {
	c := &controller.MemoryWebConfig{CpuDifficulty: intPtr(100)}
	cfg := c.ToConfig()
	assert.Equal(t, domain.DefaultMemoryConfig().CpuDifficulty, cfg.CpuDifficulty)
}

func TestMemoryWebInput_ToConfig_NilConfig(t *testing.T) {
	p := controller.MemoryWebInput{Config: nil}
	cfg := p.ToConfig()
	def := domain.DefaultMemoryConfig()
	assert.Equal(t, def.CpuDifficulty, cfg.CpuDifficulty)
}

func TestMemoryWebInput_ToConfig_WithConfig(t *testing.T) {
	p := controller.MemoryWebInput{
		Config: &controller.MemoryWebConfig{CpuDifficulty: intPtr(int(domain.MemoryCpuDifficultyEasy))},
	}
	cfg := p.ToConfig()
	assert.Equal(t, domain.MemoryCpuDifficultyEasy, cfg.CpuDifficulty)
}

// ---------------------------------------------------------------------------
// DoubtWebInput.ToConfig
// ---------------------------------------------------------------------------

func TestDoubtWebInput_ToConfig_Defaults(t *testing.T) {
	p := controller.DoubtWebInput{}
	cfg := p.ToConfig()
	def := domain.DefaultDoubtConfig()
	assert.Equal(t, def.DoubtWindowSec, cfg.DoubtWindowSec)
	assert.Equal(t, def.CpuMemoryLevel, cfg.CpuMemoryLevel)
	assert.Equal(t, def.PenaltyDrawLimit, cfg.PenaltyDrawLimit)
	assert.False(t, cfg.CpuHesitationEnabled)
	assert.False(t, cfg.CpuMetaAI)
}

func TestDoubtWebInput_ToConfig_DoubtWindowSec_Valid(t *testing.T) {
	p := controller.DoubtWebInput{DoubtWindowSec: intPtr(5)}
	cfg := p.ToConfig()
	assert.Equal(t, 5, cfg.DoubtWindowSec)
}

func TestDoubtWebInput_ToConfig_DoubtWindowSec_Zero(t *testing.T) {
	p := controller.DoubtWebInput{DoubtWindowSec: intPtr(0)}
	cfg := p.ToConfig()
	// 0 < 1: not applied
	assert.Equal(t, domain.DefaultDoubtConfig().DoubtWindowSec, cfg.DoubtWindowSec)
}

func TestDoubtWebInput_ToConfig_CpuMemoryLevel_Valid(t *testing.T) {
	p := controller.DoubtWebInput{CpuMemoryLevel: intPtr(int(domain.DoubtMemoryLevelHard))}
	cfg := p.ToConfig()
	assert.Equal(t, domain.DoubtMemoryLevelHard, cfg.CpuMemoryLevel)
}

func TestDoubtWebInput_ToConfig_CpuMemoryLevel_BelowRange(t *testing.T) {
	p := controller.DoubtWebInput{CpuMemoryLevel: intPtr(-1)}
	cfg := p.ToConfig()
	assert.Equal(t, domain.DefaultDoubtConfig().CpuMemoryLevel, cfg.CpuMemoryLevel)
}

func TestDoubtWebInput_ToConfig_CpuMemoryLevel_AboveRange(t *testing.T) {
	p := controller.DoubtWebInput{CpuMemoryLevel: intPtr(100)}
	cfg := p.ToConfig()
	assert.Equal(t, domain.DefaultDoubtConfig().CpuMemoryLevel, cfg.CpuMemoryLevel)
}

func TestDoubtWebInput_ToConfig_PenaltyDrawLimit_Valid(t *testing.T) {
	p := controller.DoubtWebInput{PenaltyDrawLimit: intPtr(3)}
	cfg := p.ToConfig()
	assert.Equal(t, 3, cfg.PenaltyDrawLimit)
}

func TestDoubtWebInput_ToConfig_PenaltyDrawLimit_Zero(t *testing.T) {
	// 0 >= 0: should be applied
	p := controller.DoubtWebInput{PenaltyDrawLimit: intPtr(0)}
	cfg := p.ToConfig()
	assert.Equal(t, 0, cfg.PenaltyDrawLimit)
}

func TestDoubtWebInput_ToConfig_CpuFlags(t *testing.T) {
	p := controller.DoubtWebInput{CpuHesitationEnabled: true, CpuMetaAI: true}
	cfg := p.ToConfig()
	assert.True(t, cfg.CpuHesitationEnabled)
	assert.True(t, cfg.CpuMetaAI)
}

// ---------------------------------------------------------------------------
// DaifugoWebConfig.ToConfig
// ---------------------------------------------------------------------------

func TestDaifugoWebConfig_ToConfig(t *testing.T) {
	c := controller.DaifugoWebConfig{
		JokerCount:                2,
		EightCutEnabled:           true,
		SuitLockMode:              1,
		ElevenBackEnabled:         true,
		SequenceEnabled:           true,
		CardExchangeEnabled:       true,
		BlindExchangeEnabled:      true,
		FiveSkipEnabled:           true,
		FiveSkipCount:             3,
		SevenPassEnabled:          true,
		TenDiscardEnabled:         true,
		SpadeThreeEnabled:         true,
		CapitalFallEnabled:        true,
		NineReverseEnabled:        true,
		CoupDetatEnabled:          true,
		NumberLockEnabled:         true,
		SandstormEnabled:          true,
		EmperorEnabled:            true,
		SequenceRevolutionEnabled: true,
		SequenceLockEnabled:       true,
		IllegalFinishEnabled:      true,
		QueenBomberEnabled:        true,
		CpuDifficulty:             2,
	}
	cfg := c.ToConfig()

	assert.Equal(t, 2, cfg.JokerCount)
	assert.True(t, cfg.EightCutEnabled)
	assert.Equal(t, domain.DaifugoSuitLockMode(1), cfg.SuitLockMode)
	assert.True(t, cfg.ElevenBackEnabled)
	assert.True(t, cfg.SequenceEnabled)
	assert.True(t, cfg.CardExchangeEnabled)
	assert.True(t, cfg.BlindExchangeEnabled)
	assert.True(t, cfg.FiveSkipEnabled)
	assert.Equal(t, 3, cfg.FiveSkipCount)
	assert.True(t, cfg.SevenPassEnabled)
	assert.True(t, cfg.TenDiscardEnabled)
	assert.True(t, cfg.SpadeThreeEnabled)
	assert.True(t, cfg.CapitalFallEnabled)
	assert.True(t, cfg.NineReverseEnabled)
	assert.True(t, cfg.CoupDetatEnabled)
	assert.True(t, cfg.NumberLockEnabled)
	assert.True(t, cfg.SandstormEnabled)
	assert.True(t, cfg.EmperorEnabled)
	assert.True(t, cfg.SequenceRevolutionEnabled)
	assert.True(t, cfg.SequenceLockEnabled)
	assert.True(t, cfg.IllegalFinishEnabled)
	assert.True(t, cfg.QueenBomberEnabled)
	assert.Equal(t, domain.DaifugoCpuDifficulty(2), cfg.CpuDifficulty)
}

func TestDaifugoWebConfig_ToConfig_Zeros(t *testing.T) {
	c := controller.DaifugoWebConfig{}
	cfg := c.ToConfig()

	assert.Equal(t, 0, cfg.JokerCount)
	assert.False(t, cfg.EightCutEnabled)
	assert.Equal(t, domain.DaifugoSuitLockMode(0), cfg.SuitLockMode)
	assert.Equal(t, domain.DaifugoCpuDifficulty(0), cfg.CpuDifficulty)
}
