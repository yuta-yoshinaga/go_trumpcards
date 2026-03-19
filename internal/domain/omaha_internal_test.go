package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestOmaha() *Omaha {
	players := []*OmahaPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleLAP),
		NewOmahaPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultOmahaConfig()
	tc := NewTrumpCards(0)
	o := NewOmaha(tc, players, cfg)
	for _, p := range o.players {
		p.SetChips(1000)
	}
	return o
}

func TestOmaha_EvalPreFlopStrength(t *testing.T) {
	t.Run("less than 4 cards returns 0", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		assert.Equal(t, 0, o.evalPreFlopStrength(0))
	})

	t.Run("AAKK double suited is strong", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 1, false))
		o.players[0].AddCard(NewCard(CardDesignSpade, 13, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 13, false))
		strength := o.evalPreFlopStrength(0)
		assert.Greater(t, strength, 50)
	})

	t.Run("weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 2, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
		o.players[0].AddCard(NewCard(CardDesignClover, 8, false))
		o.players[0].AddCard(NewCard(CardDesignDiamond, 11, false))
		strength := o.evalPreFlopStrength(0)
		assert.Less(t, strength, 50)
	})

	t.Run("three same suit penalized", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		o.players[0].AddCard(NewCard(CardDesignSpade, 13, false))
		o.players[0].AddCard(NewCard(CardDesignSpade, 12, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 11, false))
		s1 := o.evalPreFlopStrength(0)

		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 13, false))
		o.players[0].AddCard(NewCard(CardDesignClover, 12, false))
		o.players[0].AddCard(NewCard(CardDesignDiamond, 11, false))
		s2 := o.evalPreFlopStrength(0)

		// 3 same suit should score lower than rainbow
		assert.Less(t, s1, s2)
	})
}

func TestOmaha_CpuDecide(t *testing.T) {
	t.Run("GTO preflop", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 10
		o.pot = 30
		for _, p := range o.players {
			p.Reset()
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 1, false))
			p.AddCard(NewCard(CardDesignClover, 13, false))
			p.AddCard(NewCard(CardDesignDiamond, 13, false))
		}
		// Create a GTO player
		gtoPlayer := NewOmahaPlayer(false, HoldemStyleGTO)
		gtoPlayer.SetChips(1000)
		gtoPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignHeart, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignClover, 13, false))
		gtoPlayer.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = gtoPlayer

		action, _ := o.cpuDecide(1)
		// Should produce a valid action
		assert.True(t, action >= OmahaActionFold && action <= OmahaActionAllIn)
	})

	t.Run("GTO postflop", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 0
		o.pot = 50
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		gtoPlayer := NewOmahaPlayer(false, HoldemStyleGTO)
		gtoPlayer.SetChips(1000)
		gtoPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignHeart, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignClover, 13, false))
		gtoPlayer.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = gtoPlayer

		action, _ := o.cpuDecide(1)
		assert.True(t, action >= OmahaActionFold && action <= OmahaActionAllIn)
	})

	t.Run("unknown style defaults to callOrCheck", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 10
		o.pot = 30
		unknownPlayer := NewOmahaPlayer(false, HoldemPlayStyle(99))
		unknownPlayer.SetChips(1000)
		unknownPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
		unknownPlayer.AddCard(NewCard(CardDesignHeart, 2, false))
		unknownPlayer.AddCard(NewCard(CardDesignClover, 3, false))
		unknownPlayer.AddCard(NewCard(CardDesignDiamond, 4, false))
		o.players[1] = unknownPlayer

		action, _ := o.cpuDecide(1)
		assert.True(t, action == OmahaActionCall || action == OmahaActionCheck)
	})
}

func TestOmaha_HandleCpuActionError(t *testing.T) {
	t.Run("fold on error with callAmount > 0", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 50
		o.pot = 100
		o.actedFlags = make([]bool, len(o.players))
		o.startingChips = []int{1000, 1000, 1000, 1000}
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		for _, p := range o.players {
			p.Reset()
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 13, false))
			p.AddCard(NewCard(CardDesignClover, 12, false))
			p.AddCard(NewCard(CardDesignDiamond, 11, false))
		}
		o.handleCpuActionError(1, OmahaActionRaise, fmt.Errorf("test error"))
		assert.NotNil(t, o.lastCpuError)
		assert.True(t, o.players[1].GetFolded())
	})

	t.Run("check on error with no outstanding bet", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 0
		o.pot = 100
		o.actedFlags = make([]bool, len(o.players))
		o.startingChips = []int{1000, 1000, 1000, 1000}
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		for _, p := range o.players {
			p.Reset()
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 13, false))
			p.AddCard(NewCard(CardDesignClover, 12, false))
			p.AddCard(NewCard(CardDesignDiamond, 11, false))
		}
		o.handleCpuActionError(1, OmahaActionBet, fmt.Errorf("test error"))
		assert.NotNil(t, o.lastCpuError)
		assert.False(t, o.players[1].GetFolded())
	})
}

