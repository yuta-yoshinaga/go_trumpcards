//go:build test

// Tarneeb_coverage_test.go fills in branch coverage that the broader
// behavioural tests do not naturally exercise — single-line getters/setters,
// error-guard paths, JSON edge cases, and the less-trodden CPU AI branches.
// These tests run inside the domain package so they can poke at unexported
// hooks (filterByDesign, filterAbove, etc.).

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCov(diff TarneebCpuDifficulty) *Tarneeb {
	players := []*TarneebPlayer{
		NewTarneebPlayer(true, 0),
		NewTarneebPlayer(false, 1),
		NewTarneebPlayer(false, 0),
		NewTarneebPlayer(false, 1),
	}
	cfg := DefaultTarneebConfig()
	cfg.CpuDifficulty = diff
	return NewTarneeb(NewTrumpCards(0), players, cfg)
}

func TestTarneeb_Setters(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetRoundNumber(7)
	tn.SetTrickNumber(3)
	tn.SetCurrentPlayerIdx(2)
	tn.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 4, false)}})
	tn.SetTrumpSuit(CardDesignDiamond)
	tn.SetBidPlayerIdx(3)
	tn.SetBidWinnerIdx(2)
	tn.SetHighestBid(11)
	tn.SetDealerIdx(2)
	tn.SetLeadPlayerIdx(1)
	tn.SetTeamScore(0, 13)
	tn.SetTeamScore(1, 9)

	assert.Equal(t, TarneebPhasePlay, tn.GetPhase())
	assert.Equal(t, 7, tn.GetRoundNumber())
	assert.Equal(t, 3, tn.GetTrickNumber())
	assert.Equal(t, 2, tn.GetCurrentPlayerIdx())
	assert.Len(t, tn.GetCurrentTrick(), 1)
	assert.Equal(t, CardDesignDiamond, tn.GetTrumpSuit())
	assert.Equal(t, 3, tn.GetBidPlayerIdx())
	assert.Equal(t, 2, tn.GetBidWinnerIdx())
	assert.Equal(t, 11, tn.GetHighestBid())
	assert.Equal(t, 2, tn.GetDealerIdx())
	assert.Equal(t, 1, tn.GetLeadPlayerIdx())
	assert.Equal(t, 13, tn.GetTeamScore(0))
	assert.Equal(t, 9, tn.GetTeamScore(1))
	assert.Equal(t, 0, tn.GetTeamScore(99))
	tn.SetTeamScore(-1, 100) // out-of-range no-op
	assert.Equal(t, 13, tn.GetTeamScore(0))
}

func TestTarneeb_Getters_OutOfRange(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	assert.Nil(t, tn.GetPlayer(-1))
	assert.Nil(t, tn.GetPlayer(99))
	assert.NotNil(t, tn.GetPlayer(0))
	assert.Equal(t, TarneebPlayerCnt, tn.GetPlayerCnt())
	assert.NotNil(t, tn.GetConfig())

	// IsHumanXTurn with out-of-range indices.
	tn.SetCurrentPlayerIdx(-1)
	assert.False(t, tn.IsHumanTurn())
	tn.SetBidPlayerIdx(-1)
	assert.False(t, tn.IsHumanBidTurn())
	tn.SetBidWinnerIdx(-1)
	assert.False(t, tn.IsHumanTrumpTurn())
}

func TestTarneeb_PlayerActions_GameEnded(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.gameEndFlag = true

	assert.ErrorIs(t, tn.PlayerBid(7), ErrGameEnded)
	assert.ErrorIs(t, tn.PlayerDeclareTrump(CardDesignSpade), ErrGameEnded)
	assert.ErrorIs(t, tn.PlayerPlay(0), ErrGameEnded)
	tn.CpuBid()
	tn.CpuDeclareTrump()
	tn.CpuPlay()
	// Should remain in whatever phase Reset set; gameEndFlag should still be true.
	assert.True(t, tn.GetGameEndFlag())
}

func TestTarneeb_PlayerBid_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	assert.ErrorIs(t, tn.PlayerBid(7), ErrWrongPhase)
}

func TestTarneeb_PlayerPlay_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	assert.ErrorIs(t, tn.PlayerPlay(0), ErrWrongPhase)
}

func TestTarneeb_PlayerBid_NotHumanIdx(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetBidPlayerIdx(1) // CPU's turn
	assert.ErrorIs(t, tn.PlayerBid(7), ErrNotHumanTurn)
}

func TestTarneeb_PlayerDeclareTrump_NotHuman(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(1)
	assert.ErrorIs(t, tn.PlayerDeclareTrump(CardDesignSpade), ErrNotHumanTurn)
}

