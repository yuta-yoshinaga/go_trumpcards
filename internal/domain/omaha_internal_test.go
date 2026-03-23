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

// --- Additional coverage tests ---

func TestOmaha_InternalGetters(t *testing.T) {
	o := newInternalTestOmaha()
	_ = o.Reset()

	t.Run("GetPlayers", func(t *testing.T) {
		players := o.GetPlayers()
		assert.Equal(t, 4, len(players))
	})

	t.Run("GetCurrentTurn", func(t *testing.T) {
		turn := o.GetCurrentTurn()
		assert.GreaterOrEqual(t, turn, 0)
		assert.Less(t, turn, 4)
	})

	t.Run("GetMinRaise", func(t *testing.T) {
		mr := o.GetMinRaise()
		assert.Greater(t, mr, 0)
	})

	t.Run("GetRaiseCount", func(t *testing.T) {
		rc := o.GetRaiseCount()
		assert.GreaterOrEqual(t, rc, 0)
	})
}

func TestOmaha_ResolveLastPlayer(t *testing.T) {
	t.Run("all CPUs fold leaves human winner", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.pot = 200
		o.lastBet = 50
		o.minRaise = 10
		o.actedFlags = make([]bool, 4)
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

		// Fold all CPUs manually
		for i := 1; i < 4; i++ {
			o.players[i].SetFolded(true)
		}
		o.resolveLastPlayer()

		assert.Equal(t, OmahaPhaseEnd, o.phase)
		assert.True(t, o.gameEndFlag)
		assert.Len(t, o.roundResults, 1)
		assert.Equal(t, 0, o.roundResults[0].PlayerIdx)
		assert.Equal(t, 200, o.roundResults[0].WonAmount)
	})
}

func TestOmaha_DealRemainingCommunity(t *testing.T) {
	t.Run("deals remaining cards to reach 5", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.trumpCards.Shuffle()
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		o.dealRemainingCommunity()
		assert.Equal(t, 5, len(o.communityCards))
	})

	t.Run("does nothing when already 5 cards", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignSpade, 6, false),
		}
		o.dealRemainingCommunity()
		assert.Equal(t, 5, len(o.communityCards))
	})
}

func TestOmaha_CpuDecidePreFlop_PassiveCallBranch(t *testing.T) {
	// Cover the Passive (non-aggressive) callAmount>0 branch
	t.Run("passive style with callAmount>0 calls", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 10
		o.pot = 30
		o.minRaise = 10
		// LAP is not aggressive (aggressive=false)
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		// Give a weak hand to avoid fold, but below raise threshold
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignHeart, 9, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		p.AddCard(NewCard(CardDesignDiamond, 6, false))
		o.players[1] = p

		hitCall := false
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecidePreFlop(1, holdemStyleParamsMap[HoldemStyleLAP], 10)
			if action == OmahaActionCall {
				hitCall = true
				break
			}
		}
		assert.True(t, hitCall)
	})

	t.Run("passive style bluff branch with callAmount=0", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 0
		o.pot = 30
		o.minRaise = 10
		// LAP passive with no callAmount — bluff branch
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignHeart, 9, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		p.AddCard(NewCard(CardDesignDiamond, 6, false))
		o.players[1] = p

		hitBet := false
		hitCheck := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePreFlop(1, holdemStyleParamsMap[HoldemStyleLAP], 0)
			if action == OmahaActionBet || action == OmahaActionAllIn {
				hitBet = true
			}
			if action == OmahaActionCheck {
				hitCheck = true
			}
			if hitBet && hitCheck {
				break
			}
		}
		assert.True(t, hitBet, "should hit bluff bet branch")
		assert.True(t, hitCheck, "should hit check branch")
	})
}

func TestOmaha_CpuPotBet_MinRaiseFallback(t *testing.T) {
	t.Run("minRaise > bet forces bet up to minRaise", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.pot = 10
		o.minRaise = 50
		o.config.BigBlind = 10
		bet := o.cpuPotBet(10) // 10% of 10 = 1, below BigBlind (10) and minRaise (50)
		assert.Equal(t, 50, bet)
	})
}