func TestOmaha_BettingLimits(t *testing.T) {
	t.Run("PotLimit", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.config.BettingLimit = BettingLimitPotLimit
		o.pot = 100
		o.lastBet = 20
		maxRaises, maxBetAmount := o.bettingLimits()
		assert.Greater(t, maxRaises, 0)
		assert.Greater(t, maxBetAmount, 0)
	})
}

func TestOmaha_CpuDecideWithRaiseCap(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.config.BettingLimit = BettingLimitFixed
	o.raiseCount = 4
	o.lastBet = 40
	o.pot = 200
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 1, false))
		p.AddCard(NewCard(CardDesignClover, 13, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}
	// At raise cap, a raise decision should become call/check
	action, _ := o.cpuDecide(1)
	assert.NotEqual(t, OmahaActionRaise, action)
	assert.NotEqual(t, OmahaActionBet, action)
}

func TestOmaha_TrackPreFlopStats(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhasePreFlop
	o.vpipTracked = make([]bool, 4)
	o.pfrTracked = make([]bool, 4)
	o.threeBetTracked = make([]bool, 4)

	o.trackPreFlopStats(1, OmahaActionCall)
	assert.Equal(t, 1, o.players[1].GetVPIPCount())
	assert.Equal(t, 0, o.players[1].GetPFRCount())

	o.vpipTracked[2] = false
	o.pfrTracked[2] = false
	o.trackPreFlopStats(2, OmahaActionRaise)
	assert.Equal(t, 1, o.players[2].GetVPIPCount())
	assert.Equal(t, 1, o.players[2].GetPFRCount())
}

func TestOmaha_TrackPreFlopStats_WrongPhase(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.vpipTracked = make([]bool, 4)
	o.pfrTracked = make([]bool, 4)
	o.threeBetTracked = make([]bool, 4)

	o.trackPreFlopStats(1, OmahaActionCall)
	assert.Equal(t, 0, o.players[1].GetVPIPCount())
}

func TestOmaha_TrackPostFlopStats(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop

	o.trackPostFlopStats(1, OmahaActionBet)
	assert.Equal(t, 1, o.players[1].GetPostFlopBetRaise())

	o.trackPostFlopStats(1, OmahaActionCall)
	assert.Equal(t, 1, o.players[1].GetPostFlopCall())
}

func TestOmaha_TrackPostFlopStats_WrongPhase(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhasePreFlop

	o.trackPostFlopStats(1, OmahaActionBet)
	assert.Equal(t, 0, o.players[1].GetPostFlopBetRaise())
}

func TestOmaha_ThreeBetTracking(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhasePreFlop
	o.vpipTracked = make([]bool, 4)
	o.pfrTracked = make([]bool, 4)
	o.threeBetTracked = make([]bool, 4)
	o.raiseCount = 1

	o.trackPreFlopStats(1, OmahaActionRaise)
	assert.Equal(t, 1, o.players[1].GetThreeBetOpportunity())
	assert.Equal(t, 1, o.players[1].GetThreeBetCount())
}

func TestOmaha_GetEquity(t *testing.T) {
	setup := func(phase int) *Omaha {
		o := newInternalTestOmaha()
		o.phase = phase
		o.currentTurn = 0
		o.players[0].Reset()
		o.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		o.players[0].AddCard(NewCard(CardDesignHeart, 1, false))
		o.players[0].AddCard(NewCard(CardDesignClover, 13, false))
		o.players[0].AddCard(NewCard(CardDesignDiamond, 13, false))
		return o
	}

	t.Run("returns nil for init phase", func(t *testing.T) {
		o := setup(OmahaPhaseInit)
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns nil for showdown phase", func(t *testing.T) {
		o := setup(OmahaPhaseShowdown)
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns nil for end phase", func(t *testing.T) {
		o := setup(OmahaPhaseEnd)
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns nil for rebuy phase", func(t *testing.T) {
		o := setup(OmahaPhaseRebuy)
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns nil when human folded", func(t *testing.T) {
		o := setup(OmahaPhaseFlop)
		o.players[0].SetFolded(true)
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns nil when no human player", func(t *testing.T) {
		players := []*OmahaPlayer{
			NewOmahaPlayer(false, HoldemStyleTAG),
			NewOmahaPlayer(false, HoldemStyleLAP),
		}
		cfg := DefaultOmahaConfig()
		tc := NewTrumpCards(0)
		o := NewOmaha(tc, players, cfg)
		o.phase = OmahaPhasePreFlop
		for _, p := range o.players {
			p.SetChips(1000)
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 13, false))
			p.AddCard(NewCard(CardDesignClover, 12, false))
			p.AddCard(NewCard(CardDesignDiamond, 11, false))
		}
		assert.Nil(t, o.GetEquity())
	})

	t.Run("returns result during preflop", func(t *testing.T) {
		o := setup(OmahaPhasePreFlop)
		result := o.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
		assert.Len(t, result.HandOdds, len(PokerHandNames))
	})

	t.Run("returns result during flop", func(t *testing.T) {
		o := setup(OmahaPhaseFlop)
		o.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		result := o.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
	})

	t.Run("returns result during turn", func(t *testing.T) {
		o := setup(OmahaPhaseTurn)
		o.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
		}
		result := o.GetEquity()
		assert.NotNil(t, result)
	})

	t.Run("returns result during river", func(t *testing.T) {
		o := setup(OmahaPhaseRiver)
		o.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		result := o.GetEquity()
		assert.NotNil(t, result)
	})
}

