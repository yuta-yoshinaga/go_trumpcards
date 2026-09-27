package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestOmaha() *Omaha {
	players := []*OmahaPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleLAP),
		NewOmahaPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultOmahaConfig()
	tc := NewTrumpCards(0)
	return NewOmaha(tc, players, cfg)
}

func setupOmahaForHumanAction(phase int) *Omaha {
	o := newTestOmaha()
	for _, p := range o.players {
		p.SetChips(1000)
	}
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.SetPhase(phase)
	o.SetCurrentTurn(0)
	o.SetLastBet(0)
	o.SetMinRaise(10)
	o.SetPot(30)
	o.actedFlags = []bool{false, true, true, true}
	// Give players 4 cards each
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 13, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 11, false))
	}
	return o
}

func TestNewOmaha(t *testing.T) {
	o := newTestOmaha()
	assert.Equal(t, OmahaPhaseInit, o.GetPhase())
	assert.Equal(t, 4, o.GetPlayerCnt())
	assert.NotNil(t, o.GetCommunityCards())
	assert.Equal(t, 0, o.GetPot())
	assert.False(t, o.GetGameEndFlag())
}

func TestOmahaDistributeAmongWinnersRecordsChipPayouts(t *testing.T) {
	players := []BettingPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleTAG),
	}
	won := make(map[int]int)
	payouts := make([]int, 2)
	distributeAmongWinners(players, []int{1, 0}, 11, won, payouts)
	assert.Equal(t, []int{6, 5}, payouts, "the first winner must receive the odd chip")
	assert.Equal(t, 6, won[1])
	assert.Equal(t, 5, won[0])
	assert.Equal(t, 6, players[1].GetChips())
	assert.Equal(t, 5, players[0].GetChips())
}

func TestOmahaPotAwardsSnapshotAndReset(t *testing.T) {
	o := newTestOmaha()
	o.potAwards = []OmahaPotAward{{Amount: 12, Eligible: []int{0, 1}, HiWinners: []int{1}, HiPayouts: []int{12}}}
	data, err := o.MarshalJSON()
	assert.NoError(t, err)
	var restored Omaha
	assert.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, o.potAwards, restored.GetPotAwards())
	assert.NoError(t, o.Reset())
	assert.Nil(t, o.GetPotAwards())
}