func TestOmaha_PostBlinds_AllIn(t *testing.T) {
	t.Run("SB all-in on blind", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.dealerIdx = 0
		o.actedFlags = make([]bool, 4)
		o.players[1].SetChips(3) // SB can't afford full blind
		o.postBlinds()
		assert.Equal(t, 0, o.players[1].GetChips())
		assert.True(t, o.players[1].GetAllIn())
		assert.True(t, o.actedFlags[1])
	})

	t.Run("BB all-in on blind", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.dealerIdx = 0
		o.actedFlags = make([]bool, 4)
		o.players[2].SetChips(7) // BB can't afford full blind
		o.postBlinds()
		assert.Equal(t, 0, o.players[2].GetChips())
		assert.True(t, o.players[2].GetAllIn())
		assert.True(t, o.actedFlags[2])
	})
}

func TestOmaha_IsRebuyAvailable_Positive(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyPeriodHands = 20
	o.rebuyCounts = make([]int, 4)
	o.handCount = 5
	o.players[0].SetChips(0)

	assert.True(t, o.IsRebuyAvailable())
}

func TestOmaha_IsAddonAvailable_Positive(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.AddonEnabled = true
	o.config.AddonAfterHand = 5
	o.handCount = 5
	o.addonUsed = make([]bool, 4)

	assert.True(t, o.IsAddonAvailable())
}

func TestOmaha_SkipRebuy_NonBustHuman(t *testing.T) {
	t.Run("human has chips, skip rebuy continues reset", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.config.RebuyEnabled = true
		o.config.RebuyMaxCount = 3
		o.config.RebuyChips = 1000
		o.config.RebuyPeriodHands = 20
		o.rebuyCounts = make([]int, 4)
		o.handCount = 0

		// Make CPU bust, not human
		o.players[1].SetChips(0)
		o.players[0].SetChips(1000)

		err := o.Reset()
		assert.NoError(t, err)

		// If human has chips and phase is rebuy (due to CPU), SkipRebuy shouldn't end game
		// But actually rebuy phase only triggers for human bust. So set up directly.
		o.phase = OmahaPhaseRebuy
		o.rebuyPhaseType = OmahaRebuyPhaseRebuy
		o.players[0].SetChips(1000) // human has chips

		err = o.SkipRebuy()
		assert.NoError(t, err)
		// Human has chips, so should not set gameEndFlag; should continue to reset
		assert.NotEqual(t, OmahaPhaseRebuy, o.phase)
	})
}

func TestOmaha_AdvanceTurn_GameEndFlag(t *testing.T) {
	o := newInternalTestOmaha()
	o.gameEndFlag = true
	o.currentTurn = 1
	o.advanceTurn()
	assert.Equal(t, 1, o.currentTurn) // unchanged
}

func TestOmaha_AdvanceTurn_AllActedFallback(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.currentTurn = 0
	o.trumpCards.Shuffle()
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	// All players have acted (or folded/all-in) so the loop finds nobody
	o.actedFlags = []bool{true, true, true, true}
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}

	o.advanceTurn()
	// Should advance phase since betting round is complete
	assert.NotEqual(t, OmahaPhaseFlop, o.phase)
}

func TestOmaha_AdvancePhase_TurnToRiver(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseTurn
	o.trumpCards.Shuffle()
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
	}
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}
	o.actedFlags = make([]bool, 4)

	o.advancePhase()
	assert.Equal(t, OmahaPhaseRiver, o.phase)
	assert.Equal(t, 5, len(o.communityCards))
}

func TestOmaha_AdvancePhase_RiverToShowdown(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseRiver
	o.pot = 200
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 6, false),
	}
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}
	o.actedFlags = make([]bool, 4)

	o.advancePhase()
	// River → Showdown → resolveShowdown → end
	assert.True(t, o.phase == OmahaPhaseShowdown || o.phase == OmahaPhaseEnd)
}

func TestOmaha_AdvancePhase_AllInAfterFlop(t *testing.T) {
	// Test: after advancing from preflop to flop, all players all-in → dealRemainingCommunity + showdown
	o := newInternalTestOmaha()
	o.phase = OmahaPhasePreFlop
	o.pot = 400
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.trumpCards.Shuffle()

	for _, p := range o.players {
		p.Reset()
		p.SetAllIn(true)
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}
	o.actedFlags = make([]bool, 4)

	o.advancePhase()
	// All players all-in: should deal remaining community cards and go to showdown
	assert.Equal(t, 5, len(o.communityCards))
	assert.True(t, o.phase == OmahaPhaseShowdown || o.phase == OmahaPhaseEnd)
}