func TestOmaha_GetPotOdds(t *testing.T) {
	setup := func(phase int) *Omaha {
		o := newInternalTestOmaha()
		o.phase = phase
		o.pot = 100
		o.lastBet = 50
		o.players[0].SetCurrentBet(0)
		return o
	}

	t.Run("returns 0 for init phase", func(t *testing.T) {
		o := setup(OmahaPhaseInit)
		assert.Equal(t, 0.0, o.GetPotOdds())
	})

	t.Run("returns 0 for showdown phase", func(t *testing.T) {
		o := setup(OmahaPhaseShowdown)
		assert.Equal(t, 0.0, o.GetPotOdds())
	})

	t.Run("returns correct pot odds during active phase", func(t *testing.T) {
		o := setup(OmahaPhasePreFlop)
		result := o.GetPotOdds()
		assert.InDelta(t, 33.33, result, 0.01)
	})

	t.Run("returns 0 when no outstanding bet", func(t *testing.T) {
		o := setup(OmahaPhaseFlop)
		o.lastBet = 0
		assert.Equal(t, 0.0, o.GetPotOdds())
	})

	t.Run("accounts for human current bet", func(t *testing.T) {
		o := setup(OmahaPhasePreFlop)
		o.players[0].SetCurrentBet(20)
		result := o.GetPotOdds()
		assert.InDelta(t, 23.08, result, 0.01)
	})

	t.Run("returns 0 when humanCurrentBet exceeds lastBet", func(t *testing.T) {
		o := setup(OmahaPhasePreFlop)
		o.lastBet = 10
		o.players[0].SetCurrentBet(50)
		assert.Equal(t, 0.0, o.GetPotOdds())
	})
}

func TestOmaha_CpuBetOrAllIn(t *testing.T) {
	o := newInternalTestOmaha()
	p := o.players[1]
	p.SetChips(50)

	t.Run("bet when enough chips", func(t *testing.T) {
		action, amount := o.cpuBetOrAllIn(p, 30)
		assert.Equal(t, OmahaActionBet, action)
		assert.Equal(t, 30, amount)
	})

	t.Run("all-in when not enough chips", func(t *testing.T) {
		action, _ := o.cpuBetOrAllIn(p, 100)
		assert.Equal(t, OmahaActionAllIn, action)
	})
}

func TestOmaha_GetHandName(t *testing.T) {
	o := newInternalTestOmaha()
	assert.Equal(t, PokerHandNames[PokerHandHighCard], o.getHandName(PokerHandHighCard))
	assert.Equal(t, "Unknown", o.getHandName(999))
}

func TestOmaha_TournamentBlindEscalation(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.TournamentMode = true
	o.config.BlindLevelHands = 2
	o.config.BlindMultiplier = 200

	originalSB := o.config.SmallBlind
	originalBB := o.config.BigBlind

	_ = o.Reset() // handCount 0→1
	_ = o.Reset() // handCount 1→2
	_ = o.Reset() // handCount 2 → escalation triggers, then handCount→3
	assert.GreaterOrEqual(t, o.config.SmallBlind, originalSB*2)
	assert.GreaterOrEqual(t, o.config.BigBlind, originalBB*2)
}

func TestOmaha_RebuyFlow(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20

	// Make human bust
	o.players[0].SetChips(0)
	o.handCount = 0
	o.rebuyCounts = make([]int, 4)

	err := o.Reset()
	assert.NoError(t, err)
	// Should be in rebuy phase
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseRebuy, o.rebuyPhaseType)

	// Execute rebuy
	err = o.Rebuy()
	assert.NoError(t, err)
	assert.True(t, o.players[0].GetChips() > 0)
}