func TestTarneeb_ResolveTrick_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.ResolveTrick() // no-op
	assert.Equal(t, TarneebPhaseBid, tn.GetPhase())
}

func TestTarneeb_NextTrick_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.NextTrick()
	assert.Equal(t, TarneebPhaseBid, tn.GetPhase())
}

func TestTarneeb_NextRound_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.NextRound() // no-op
	assert.Equal(t, 1, tn.GetRoundNumber())
}

func TestTarneeb_ScoreRound_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.ScoreRound() // no-op
	assert.Equal(t, 0, tn.GetTeamScore(0))
}

func TestTarneeb_CpuBid_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.CpuBid() // no-op
	assert.Equal(t, TarneebPhasePlay, tn.GetPhase())
}

func TestTarneeb_CpuDeclareTrump_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.CpuDeclareTrump()
	assert.Equal(t, TarneebPhaseBid, tn.GetPhase())
}

func TestTarneeb_CpuPlay_WrongPhase(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseBid)
	tn.CpuPlay() // no-op
}

func TestTarneeb_CpuBid_HumanTurn(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	tn.CpuBid() // human's turn → no-op
	assert.Equal(t, -1, tn.GetPlayer(0).GetBid())
}

func TestTarneeb_CpuPlay_HumanTurn(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(0)
	before := tn.GetPlayer(0).GetCardsSize()
	tn.CpuPlay() // human's turn → no-op
	assert.Equal(t, before, tn.GetPlayer(0).GetCardsSize())
}

func TestTarneeb_PlayerPlay_NotHumanTurn(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, tn.PlayerPlay(0), ErrNotHumanTurn)
}

func TestTarneeb_PlayerPlay_BadValidate(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	leadCard := NewCard(CardDesignHeart, 6, false)
	tn.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: leadCard}})
	fillHand(tn.GetPlayer(0), []*Card{
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignClover, 2, false),
	})
	// Trying to play clover while holding hearts → invalid play.
	err := tn.PlayerPlay(1)
	require.Error(t, err)
}

func TestTarneeb_JSON_RoundTrip_WithTrick(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetTrickNumber(5)
	tn.SetBidWinnerIdx(2)
	tn.SetHighestBid(9)
	tn.SetTeamScore(0, 4)
	tn.SetTeamScore(1, -1)
	tn.SetLeadPlayerIdx(2)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 1, false)},
	})

	data, err := json.Marshal(tn)
	require.NoError(t, err)
	got := NewTarneeb(NewTrumpCards(0), nil, DefaultTarneebConfig())
	require.NoError(t, json.Unmarshal(data, got))
	assert.Equal(t, tn.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, tn.GetBidWinnerIdx(), got.GetBidWinnerIdx())
	assert.Equal(t, tn.GetHighestBid(), got.GetHighestBid())
	assert.Equal(t, tn.GetTeamScore(0), got.GetTeamScore(0))
	assert.Equal(t, tn.GetTeamScore(1), got.GetTeamScore(1))
	assert.Len(t, got.GetCurrentTrick(), 2)
	assert.Equal(t, tn.GetTrickNumber(), got.GetTrickNumber())
}

func TestTarneeb_UnmarshalJSON_OversizedArray(t *testing.T) {
	tn := NewTarneeb(NewTrumpCards(0), nil, DefaultTarneebConfig())
	// Fabricate an oversized players array via raw JSON.
	huge := `{"ps":[`
	for i := 0; i < tarneebMaxSliceLen+1; i++ {
		if i > 0 {
			huge += ","
		}
		huge += `{"gp":{"ih":true}}`
	}
	huge += `]}`
	assert.Error(t, tn.UnmarshalJSON([]byte(huge)))
}

func TestTarneeb_PlayerName_OutOfRange(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	assert.Contains(t, tn.playerName(-1), "Player")
	assert.Equal(t, "You", tn.playerName(0))
	assert.Contains(t, tn.playerName(1), "CPU")
}

func TestTarneeb_TrickWinner_AllTrumps(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 11, false)},
	})
	assert.Equal(t, 1, tn.trickWinner())
}

func TestTarneeb_TrickWinner_NoCards(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetCurrentTrick(nil)
	assert.Equal(t, 0, tn.trickWinner())
}

func TestTarneeb_SummariseTrick_NoTrump(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 9, false)},
	})
	maxLead, hasLead, maxTrump, hasTrump := tn.summariseTrick(CardDesignHeart)
	assert.Equal(t, 9, maxLead)
	assert.True(t, hasLead)
	assert.Equal(t, 0, maxTrump)
	assert.False(t, hasTrump)
}