func TestOmaha_FindNextActive_NoActive(t *testing.T) {
	o := newInternalTestOmaha()
	for _, p := range o.players {
		p.SetFolded(true)
	}
	// All folded → fallback returns (fromIdx+1)%len
	result := o.findNextActive(0)
	assert.Equal(t, 1, result)
}

func TestOmaha_IsHumanTurn_OutOfRange(t *testing.T) {
	o := newInternalTestOmaha()
	o.currentTurn = -1
	assert.False(t, o.IsHumanTurn())

	o.currentTurn = 100
	assert.False(t, o.IsHumanTurn())
}

func TestOmaha_RunCpuActions_GameEndFlag(t *testing.T) {
	o := newInternalTestOmaha()
	o.gameEndFlag = true
	o.phase = OmahaPhaseFlop
	err := o.runCpuActions()
	assert.NoError(t, err)
}

func TestOmaha_CpuDecide_GTOPostFlop_PotLimitCapping(t *testing.T) {
	t.Run("pot limit caps amount", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.config.BettingLimit = BettingLimitPotLimit
		o.lastBet = 0
		o.pot = 50
		o.minRaise = 10
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		gtoPlayer := NewOmahaPlayer(false, HoldemStyleGTO)
		gtoPlayer.SetChips(10000)
		gtoPlayer.AddCard(NewCard(CardDesignSpade, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignHeart, 1, false))
		gtoPlayer.AddCard(NewCard(CardDesignClover, 13, false))
		gtoPlayer.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = gtoPlayer

		// Run multiple times to hit a bet that gets capped
		for i := 0; i < 100; i++ {
			action, amount := o.cpuDecide(1)
			if action == OmahaActionBet || action == OmahaActionRaise {
				_, maxBetAmount := o.bettingLimits()
				assert.LessOrEqual(t, amount, maxBetAmount)
			}
		}
	})
}

func TestOmaha_CpuDecide_NonGTO_RaiseCapCheckFallback(t *testing.T) {
	t.Run("non-GTO raise cap with callAmount=0 falls back to check", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.config.BettingLimit = BettingLimitFixed
		o.raiseCount = 4
		o.lastBet = 0 // callAmount=0
		o.pot = 200
		o.minRaise = 10
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		// Use TAG (aggressive) with strong hand to attempt bet
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 1, false))
		p.AddCard(NewCard(CardDesignClover, 13, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = p

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
	})
}

func TestOmaha_CpuDecide_NonGTO_RaiseCapCallFallback(t *testing.T) {
	t.Run("non-GTO raise cap with callAmount>0 falls back to call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.config.BettingLimit = BettingLimitFixed
		o.raiseCount = 4
		o.lastBet = 40
		o.pot = 200
		o.minRaise = 10
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		// TAG aggressive with strong hand
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 1, false))
		p.AddCard(NewCard(CardDesignClover, 13, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = p

		hitCall := false
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecide(1)
			if action == OmahaActionCall {
				hitCall = true
			}
			assert.NotEqual(t, OmahaActionRaise, action)
			assert.NotEqual(t, OmahaActionBet, action)
		}
		assert.True(t, hitCall)
	})
}

func TestOmaha_CpuDecide_NonGTO_MaxBetAmountCapping(t *testing.T) {
	t.Run("pot limit caps non-GTO bet amount when minRaise exceeds pot", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.config.BettingLimit = BettingLimitPotLimit
		o.lastBet = 0
		o.pot = 10       // Small pot → small maxBetAmount
		o.minRaise = 200 // Large minRaise forces cpuPotBet to return > maxBetAmount
		o.config.BigBlind = 10
		o.communityCards = []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false),
		}
		// TAG aggressive with strong hand → will try to bet/raise
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(10000)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 1, false))
		p.AddCard(NewCard(CardDesignClover, 13, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = p

		hitCapped := false
		for i := 0; i < 1000; i++ {
			action, amount := o.cpuDecide(1)
			_, maxBetAmount := o.bettingLimits()
			if (action == OmahaActionBet || action == OmahaActionRaise) && maxBetAmount > 0 {
				assert.LessOrEqual(t, amount, maxBetAmount)
				hitCapped = true
			}
		}
		assert.True(t, hitCapped, "should hit the maxBetAmount capping branch")
	})
}