func TestOmahaUnmarshalPotAwardSliceLimits(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"awards", fmt.Sprintf(`{"pa":[%s]}`, strings.TrimSuffix(strings.Repeat(`{},`, omahaMaxSliceLen+1), ","))},
		{"eligible", `{"pa":[{"eligible":[` + oversizedJSONElements() + `]}]}`},
		{"high winners", `{"pa":[{"hiWinners":[` + oversizedJSONElements() + `]}]}`},
		{"high payouts", `{"pa":[{"hiPayouts":[` + oversizedJSONElements() + `]}]}`},
		{"low winners", `{"pa":[{"loWinners":[` + oversizedJSONElements() + `]}]}`},
		{"low payouts", `{"pa":[{"loPayouts":[` + oversizedJSONElements() + `]}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var o Omaha
			err := json.Unmarshal([]byte(tt.json), &o)
			assert.EqualError(t, err, "omaha: input array exceeds maximum allowed size")
		})
	}
}

func oversizedJSONElements() string {
	return strings.TrimSuffix(strings.Repeat(`0,`, omahaMaxSliceLen+1), ",")
}

func TestOmahaResolveShowdownPotAwards(t *testing.T) {
	for _, hiLo := range []bool{false, true} {
		t.Run(fmt.Sprintf("hiLo=%t", hiLo), func(t *testing.T) {
			players := []*OmahaPlayer{
				NewOmahaPlayer(true, HoldemStyleTAG),
				NewOmahaPlayer(false, HoldemStyleTAG),
				NewOmahaPlayer(false, HoldemStyleTAG),
			}
			cfg := DefaultOmahaConfig()
			o := NewOmaha(NewTrumpCards(0), players, cfg)
			if hiLo {
				o.hiLo = true
			}
			// The fixed holdings and board make each showdown deterministic;
			// seats 0 and 1 share the same qualifying low.
			holes := [][]*Card{
				{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 6, false), NewCard(CardDesignHeart, 12, false), NewCard(CardDesignHeart, 11, false)},
				{NewCard(CardDesignClover, 1, false), NewCard(CardDesignClover, 7, false), NewCard(CardDesignDiamond, 12, false), NewCard(CardDesignDiamond, 11, false)},
				{NewCard(CardDesignHeart, 10, false), NewCard(CardDesignDiamond, 10, false), NewCard(CardDesignHeart, 9, false), NewCard(CardDesignDiamond, 9, false)},
			}
			remaining := []int{400, 200, 0}
			for i, p := range players {
				p.SetChips(remaining[i])
				p.SetAllIn(true)
				for _, c := range holes[i] {
					p.AddCard(c)
				}
			}
			o.startingChips = []int{500, 500, 500}
			o.pot = 900
			o.communityCards = []*Card{
				NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignClover, 4, false), NewCard(CardDesignDiamond, 5, false),
				NewCard(CardDesignSpade, 9, false),
			}
			o.resolveShowdown()
			awards := o.GetPotAwards()
			assert.Len(t, awards, 3)
			assert.Equal(t, []int{300, 400, 200}, []int{awards[0].Amount, awards[1].Amount, awards[2].Amount})
			assert.Equal(t, []int{0, 1, 2}, awards[0].Eligible)
			assert.Equal(t, []int{1, 2}, awards[1].Eligible)
			assert.Equal(t, []int{2}, awards[2].Eligible)
			if !hiLo {
				assert.Equal(t, []int{2}, awards[0].HiWinners)
				assert.Equal(t, []int{300}, awards[0].HiPayouts)
				assert.Equal(t, []int{2}, awards[1].HiWinners)
				assert.Equal(t, []int{400}, awards[1].HiPayouts)
				assert.Equal(t, []int{2}, awards[2].HiWinners)
				assert.Equal(t, []int{200}, awards[2].HiPayouts)
			} else {
				assert.Equal(t, []int{2}, awards[0].HiWinners)
				assert.Equal(t, []int{150}, awards[0].HiPayouts)
				assert.Equal(t, []int{2}, awards[1].HiWinners)
				assert.Equal(t, []int{200}, awards[1].HiPayouts)
				assert.Equal(t, []int{2}, awards[2].HiWinners)
				assert.Equal(t, []int{200}, awards[2].HiPayouts)
				assert.Equal(t, []int{0}, awards[0].LoWinners)
				assert.Equal(t, []int{150}, awards[0].LoPayouts)
				assert.Equal(t, []int{1}, awards[1].LoWinners)
				assert.Equal(t, []int{200}, awards[1].LoPayouts)
				assert.Empty(t, awards[2].LoWinners)
				assert.Empty(t, awards[2].LoPayouts)
			}
			paid := make([]int, len(players))
			for _, award := range awards {
				for n, idx := range award.HiWinners {
					paid[idx] += award.HiPayouts[n]
				}
				for n, idx := range award.LoWinners {
					paid[idx] += award.LoPayouts[n]
				}
			}
			for _, result := range o.GetRoundResults() {
				assert.Equal(t, result.WonAmount, paid[result.PlayerIdx], "seat %d", result.PlayerIdx)
			}
		})
	}
}

func TestOmaha_Reset(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()

	// After reset, each player should have 4 cards (Omaha)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		p := o.GetPlayer(i)
		assert.Equal(t, 4, p.GetCardsSize(), "player %d should have 4 cards", i)
		assert.True(t, p.GetChips() > 0 || p.GetAllIn())
	}
	assert.True(t, o.GetPot() > 0)
	blindCodes := map[string]bool{}
	for _, entry := range o.GetActionLog() {
		if entry.ActionType == "blind" {
			blindCodes[entry.DetailCode] = true
			assert.Equal(t, map[string]string{"amount": entry.DetailParams["amount"]}, entry.DetailParams)
		}
	}
	assert.True(t, blindCodes["omaha.log.smallBlind"])
	assert.True(t, blindCodes["omaha.log.bigBlind"])
}

func TestOmaha_Reset_Deals4Cards(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, 4, o.GetPlayer(i).GetCardsSize())
	}
}

func TestOmaha_Resize(t *testing.T) {
	o := newTestOmaha()
	assert.Equal(t, 4, o.GetPlayerCnt())

	newPlayers := make([]*OmahaPlayer, 6)
	newPlayers[0] = NewOmahaPlayer(true, HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = NewOmahaPlayer(false, HoldemStyleLAP)
	}
	o.Resize(newPlayers)
	assert.Equal(t, 6, o.GetPlayerCnt())
	assert.Equal(t, 0, o.GetHandCount())
	assert.Equal(t, 6, len(o.GetActedFlags()))

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 6, o.GetPlayerCnt())
	for i := 0; i < 6; i++ {
		assert.Equal(t, 4, o.GetPlayer(i).GetCardsSize())
	}
}

func TestOmaha_PlayerAction_Fold(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.lastBet = 50
	o.players[0].SetCurrentBet(0)

	err := o.PlayerAction(OmahaActionFold, 0, 0)
	assert.NoError(t, err)
}

func TestOmaha_PlayerAction_Check(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestOmaha_PlayerAction_Bet(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(OmahaActionBet, 20, 0)
	assert.NoError(t, err)
	var entry *ActionLogEntry
	for _, candidate := range o.GetActionLog() {
		if candidate.DetailCode == "omaha.log.bet" && candidate.DetailParams["amount"] == "20" {
			entry = candidate
		}
	}
	assert.NotNil(t, entry)
	assert.Equal(t, map[string]string{"amount": "20"}, entry.DetailParams)
}

func TestOmaha_PlayerAction_GameEnded(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.gameEndFlag = true
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_PlayerAction_WrongPhase(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseShowdown)
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_PlayerAction_NotHumanTurn(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.currentTurn = 1 // CPU
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_Getters(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()

	assert.NotNil(t, o.GetCommunityCards())
	assert.NotNil(t, o.GetSidePots())
	assert.Equal(t, 0, o.GetDealerIdx())
	// After reset, lastBet is BB (10) from blind posting
	assert.True(t, o.GetLastBet() >= 0)
	assert.NotNil(t, o.GetRoundResults())
	assert.NotNil(t, o.GetCpuActions())
	assert.Nil(t, o.GetLastCpuError())
	assert.Equal(t, 4, o.GetConfig().TableSize)
	assert.NotNil(t, o.GetActedFlags())
	assert.Equal(t, 1, o.GetHandCount())
	assert.NotNil(t, o.GetActionLog())
	assert.NotNil(t, o.GetRebuyCounts())
	assert.NotNil(t, o.GetAddonUsed())
	assert.Equal(t, OmahaRebuyPhaseNone, o.GetRebuyPhaseType())
}

func TestOmaha_GetPlayer_InvalidIndex(t *testing.T) {
	o := newTestOmaha()
	assert.Nil(t, o.GetPlayer(-1))
	assert.Nil(t, o.GetPlayer(100))
}

func TestOmaha_IsHumanTurn(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	o.currentTurn = 0
	assert.True(t, o.IsHumanTurn())
	o.currentTurn = 1
	assert.False(t, o.IsHumanTurn())
}

func TestOmaha_SetConfig(t *testing.T) {
	o := newTestOmaha()
	cfg := DefaultOmahaConfig()
	cfg.BigBlind = 20
	o.SetConfig(cfg)
	assert.Equal(t, 20, o.GetConfig().BigBlind)
}

func TestOmaha_Muck(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Muck())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		err := o.Muck()
		assert.NoError(t, err)
		assert.Equal(t, OmahaPhaseEnd, o.phase)
	})
}

func TestOmaha_ShowHand(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.ShowHand())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		err := o.ShowHand()
		assert.NoError(t, err)
		assert.Equal(t, OmahaPhaseEnd, o.phase)
	})
}

func TestOmaha_IsMuckAvailable(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.False(t, o.IsMuckAvailable())
	})
	t.Run("human lost", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		assert.True(t, o.IsMuckAvailable())
	})
	t.Run("human won", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 100},
		}
		assert.False(t, o.IsMuckAvailable())
	})
}

func TestOmaha_Rebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Rebuy())
	})
	t.Run("wrong rebuy type", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseRebuy
		o.rebuyPhaseType = OmahaRebuyPhaseAddon
		assert.Error(t, o.Rebuy())
	})
}

func TestOmaha_SkipRebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.SkipRebuy())
	})
}

func TestOmaha_Addon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Addon())
	})
	t.Run("wrong addon type", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseRebuy
		o.rebuyPhaseType = OmahaRebuyPhaseRebuy
		assert.Error(t, o.Addon())
	})
}

func TestOmaha_SkipAddon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.SkipAddon())
	})
}

func TestOmaha_IsRebuyAvailable(t *testing.T) {
	t.Run("rebuy disabled", func(t *testing.T) {
		o := newTestOmaha()
		assert.False(t, o.IsRebuyAvailable())
	})
}

func TestOmaha_IsAddonAvailable(t *testing.T) {
	t.Run("addon disabled", func(t *testing.T) {
		o := newTestOmaha()
		assert.False(t, o.IsAddonAvailable())
	})
}

func TestOmaha_PhaseConstants(t *testing.T) {
	assert.Equal(t, 0, OmahaPhaseInit)
	assert.Equal(t, 1, OmahaPhasePreFlop)
	assert.Equal(t, 2, OmahaPhaseFlop)
	assert.Equal(t, 3, OmahaPhaseTurn)
	assert.Equal(t, 4, OmahaPhaseRiver)
	assert.Equal(t, 5, OmahaPhaseShowdown)
	assert.Equal(t, 6, OmahaPhaseEnd)
	assert.Equal(t, 7, OmahaPhaseRebuy)
}

func TestOmaha_ActionConstants(t *testing.T) {
	assert.Equal(t, HoldemActionFold, OmahaActionFold)
	assert.Equal(t, HoldemActionCheck, OmahaActionCheck)
	assert.Equal(t, HoldemActionCall, OmahaActionCall)
	assert.Equal(t, HoldemActionBet, OmahaActionBet)
	assert.Equal(t, HoldemActionRaise, OmahaActionRaise)
	assert.Equal(t, HoldemActionAllIn, OmahaActionAllIn)
}

func TestOmaha_FullGame(t *testing.T) {
	// Play through multiple resets to verify game flow
	o := newTestOmaha()
	for i := 0; i < 3; i++ {
		err := o.Reset()
		assert.NoError(t, err)
		// Verify 4 cards per player
		for j := 0; j < o.GetPlayerCnt(); j++ {
			assert.Equal(t, 4, o.GetPlayer(j).GetCardsSize())
		}
	}
}

// ---------------------------------------------------------------------------
// Meta-AI integration tests
// ---------------------------------------------------------------------------

func TestOmaha_MetaAI_ProfileSurvivesReset(t *testing.T) {
	o := newTestOmaha()
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)

	_ = o.Reset()
	profile := o.GetHumanProfile()
	assert.NotNil(t, profile, "profile should be created on first Reset with CpuMetaAI=true")
	assert.Equal(t, 0, profile.GamesPlayed)

	_ = o.Reset()
	profile2 := o.GetHumanProfile()
	assert.NotNil(t, profile2)
	assert.Equal(t, 1, profile2.GamesPlayed)
}

func TestOmaha_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	assert.Nil(t, o.GetHumanProfile())
}

func TestOmaha_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	o := newTestOmaha()
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)
	_ = o.Reset()
	assert.NotNil(t, o.GetHumanProfile())

	o.ResetProfile()
	assert.Nil(t, o.GetHumanProfile())
}

func TestOmaha_MetaAI_LastHumanPlayMsResetOnReset(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)
	o.SetHumanProfile(&BettingHumanProfile{})
	o.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	})
	_ = o.PlayerAction(OmahaActionBet, 20, 800)
	assert.Equal(t, 800, o.GetLastHumanPlayMs(), "lastHumanPlayMs should be set after PlayerAction")

	_ = o.Reset()
	assert.Equal(t, 0, o.GetLastHumanPlayMs(), "lastHumanPlayMs should be reset to 0 on Reset")
}

func TestOmaha_MetaAI_PlayerActionRecordsAction(t *testing.T) {
	t.Run("aggressive action is recorded", func(t *testing.T) {
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
		cfg := o.GetConfig()
		cfg.CpuMetaAI = true
		o.SetConfig(cfg)
		o.SetHumanProfile(&BettingHumanProfile{})
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(OmahaActionBet, 20, 800)
		assert.NoError(t, err)

		profile := o.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 1, profile.HesitationCount)
		total := 0
		for i := 0; i < 3; i++ {
			total += profile.AggressiveByBracket[i].Total
		}
		assert.Equal(t, 1, total)
	})

	t.Run("fold on bet records fold-to-bet", func(t *testing.T) {
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
		cfg := o.GetConfig()
		cfg.CpuMetaAI = true
		o.SetConfig(cfg)
		o.SetHumanProfile(&BettingHumanProfile{})
		o.SetLastBet(40)
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(OmahaActionFold, 0, 0)
		assert.NoError(t, err)

		profile := o.GetHumanProfile()
		assert.Equal(t, 1, profile.FoldToBetCount)
		assert.Equal(t, 1, profile.FoldToBetTotal)
	})

	t.Run("no recording when CpuMetaAI is disabled", func(t *testing.T) {
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(OmahaActionBet, 20, 500)
		assert.NoError(t, err)
		assert.Nil(t, o.GetHumanProfile())
	})
}

func TestOmaha_GetPreflopCommunityCount(t *testing.T) {
	t.Run("Omaha", func(t *testing.T) {
		o := NewDefaultOmaha()
		assert.Equal(t, 0, o.GetPreflopCommunityCount())
	})
	t.Run("Big O", func(t *testing.T) {
		o := NewDefaultBigO()
		assert.Equal(t, 0, o.GetPreflopCommunityCount())
	})
	t.Run("Courchevel", func(t *testing.T) {
		o := NewDefaultCourchevel()
		assert.Equal(t, courchevelPreflopCommunity, o.GetPreflopCommunityCount())
	})
	t.Run("Courchevel Hi-Lo", func(t *testing.T) {
		o := NewDefaultCourchevelHiLo()
		assert.Equal(t, courchevelPreflopCommunity, o.GetPreflopCommunityCount())
	})
}