func TestOmaha_SkipRebuyBust(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20

	o.players[0].SetChips(0)
	o.handCount = 0
	o.rebuyCounts = make([]int, 4)

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, OmahaPhaseRebuy, o.phase)

	err = o.SkipRebuy()
	assert.NoError(t, err)
	assert.Equal(t, OmahaPhaseEnd, o.phase)
	assert.True(t, o.gameEndFlag)
}

func TestOmaha_AddonFlow(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.AddonEnabled = true
	o.config.AddonChips = 1500
	o.config.AddonAfterHand = 1
	o.handCount = 0
	o.addonUsed = make([]bool, 4)

	err := o.Reset()
	assert.NoError(t, err)
	// handCount becomes 1, matches AddonAfterHand
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseAddon, o.rebuyPhaseType)

	chipsBefore := o.players[0].GetChips()
	err = o.Addon()
	assert.NoError(t, err)
	assert.Equal(t, chipsBefore+1500, o.players[0].GetChips())
}

func TestOmaha_SkipAddonFlow(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.AddonEnabled = true
	o.config.AddonChips = 1500
	o.config.AddonAfterHand = 1
	o.handCount = 0
	o.addonUsed = make([]bool, 4)

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, OmahaRebuyPhaseAddon, o.rebuyPhaseType)

	chipsBefore := o.players[0].GetChips()
	err = o.SkipAddon()
	assert.NoError(t, err)
	assert.Equal(t, chipsBefore, o.players[0].GetChips())
}

func TestOmaha_CpuPreFlopDecisions(t *testing.T) {
	styles := []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG}
	for _, style := range styles {
		t.Run(fmt.Sprintf("style_%d", style), func(t *testing.T) {
			o := newInternalTestOmaha()
			o.phase = OmahaPhasePreFlop
			o.lastBet = 10
			o.pot = 30
			o.minRaise = 10
			p := NewOmahaPlayer(false, style)
			p.SetChips(1000)
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 1, false))
			p.AddCard(NewCard(CardDesignClover, 13, false))
			p.AddCard(NewCard(CardDesignDiamond, 13, false))
			o.players[1] = p

			action, _ := o.cpuDecide(1)
			assert.True(t, action >= OmahaActionFold && action <= OmahaActionAllIn)
		})
	}
}

func TestOmaha_CpuPostFlopDecisions(t *testing.T) {
	styles := []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG}
	for _, style := range styles {
		t.Run(fmt.Sprintf("style_%d", style), func(t *testing.T) {
			o := newInternalTestOmaha()
			o.phase = OmahaPhaseFlop
			o.lastBet = 0
			o.pot = 50
			o.minRaise = 10
			o.communityCards = []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignClover, 9, false),
			}
			p := NewOmahaPlayer(false, style)
			p.SetChips(1000)
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 1, false))
			p.AddCard(NewCard(CardDesignClover, 13, false))
			p.AddCard(NewCard(CardDesignDiamond, 13, false))
			o.players[1] = p

			action, _ := o.cpuDecide(1)
			assert.True(t, action >= OmahaActionFold && action <= OmahaActionAllIn)
		})
	}
}

func TestOmaha_GTO_RaiseCap(t *testing.T) {
	t.Run("GTO preflop raise capped to call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.config.BettingLimit = BettingLimitFixed
		o.raiseCount = 4
		o.lastBet = 40
		o.pot = 200
		o.minRaise = 10
		gtoPlayer := NewOmahaPlayer(false, HoldemStyleGTO)
		gtoPlayer.SetChips(1000)
		gtoPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignHeart, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignClover, 13, false))
		gtoPlayer.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = gtoPlayer

		// Run multiple times to hit the raise cap branch
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecide(1)
			assert.NotEqual(t, OmahaActionRaise, action)
			assert.NotEqual(t, OmahaActionBet, action)
		}
	})
}

func TestOmaha_GTO_RaiseCap_CheckFallback(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhasePreFlop
	o.config.BettingLimit = BettingLimitFixed
	o.raiseCount = 4
	o.lastBet = 0 // no call required
	o.pot = 200
	o.minRaise = 10
	gtoPlayer := NewOmahaPlayer(false, HoldemStyleGTO)
	gtoPlayer.SetChips(1000)
	gtoPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
	gtoPlayer.AddCard(NewCard(CardDesignHeart, 1, false))
	gtoPlayer.AddCard(NewCard(CardDesignClover, 13, false))
	gtoPlayer.AddCard(NewCard(CardDesignDiamond, 13, false))
	o.players[1] = gtoPlayer

	hitCheck := false
	for i := 0; i < 100; i++ {
		action, _ := o.cpuDecide(1)
		if action == OmahaActionCheck {
			hitCheck = true
		}
		assert.NotEqual(t, OmahaActionRaise, action)
		assert.NotEqual(t, OmahaActionBet, action)
	}
	assert.True(t, hitCheck)
}