func TestOmaha_PlayerAction_Errors(t *testing.T) {
	t.Run("game ended error", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.gameEndFlag = true
		o.phase = OmahaPhaseFlop
		err := o.PlayerAction(OmahaActionCheck, 0, 0)
		assert.Error(t, err)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseShowdown
		err := o.PlayerAction(OmahaActionCheck, 0, 0)
		assert.Error(t, err)
	})

	t.Run("not human turn error", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.currentTurn = 1 // CPU player
		err := o.PlayerAction(OmahaActionCheck, 0, 0)
		assert.Error(t, err)
	})
}

func TestOmaha_PlayerAction_ExecuteActionError(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.lastBet = 0
	o.pot = 50
	o.minRaise = 10
	o.actedFlags = make([]bool, 4)
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.currentTurn = 0 // human
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

	// Invalid action: raise with 0 bet (should error)
	err := o.PlayerAction(OmahaActionRaise, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_ResolveShowdown_HumanLost(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseShowdown
	o.pot = 200
	o.startingChips = []int{1000, 1000, 1000, 1000}
	// Community: K, Q, J, 10, 9 — all different suits to avoid flush
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 9, false),
	}
	// Human gets weak cards (2, 3 → best is pair of nothing, just high cards from community)
	o.players[0].Reset()
	o.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	o.players[0].AddCard(NewCard(CardDesignClover, 3, false))
	o.players[0].AddCard(NewCard(CardDesignDiamond, 5, false))
	o.players[0].AddCard(NewCard(CardDesignSpade, 4, false))

	// CPU gets A + K → can make A-K-Q-J-10 straight (using A and any one card from hand + 3 community)
	o.players[1].Reset()
	o.players[1].AddCard(NewCard(CardDesignHeart, 1, false))
	o.players[1].AddCard(NewCard(CardDesignClover, 8, false))
	o.players[1].AddCard(NewCard(CardDesignDiamond, 7, false))
	o.players[1].AddCard(NewCard(CardDesignSpade, 6, false))

	for i := 2; i < 4; i++ {
		o.players[i].SetFolded(true)
	}

	o.resolveShowdown()
	// humanLost=true means finalizeShowdown not called, phase stays at showdown
	assert.Equal(t, OmahaPhaseShowdown, o.phase)
	assert.False(t, o.gameEndFlag)
}

func TestOmaha_RunCpuActions_FoldedSkip(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.lastBet = 0
	o.pot = 50
	o.minRaise = 10
	o.actedFlags = make([]bool, 4)
	o.actedFlags[0] = true // human acted
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
	// CPU 1 is folded, CPU 2 is all-in — should skip them
	o.players[1].SetFolded(true)
	o.players[2].SetAllIn(true)
	o.actedFlags[1] = true
	o.actedFlags[2] = true
	o.currentTurn = 1

	err := o.runCpuActions()
	assert.NoError(t, err)
}

func TestOmaha_AdvanceTurn_FindNextUndone(t *testing.T) {
	// Cover the inner loop in advanceTurn that skips folded/allin/acted players
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.currentTurn = 0
	// Player 0: acted. Player 1: folded. Player 2: not acted. Player 3: all-in
	o.actedFlags = []bool{true, true, false, true}
	o.players[1].SetFolded(true)
	o.players[3].SetAllIn(true)

	o.advanceTurn()
	assert.Equal(t, 2, o.currentTurn)
}