func TestTarneeb_IsPartnerCurrentlyWinning(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetTrumpSuit(CardDesignSpade)

	// Empty trick → false
	tn.SetCurrentTrick(nil)
	assert.False(t, tn.isPartnerCurrentlyWinning(0))

	// Idx 2 is partner of 0; idx 2 leads the trick → from idx 0's POV partner is winning.
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 12, false)},
	})
	assert.True(t, tn.isPartnerCurrentlyWinning(0))

	// Idx 1 winning is not partner of idx 0.
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 5, false)},
	})
	assert.False(t, tn.isPartnerCurrentlyWinning(0))

	// Self-winning is not "partner winning".
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, false)},
	})
	assert.False(t, tn.isPartnerCurrentlyWinning(0))
}

func TestTarneeb_FilterByDesign_And_FilterAbove(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	p := tn.GetPlayer(0)
	fillHand(p, []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 9, false),
	})
	valid := []int{0, 1, 2, 3}
	spades := tn.filterByDesign(p, valid, CardDesignSpade)
	assert.Equal(t, []int{0, 1}, spades)
	hearts := tn.filterByDesign(p, valid, CardDesignHeart)
	assert.Equal(t, []int{2, 3}, hearts)

	above := tn.filterAbove(p, valid, 5)
	assert.Equal(t, []int{1, 3}, above)
}

func TestTarneeb_PickHighestLowest(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	p := tn.GetPlayer(0)
	fillHand(p, []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 3, false),
	})
	assert.Equal(t, 1, tn.pickHighest(p, []int{0, 1, 2}))
	assert.Equal(t, 2, tn.pickLowest(p, []int{0, 1, 2}))
}

func TestTarneeb_CpuPlayNormal_Lead_Overcards(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	// Existing trick led with H-9. Player 1 has H-10 and H-K — should play the lowest over (10).
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 9, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 13, false),
	})
	idx := tn.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 0, idx)
}

func TestTarneeb_CpuPlayNormal_Void_TrumpAvailable(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 9, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignSpade, 4, false),
	})
	idx := tn.cpuPlayNormal(1, []int{0, 1})
	// Void of hearts but holds trump → trump cut (idx 1).
	assert.Equal(t, 1, idx)
}

func TestTarneeb_CpuPlayNormal_Void_TrumpInTrick(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 9, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 5, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 2, false),
	})
	idx := tn.cpuPlayNormal(1, []int{0, 1})
	// Over-cut with S-8.
	assert.Equal(t, 0, idx)
}

func TestTarneeb_CpuPlayHard_Lead(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick(nil)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 13, false),  // trump K
		NewCard(CardDesignHeart, 1, false),   // ace
		NewCard(CardDesignDiamond, 4, false), // low
	})
	idx := tn.cpuPlayHard(1, []int{0, 1, 2})
	// Hard CPU prefers the off-suit ace (boosted score, no trump penalty).
	assert.Equal(t, 1, idx)
}

func TestTarneeb_CpuPlayHard_Void_NoTrump(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignClover, 11, false),
	})
	idx := tn.cpuPlayHard(1, []int{0, 1})
	// No trump available, partner isn't winning, discard highest non-trump (clover J).
	assert.Equal(t, 1, idx)
}

func TestTarneeb_CpuPlayHard_PartnerWinning_Void(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	// Idx 1's partner is idx 3; idx 3 leads with H-A → partner winning. Idx 1 is void.
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 1, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 13, false),
	})
	idx := tn.cpuPlayHard(1, []int{0, 1})
	// Partner winning → low card discard (diamond 4).
	assert.Equal(t, 0, idx)
}

func TestTarneeb_CpuPlayHard_LeadSuit_PartnerWinning(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	// Partner (idx 3) of idx 1 winning with H-A. Idx 1 has H-5 and H-K → must play low.
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 1, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 13, false),
	})
	idx := tn.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx) // H-5
}

func TestTarneeb_CpuSelectPlayCard_Single(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
	})
	fillHand(tn.GetPlayer(1), []*Card{NewCard(CardDesignHeart, 3, false)})
	idx := tn.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestTarneeb_CpuSelectPlayCard_Empty(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyHard)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
	})
	// Empty hand → fallback 0.
	fillHand(tn.GetPlayer(1), []*Card{})
	idx := tn.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestTarneeb_AdjustToValidBid_ClampMin(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.SetHighestBid(0)
	// candidate < MinBid → clamps to MinBid.
	assert.Equal(t, tn.config.MinBid, tn.adjustToValidBid(2))
}