func TestOmaha_CpuDecidePreFlop_AllStyles_Detailed(t *testing.T) {
	// TAG aggressive: strong hand → raise, weak hand → fold
	t.Run("TAG aggressive fold with weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 20
		o.pot = 50
		o.minRaise = 10
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		// Very weak hand (strength < 40)
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignHeart, 5, false))
		p.AddCard(NewCard(CardDesignClover, 8, false))
		p.AddCard(NewCard(CardDesignDiamond, 4, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAG]

		hitFold := false
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecidePreFlop(1, params, 20)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "TAG should fold weak hand")
	})

	t.Run("TAG aggressive raise with strong hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 10
		o.pot = 30
		o.minRaise = 10
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		// Very strong hand (strength >= 70)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 1, false))
		p.AddCard(NewCard(CardDesignClover, 13, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAG]

		hitRaise := false
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecidePreFlop(1, params, 10)
			if action == OmahaActionRaise || action == OmahaActionBet {
				hitRaise = true
				break
			}
		}
		assert.True(t, hitRaise, "TAG should raise strong hand")
	})

	t.Run("TAG aggressive bluff branch", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 10
		o.pot = 30
		o.minRaise = 10
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		// Medium hand (between fold and raise thresholds)
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 9, false))
		p.AddCard(NewCard(CardDesignDiamond, 8, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAG]

		hitRaise := false
		hitCall := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePreFlop(1, params, 10)
			if action == OmahaActionRaise || action == OmahaActionBet {
				hitRaise = true
			}
			if action == OmahaActionCall || action == OmahaActionCheck {
				hitCall = true
			}
			if hitRaise && hitCall {
				break
			}
		}
		assert.True(t, hitRaise, "TAG should sometimes bluff-raise medium hand")
		assert.True(t, hitCall, "TAG should sometimes call medium hand")
	})

	// LAP passive: compound fold with high callAmount
	t.Run("LAP compound fold", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 30
		o.pot = 50
		o.minRaise = 10
		o.config.BigBlind = 10
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		// Very weak hand: vals=[2,5,9,12(Q)] all different suits, no connectors near each other
		// score = 12+9=21 high cards. Gaps too large for connectors. No pairs/suited. <15? No.
		// Need truly trash: 2,5,8,11(J) → sorted=[11,8,5,2] → 11+8=19. No pairs/suits. 11-8=3, 8-5=3, 5-2=3 → no connectors. 11-2=9>4 no wrap. No high cards (<12). = 19. Still >15.
		// Use the most spread-out hand possible, all different suits, low values
		// 2,4,8,11 → sorted=[11,8,4,2] → 11+8=19. Too high.
		// The threshold for LAP is 15. We need score < 15. Use 2,5,9,3 → sorted=[9,5,3,2] → 9+5=14. Connectors: 3-2=1 (+4) → 18. Too high.
		// Try 2,6,10,4 → sorted=[10,6,4,2] → 10+6=16. Connector: 4-2=2 (+2)=18. Too high.
		// For Omaha pre-flop strength < 15, we need very low top cards and no connectors.
		// 2,5,8,3 3 same suit = penalty: score = 8+5=13, 3-2=1(+4), 5-3=2(+2) → 19. With 3 same suit: -5 = 14. Yes!
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAP]

		// callAmount=30 > BB*2=20, so compound fold should trigger
		action, _ := o.cpuDecidePreFlop(1, params, 30)
		assert.Equal(t, OmahaActionFold, action)
	})

	t.Run("LAP compound no fold when callAmount small", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 15
		o.pot = 30
		o.minRaise = 10
		o.config.BigBlind = 10
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		// Same weak hand
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAP]

		// callAmount=15 <= BB*2=20, so should call instead of fold
		action, _ := o.cpuDecidePreFlop(1, params, 15)
		assert.Equal(t, OmahaActionCall, action)
	})

	// TAP passive: non-compound fold
	t.Run("TAP passive fold weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 20
		o.pot = 50
		o.minRaise = 10
		p := NewOmahaPlayer(false, HoldemStyleTAP)
		p.SetChips(1000)
		// Very weak hand (strength < 30)
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignHeart, 4, false))
		p.AddCard(NewCard(CardDesignClover, 7, false))
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAP]

		hitFold := false
		for i := 0; i < 100; i++ {
			action, _ := o.cpuDecidePreFlop(1, params, 20)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "TAP should fold weak hand with callAmount")
	})

	// LAG aggressive: compound fold and bluff raise
	t.Run("LAG compound fold with high callAmount", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhasePreFlop
		o.lastBet = 40
		o.pot = 80
		o.minRaise = 10
		o.config.BigBlind = 10
		p := NewOmahaPlayer(false, HoldemStyleLAG)
		p.SetChips(1000)
		// Very weak hand (strength < 15): 3 same suit penalty
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAG]

		// callAmount=40 > BB*3=30, compound fold
		action, _ := o.cpuDecidePreFlop(1, params, 40)
		assert.Equal(t, OmahaActionFold, action)
	})
}

func TestOmaha_CpuDecidePostFlop_AllStyles_Detailed(t *testing.T) {
	community := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}

	// TAG aggressive: fallbackFold=true, condCallRank=OnePair
	t.Run("TAG fallback fold with weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 20
		o.pot = 100
		o.minRaise = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		// Weak hand: high card only
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAG]

		hitFold := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 20)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "TAG should fold weak hand postflop")
	})

	t.Run("TAG fallback call with pair", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 20
		o.pot = 100
		o.minRaise = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.SetChips(1000)
		// Give a pair: 7 matches community 7
		p.AddCard(NewCard(CardDesignDiamond, 7, false))
		p.AddCard(NewCard(CardDesignSpade, 3, false))
		p.AddCard(NewCard(CardDesignHeart, 5, false))
		p.AddCard(NewCard(CardDesignClover, 8, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAG]

		hitCall := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 20)
			if action == OmahaActionCall {
				hitCall = true
				break
			}
		}
		assert.True(t, hitCall, "TAG should call with pair")
	})

	// LAG aggressive: fallbackFold=false, postFlopAggrFoldRank=HighCard, Mult=4
	t.Run("LAG aggressive fold on high call with weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 50
		o.pot = 200
		o.minRaise = 10
		o.config.BigBlind = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAG)
		p.SetChips(1000)
		// Weak hand
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAG]

		hitFold := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 50)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "LAG should fold weak hand with high call")
	})

	t.Run("LAG aggressive call with moderate call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 20
		o.pot = 100
		o.minRaise = 10
		o.config.BigBlind = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAG)
		p.SetChips(1000)
		// Weak hand, but call amount <= BB*4
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAG]

		hitCall := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 20)
			if action == OmahaActionCall {
				hitCall = true
				break
			}
		}
		assert.True(t, hitCall, "LAG should call with moderate bet")
	})

	t.Run("LAG aggressive check with no call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 0
		o.pot = 100
		o.minRaise = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAG)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAG]

		hitCheck := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 0)
			if action == OmahaActionCheck {
				hitCheck = true
				break
			}
		}
		assert.True(t, hitCheck, "LAG should sometimes check")
	})

	// LAP passive: postFlopPassFoldRank=HighCard, Mult=3
	t.Run("LAP passive fold with weak hand and high call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 40
		o.pot = 100
		o.minRaise = 10
		o.config.BigBlind = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAP]

		// callAmount=40 > BB*3=30 → fold
		hitFold := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 40)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "LAP should fold weak hand with high call")
	})

	t.Run("LAP passive call with moderate call", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 20
		o.pot = 100
		o.minRaise = 10
		o.config.BigBlind = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAP]

		// callAmount=20 <= BB*3=30 → call
		action, _ := o.cpuDecidePostFlop(1, params, 20)
		assert.Equal(t, OmahaActionCall, action)
	})

	t.Run("LAP passive bluff", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 0
		o.pot = 100
		o.minRaise = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleLAP)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleLAP]

		hitBet := false
		hitCheck := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 0)
			if action == OmahaActionBet || action == OmahaActionAllIn {
				hitBet = true
			}
			if action == OmahaActionCheck {
				hitCheck = true
			}
			if hitBet && hitCheck {
				break
			}
		}
		assert.True(t, hitBet, "LAP should sometimes bluff")
		assert.True(t, hitCheck, "LAP should sometimes check")
	})

	// TAP passive: postFlopPassFoldMult=-1 (always fold weak hand with any call)
	t.Run("TAP passive fold with any call and weak hand", func(t *testing.T) {
		o := newInternalTestOmaha()
		o.phase = OmahaPhaseFlop
		o.lastBet = 10
		o.pot = 50
		o.minRaise = 10
		o.communityCards = community
		p := NewOmahaPlayer(false, HoldemStyleTAP)
		p.SetChips(1000)
		p.AddCard(NewCard(CardDesignDiamond, 3, false))
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 8, false))
		p.AddCard(NewCard(CardDesignClover, 10, false))
		o.players[1] = p
		params := holdemStyleParamsMap[HoldemStyleTAP]

		// postFlopPassFoldMult=-1 → always fold with HighCard regardless of callAmount
		hitFold := false
		for i := 0; i < 1000; i++ {
			action, _ := o.cpuDecidePostFlop(1, params, 10)
			if action == OmahaActionFold {
				hitFold = true
				break
			}
		}
		assert.True(t, hitFold, "TAP should fold weak hand with any call")
	})
}