func TestTarneeb_FinishBidPhase_Redeal(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetDealerIdx(3)
	tn.SetBidPlayerIdx(0)
	// All 4 players pass directly via the internal path so we don't depend on CPU RNG.
	for i := 0; i < TarneebPlayerCnt; i++ {
		tn.applyBid(i, TarneebPassBid)
	}
	assert.Equal(t, 1, tn.GetRedealCount())
	assert.Equal(t, TarneebPhaseBid, tn.GetPhase())
}

func TestTarneeb_CpuBid_FullDifficultyMatrix(t *testing.T) {
	for _, diff := range []TarneebCpuDifficulty{
		TarneebCpuDifficultyEasy,
		TarneebCpuDifficultyNormal,
		TarneebCpuDifficultyHard,
	} {
		tn := newCov(diff)
		tn.Reset()
		tn.SetBidPlayerIdx(1)
		tn.CpuBid()
		bid := tn.GetPlayer(1).GetBid()
		assert.True(t, bid == TarneebPassBid || (bid >= tn.config.MinBid && bid <= TarneebMaxBid),
			"unexpected bid for difficulty %d: %d", diff, bid)
	}
}

func TestTarneeb_CpuDeclareTrump_FullDifficultyMatrix(t *testing.T) {
	for _, diff := range []TarneebCpuDifficulty{
		TarneebCpuDifficultyEasy,
		TarneebCpuDifficultyNormal,
		TarneebCpuDifficultyHard,
	} {
		tn := newCov(diff)
		tn.Reset()
		tn.SetPhase(TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(1)
		tn.CpuDeclareTrump()
		assert.True(t, isValidSuit(tn.GetTrumpSuit()), "invalid trump for difficulty %d", diff)
	}
}

func TestTarneeb_GetHint_NoHuman(t *testing.T) {
	// Build a no-human game; GetHint should return nil immediately.
	players := []*TarneebPlayer{
		NewTarneebPlayer(false, 0),
		NewTarneebPlayer(false, 1),
		NewTarneebPlayer(false, 0),
		NewTarneebPlayer(false, 1),
	}
	tn := NewTarneeb(NewTrumpCards(0), players, DefaultTarneebConfig())
	tn.Reset()
	assert.Nil(t, tn.GetHint())
}

func TestTarneeb_GetHint_NotMyTurn(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetBidPlayerIdx(1)
	assert.Nil(t, tn.GetHint())

	tn.SetPhase(TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(1)
	assert.Nil(t, tn.GetHint())

	tn.SetPhase(TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(1)
	assert.Nil(t, tn.GetHint())
}

func TestTarneeb_GetValidPlayIndices_Public(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(0)
	leadCard := NewCard(CardDesignHeart, 5, false)
	tn.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: leadCard}})
	fillHand(tn.GetPlayer(0), []*Card{
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignSpade, 4, false),
	})
	valid := tn.GetValidPlayIndices(0)
	assert.Equal(t, []int{0}, valid)
}

func TestTarneeb_NextRound_Resets(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseRoundEnd)
	tn.SetTeamScore(0, 5)
	tn.SetTeamScore(1, 3)
	dealerBefore := tn.GetDealerIdx()
	tn.NextRound()
	assert.Equal(t, (dealerBefore+1)%TarneebPlayerCnt, tn.GetDealerIdx())
	assert.Equal(t, 2, tn.GetRoundNumber())
	assert.Equal(t, TarneebPhaseBid, tn.GetPhase())
	// Team scores persist into the next round.
	assert.Equal(t, 5, tn.GetTeamScore(0))
	assert.Equal(t, 3, tn.GetTeamScore(1))
}

func TestTarneeb_PlayerBid_BadBidPlayerIdx(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetBidPlayerIdx(-1)
	assert.ErrorIs(t, tn.PlayerBid(7), ErrWrongPhase)

	tn.SetBidPlayerIdx(99)
	assert.ErrorIs(t, tn.PlayerBid(7), ErrWrongPhase)
}

func TestTarneeb_SetConfig_GetActionLog(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	cfg := tn.GetConfig()
	cfg.PointLimit = 41
	tn.SetConfig(cfg)
	assert.Equal(t, 41, tn.GetConfig().PointLimit)

	tn.SetPhase(TarneebPhaseRoundEnd)
	tn.SetBidWinnerIdx(0)
	tn.SetHighestBid(7)
	tn.ScoreRound()
	logEntries := tn.GetActionLog()
	assert.NotEmpty(t, logEntries)
}

func TestTarneeb_PlayerDeclareTrump_BadBidWinnerIdx(t *testing.T) {
	tn := newCov(TarneebCpuDifficultyNormal)
	tn.Reset()
	tn.SetPhase(TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(-1)
	assert.ErrorIs(t, tn.PlayerDeclareTrump(CardDesignSpade), ErrNotHumanTurn)
}