func TestOmaha_Reset_BlindMinimumClamp(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.TournamentMode = true
	o.config.BlindLevelHands = 1
	o.config.BlindMultiplier = 10 // Very small multiplier to trigger minimum clamp
	o.config.SmallBlind = 1
	o.config.BigBlind = 2
	o.handCount = 1 // Will escalate on next Reset

	_ = o.Reset()
	// After escalation: SB = 1*10/100 = 0 → clamped to 1, BB = 2*10/100 = 0 → clamped to 2
	assert.GreaterOrEqual(t, o.config.SmallBlind, 1)
	assert.GreaterOrEqual(t, o.config.BigBlind, 2)
}

func TestOmaha_IsRebuyAvailable_NoBust(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyPeriodHands = 20
	o.rebuyCounts = make([]int, 4)
	o.handCount = 5
	// Human has chips → not eligible
	o.players[0].SetChips(100)
	assert.False(t, o.IsRebuyAvailable())
}

func TestOmaha_IsAddonAvailable_AlreadyUsed(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.AddonEnabled = true
	o.config.AddonAfterHand = 5
	o.handCount = 5
	o.addonUsed = []bool{true, true, true, true}
	// Human already used addon
	assert.False(t, o.IsAddonAvailable())
}

func TestOmaha_Rebuy_CpuAutoRebuy(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20
	o.handCount = 0
	o.rebuyCounts = make([]int, 4)

	// Make both human and CPU bust
	o.players[0].SetChips(0)
	o.players[1].SetChips(0)

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, OmahaPhaseRebuy, o.phase)

	// CPU should have been auto-rebuyed
	assert.True(t, o.players[1].GetChips() > 0)

	// Do rebuy for human
	err = o.Rebuy()
	assert.NoError(t, err)
	assert.True(t, o.players[0].GetChips() > 0)
}

func TestOmaha_SkipRebuy_ThenAddon(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20
	o.config.AddonEnabled = true
	o.config.AddonChips = 500
	// handCount will become 1 after Reset, set AddonAfterHand=1
	o.config.AddonAfterHand = 1
	o.handCount = 0
	o.rebuyCounts = make([]int, 4)
	o.addonUsed = make([]bool, 4)

	// Human has chips, no rebuy needed
	for _, p := range o.players {
		p.SetChips(1000)
	}

	err := o.Reset()
	assert.NoError(t, err)
	// Should go to addon phase since handCount matches AddonAfterHand
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseAddon, o.rebuyPhaseType)
}

func TestOmaha_Rebuy_ThenAddon(t *testing.T) {
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20
	o.config.AddonEnabled = true
	o.config.AddonChips = 500
	o.config.AddonAfterHand = 1
	o.handCount = 0
	o.rebuyCounts = make([]int, 4)
	o.addonUsed = make([]bool, 4)

	// Human bust
	o.players[0].SetChips(0)

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseRebuy, o.rebuyPhaseType)

	// Rebuy, then should transition to addon
	err = o.Rebuy()
	assert.NoError(t, err)
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseAddon, o.rebuyPhaseType)
}

func TestOmaha_DealRemainingCommunity_EmptyDeck(t *testing.T) {
	o := newInternalTestOmaha()
	// Empty the deck
	o.trumpCards = NewTrumpCards(0)
	for o.trumpCards.DrawCard() != nil {
	}
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
	}
	o.dealRemainingCommunity()
	// Should stop when deck is empty, not reach 5
	assert.Equal(t, 1, len(o.communityCards))
}

func TestOmaha_ExecuteAction_TriggerResolveLastPlayer(t *testing.T) {
	// Test that executeAction calls resolveLastPlayer when only 1 player left
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.lastBet = 20
	o.pot = 100
	o.minRaise = 10
	o.actedFlags = make([]bool, 4)
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
	// All but player 0 and player 1 are folded
	o.players[2].SetFolded(true)
	o.players[3].SetFolded(true)

	// Player 1 folds → only player 0 left → resolveLastPlayer triggered
	err := o.executeAction(1, OmahaActionFold, 0)
	assert.NoError(t, err)
	assert.True(t, o.gameEndFlag)
	assert.Equal(t, OmahaPhaseEnd, o.phase)
}

func TestOmaha_AdvanceTurn_AllActedAllInFallback(t *testing.T) {
	// Cover the fallback in advanceTurn where the loop finds no unacted active player
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.currentTurn = 0
	o.trumpCards.Shuffle()
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 10, false))
		p.AddCard(NewCard(CardDesignHeart, 11, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 13, false))
	}
	// Player 0 acted, players 1-3 are all-in (not folded, not needing to act)
	o.actedFlags = []bool{true, true, true, true}
	o.players[1].SetAllIn(true)
	o.players[2].SetAllIn(true)
	o.players[3].SetAllIn(true)

	// isBettingRoundComplete() returns true → advancePhase is called via first path
	o.advanceTurn()
	assert.NotEqual(t, OmahaPhaseFlop, o.phase)
}

func TestOmaha_SkipRebuy_AddonTransition(t *testing.T) {
	// Test SkipRebuy with human having chips + addon available
	o := newInternalTestOmaha()
	o.config.RebuyEnabled = true
	o.config.RebuyMaxCount = 3
	o.config.RebuyChips = 1000
	o.config.RebuyPeriodHands = 20
	o.config.AddonEnabled = true
	o.config.AddonChips = 500
	o.config.AddonAfterHand = 1
	o.handCount = 1 // matches AddonAfterHand
	o.rebuyCounts = make([]int, 4)
	o.addonUsed = make([]bool, 4)

	// Set up rebuy phase but human has chips
	o.phase = OmahaPhaseRebuy
	o.rebuyPhaseType = OmahaRebuyPhaseRebuy
	o.players[0].SetChips(1000)

	err := o.SkipRebuy()
	assert.NoError(t, err)
	// Should transition to addon
	assert.Equal(t, OmahaPhaseRebuy, o.phase)
	assert.Equal(t, OmahaRebuyPhaseAddon, o.rebuyPhaseType)
}

func TestOmaha_RunCpuActions_SkipFoldedAndAllIn(t *testing.T) {
	o := newInternalTestOmaha()
	o.phase = OmahaPhaseFlop
	o.lastBet = 0
	o.pot = 50
	o.minRaise = 10
	o.actedFlags = make([]bool, 4)
	o.actedFlags[0] = true // human acted
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
	// Player 1: folded, Player 2: all-in, Player 3: needs to act
	o.players[1].SetFolded(true)
	o.actedFlags[1] = true
	o.players[2].SetAllIn(true)
	o.actedFlags[2] = true
	o.currentTurn = 1 // Start at folded player

	err := o.runCpuActions()
	assert.NoError(t, err)
}

func TestOmaha_EvalPreFlopStrength_Gap2Connector(t *testing.T) {
	o := newInternalTestOmaha()
	o.players[0].Reset()
	// Cards with gap=2: 10 and 8 (gap=2)
	o.players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	o.players[0].AddCard(NewCard(CardDesignHeart, 8, false))
	o.players[0].AddCard(NewCard(CardDesignClover, 3, false))
	o.players[0].AddCard(NewCard(CardDesignDiamond, 13, false))
	strength := o.evalPreFlopStrength(0)
	assert.Greater(t, strength, 0)
}

func TestOmaha_ContinueReset_NilCard(t *testing.T) {
	// Test continueReset when deck runs out of cards during deal
	o := newInternalTestOmaha()
	// Drain most of the deck
	o.trumpCards = NewTrumpCards(0)
	o.trumpCards.Shuffle()
	// Draw 48 cards to leave only 4 in deck (not enough for 4 players * 4 cards = 16)
	for i := 0; i < 48; i++ {
		o.trumpCards.DrawCard()
	}
	o.actedFlags = make([]bool, 4)

	err := o.continueReset()
	// Should not panic even with insufficient cards
	assert.NoError(t, err)
}
