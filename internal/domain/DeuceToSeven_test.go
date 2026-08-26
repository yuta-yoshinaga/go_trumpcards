package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestDeuceToSeven() (*DeuceToSeven, []*DeuceToSevenPlayer) {
	tc := NewTrumpCards(0)
	p0 := NewDeuceToSevenPlayer(true, DeuceToSevenStyleBalanced)
	p1 := NewDeuceToSevenPlayer(false, DeuceToSevenStyleConservative)
	p2 := NewDeuceToSevenPlayer(false, DeuceToSevenStyleAggressive)
	p3 := NewDeuceToSevenPlayer(false, DeuceToSevenStyleBluffer)
	players := []*DeuceToSevenPlayer{p0, p1, p2, p3}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	return NewDeuceToSeven(tc, players, DefaultDeuceToSevenConfig()), players
}

func setupDeuceToSevenForHumanBet(drawIdx int) (*DeuceToSeven, []*DeuceToSevenPlayer) {
	dt, players := newTestDeuceToSeven()
	if drawIdx == 0 {
		dt.SetPhase(DeuceToSevenPhaseDeal)
	} else {
		dt.SetPhase(DeuceToSevenPhaseBet)
	}
	dt.SetDrawIndex(drawIdx)
	dt.SetCurrentTurn(0)
	dt.setActedFlags([]bool{false, true, true, true})
	dt.SetLastBet(0)
	dt.SetMinRaise(10)
	dt.SetPot(40)
	dt.setStartingChips([]int{1000, 1000, 1000, 1000})
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	hand := []struct{ d, v int }{{S, 2}, {H, 3}, {D, 4}, {C, 5}, {S, 7}}
	for _, pl := range players {
		pl.Reset()
		pl.SetChips(990)
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		for _, hc := range hand {
			pl.AddCard(NewCard(hc.d, hc.v, false))
		}
	}
	return dt, players
}

func giveDeuceToSevenHand(pl *DeuceToSevenPlayer, cards []*Card) {
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// ---------------------------------------------------------------------------
// construction
// ---------------------------------------------------------------------------

func TestNewDeuceToSeven(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	assert.Equal(t, DeuceToSevenPhaseInit, dt.GetPhase())
	assert.Equal(t, 4, len(dt.GetPlayers()))
	assert.Equal(t, 0, dt.GetPot())
	assert.False(t, dt.GetGameEndFlag())
	assert.Equal(t, players, dt.GetPlayers())
	assert.Equal(t, 0, dt.GetDrawIndex())
}

func TestNewDefaultDeuceToSeven(t *testing.T) {
	dt := NewDefaultDeuceToSeven()
	assert.Equal(t, 4, len(dt.GetPlayers()))
	assert.True(t, dt.GetPlayers()[0].GetIsHuman())
	for _, pl := range dt.GetPlayers()[1:] {
		assert.False(t, pl.GetIsHuman())
	}
}

// ---------------------------------------------------------------------------
// Reset behaviour
// ---------------------------------------------------------------------------

func TestDeuceToSeven_Reset_DealsFiveCardsAndCollectsAntes(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	assert.GreaterOrEqual(t, dt.GetPot(), 40)
	for _, pl := range players {
		assert.Equal(t, 5, pl.GetCardsSize(), "every active seat gets exactly 5 cards")
	}
	assert.False(t, players[0].GetFolded())
}

func TestDeuceToSeven_Reset_RejectsInvalidConfig(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetConfig(DeuceToSevenConfig{
		BettingLimit: BettingLimitType(-1),
		CpuCount:     1,
	})
	assert.Error(t, dt.Reset())
}

func TestDeuceToSeven_Reset_FoldsBeyondSeatCap(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	cfg := dt.GetConfig()
	cfg.CpuCount = 1 // only 2 active seats
	dt.SetConfig(cfg)
	require.NoError(t, dt.Reset())
	assert.False(t, players[0].GetFolded())
	assert.False(t, players[1].GetFolded())
	assert.True(t, players[2].GetFolded(), "seat index 2 exceeds cap")
	assert.True(t, players[3].GetFolded())
}

// ---------------------------------------------------------------------------
// PlayerAction — wrong state / turn
// ---------------------------------------------------------------------------

func TestDeuceToSeven_PlayerAction_RejectsWhenGameEnded(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetGameEndFlag(true)
	assert.Error(t, dt.PlayerAction(DeuceToSevenActionCheck, 0, 0))
}

func TestDeuceToSeven_PlayerAction_RejectsWrongPhase(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetPhase(DeuceToSevenPhaseDraw)
	assert.Error(t, dt.PlayerAction(DeuceToSevenActionCheck, 0, 0))
}

func TestDeuceToSeven_PlayerAction_RejectsNotHumanTurn(t *testing.T) {
	dt, _ := setupDeuceToSevenForHumanBet(0)
	dt.SetCurrentTurn(1) // CPU seat
	assert.Error(t, dt.PlayerAction(DeuceToSevenActionCheck, 0, 0))
}

func TestDeuceToSeven_PlayerAction_CheckAdvancesToDraw(t *testing.T) {
	dt, _ := setupDeuceToSevenForHumanBet(0)
	require.NoError(t, dt.PlayerAction(DeuceToSevenActionCheck, 0, 0))
	phase := dt.GetPhase()
	assert.True(t, phase == DeuceToSevenPhaseDraw || phase == DeuceToSevenPhaseBet ||
		phase == DeuceToSevenPhaseEnd, "unexpected phase %d", phase)
}

// ---------------------------------------------------------------------------
// PlayerExchange
// ---------------------------------------------------------------------------

func TestDeuceToSeven_PlayerExchange_RejectsOutsideDraw(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetPhase(DeuceToSevenPhaseBet)
	assert.Error(t, dt.PlayerExchange([]int{0}))
}

func TestDeuceToSeven_PlayerExchange_ReplacesSelectedIndicesOnly(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	dt.SetPhase(DeuceToSevenPhaseDraw)
	dt.SetDrawIndex(1)
	dt.SetCurrentTurn(0)
	dt.setActedFlags([]bool{false, true, true, true})
	dt.ResetPlayerHand(0)
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	hand := []*Card{
		NewCard(S, 5, false), NewCard(H, 9, false),
		NewCard(D, 11, false), NewCard(C, 13, false), NewCard(S, 2, false),
	}
	giveDeuceToSevenHand(players[0], hand)
	before := []int{
		players[0].GetCard(0).GetValue(),
		players[0].GetCard(2).GetValue(),
		players[0].GetCard(4).GetValue(),
	}
	require.NoError(t, dt.PlayerExchange([]int{1, 3}))
	assert.Equal(t, before[0], players[0].GetCard(0).GetValue())
	assert.Equal(t, before[1], players[0].GetCard(2).GetValue())
	assert.Equal(t, before[2], players[0].GetCard(4).GetValue())
	assert.Equal(t, 2, players[0].GetDrawCount())
}

func TestDeuceToSeven_PlayerStand_ExchangesZeroCards(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	dt.SetPhase(DeuceToSevenPhaseDraw)
	dt.SetDrawIndex(1)
	dt.SetCurrentTurn(0)
	dt.setActedFlags([]bool{false, true, true, true})
	require.NoError(t, dt.PlayerStand())
	assert.Equal(t, 0, players[0].GetDrawCount())
}

// ---------------------------------------------------------------------------
// Betting resolution — everyone folds to one player
// ---------------------------------------------------------------------------

func TestDeuceToSeven_LastPlayerStandingWinsPot(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	dt.SetCurrentTurn(0)
	dt.SetPhase(DeuceToSevenPhaseDeal)
	dt.SetLastBet(0)
	players[0].SetCurrentBet(0)
	dt.setActedFlags([]bool{false, true, true, true})
	prevChips := players[0].GetChips()
	pot := dt.GetPot()
	require.NoError(t, dt.PlayerAction(DeuceToSevenActionCheck, 0, 0))
	assert.True(t, dt.GetGameEndFlag())
	assert.Equal(t, prevChips+pot, players[0].GetChips())
	assert.Equal(t, 0, dt.GetPot())
}

// ---------------------------------------------------------------------------
// Showdown
// ---------------------------------------------------------------------------

func TestDeuceToSeven_Showdown_SplitsEqualHands(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	nut := []*Card{
		NewCard(S, 7, false), NewCard(H, 5, false), NewCard(D, 4, false),
		NewCard(C, 3, false), NewCard(S, 2, false),
	}
	giveDeuceToSevenHand(players[0], nut)
	giveDeuceToSevenHand(players[1], nut)

	dt.SetPot(200)
	dt.setStartingChips([]int{1000, 1000, 1000, 1000})
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetDrawIndex(3)
	dt.resolveShowdown()

	assert.True(t, dt.GetGameEndFlag())
	results := dt.GetRoundResults()
	assert.Len(t, results, 2)
	total := 0
	for _, r := range results {
		total += r.WonAmount
	}
	assert.Equal(t, 200, total)
}

func TestDeuceToSeven_Showdown_BestLowTakesPot(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// Player 0: nut low 7-5-4-3-2.
	giveDeuceToSevenHand(players[0], []*Card{
		NewCard(S, 7, false), NewCard(H, 5, false), NewCard(D, 4, false),
		NewCard(C, 3, false), NewCard(S, 2, false),
	})
	// Player 1: a pair of kings — far worse for a low.
	giveDeuceToSevenHand(players[1], []*Card{
		NewCard(S, 13, false), NewCard(H, 13, false), NewCard(D, 4, false),
		NewCard(C, 3, false), NewCard(S, 2, false),
	})

	dt.SetPot(200)
	dt.setStartingChips([]int{1000, 1000, 1000, 1000})
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetDrawIndex(3)
	dt.resolveShowdown()

	results := dt.GetRoundResults()
	var wonByP0, wonByP1 int
	for _, r := range results {
		switch r.PlayerIdx {
		case 0:
			wonByP0 = r.WonAmount
		case 1:
			wonByP1 = r.WonAmount
		}
	}
	assert.Equal(t, 200, wonByP0)
	assert.Equal(t, 0, wonByP1)
}

// ---------------------------------------------------------------------------
// FixedLimit late-round bet sizing
// ---------------------------------------------------------------------------

func TestDeuceToSeven_FixedLimit_LateRoundUsesBigBet(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	cfg := dt.GetConfig()
	cfg.BettingLimit = BettingLimitFixed
	cfg.MinBet = 10
	dt.SetConfig(cfg)
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetDrawIndex(2)
	assert.Equal(t, 20, dt.currentMinBet())

	dt.SetDrawIndex(1)
	assert.Equal(t, 10, dt.currentMinBet())
}

func TestDeuceToSeven_PotLimit_UsesMinBetConstant(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	cfg := dt.GetConfig()
	cfg.BettingLimit = BettingLimitPotLimit
	cfg.MinBet = 10
	dt.SetConfig(cfg)
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetDrawIndex(3)
	assert.Equal(t, 10, dt.currentMinBet())
}

func TestDeuceToSeven_NoLimit_UsesMinBetConstant(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	cfg := dt.GetConfig()
	cfg.BettingLimit = BettingLimitNoLimit
	cfg.MinBet = 10
	dt.SetConfig(cfg)
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetDrawIndex(3)
	assert.Equal(t, 10, dt.currentMinBet())
}

// ---------------------------------------------------------------------------
// CPU exchange logic
// ---------------------------------------------------------------------------

func TestDeuceToSeven_CpuDecideExchange_StandsPatOnMadeLow(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// Conservative (drawStandPatSize 4) with a pat 7-5-4-3-2 stands pat.
	giveDeuceToSevenHand(players[1], []*Card{
		NewCard(S, 7, false), NewCard(H, 5, false), NewCard(D, 4, false),
		NewCard(C, 3, false), NewCard(S, 2, false),
	})
	assert.Empty(t, dt.cpuDecideExchange(1), "made 7-low should stand pat")
}

func TestDeuceToSeven_CpuDecideExchange_DropsPairAndHighCards(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// Pair of aces + a king alongside two low cards; Conservative bluffs 0%.
	giveDeuceToSevenHand(players[1], []*Card{
		NewCard(S, 1, false), NewCard(C, 1, false), NewCard(H, 13, false),
		NewCard(D, 3, false), NewCard(S, 4, false),
	})
	for range 10 {
		indices := dt.cpuDecideExchange(1)
		assert.NotEmpty(t, indices)
		// The two low cards (3 and 4 at indices 3,4) must be kept.
		for _, idx := range indices {
			assert.NotContains(t, []int{3, 4}, idx)
		}
	}
}

func TestDeuceToSeven_CpuDecideExchange_BreaksMadeStraight(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// A made straight 2-3-4-5-6 (strength 1) must be broken by drawing.
	giveDeuceToSevenHand(players[1], []*Card{
		NewCard(S, 2, false), NewCard(H, 3, false), NewCard(D, 4, false),
		NewCard(C, 5, false), NewCard(S, 6, false),
	})
	indices := dt.cpuDecideExchange(1)
	assert.Len(t, indices, 1, "breaking a straight discards exactly the highest card")
	assert.Equal(t, 4, indices[0], "the 6 (index 4) is the highest and is dropped")
}

// ---------------------------------------------------------------------------
// JSON round-trip
// ---------------------------------------------------------------------------

func TestDeuceToSeven_JSON_RoundTrip(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())
	data, err := json.Marshal(dt)
	require.NoError(t, err)

	round := &DeuceToSeven{}
	require.NoError(t, json.Unmarshal(data, round))
	assert.Equal(t, dt.GetPhase(), round.GetPhase())
	assert.Equal(t, dt.GetPot(), round.GetPot())
	assert.Equal(t, dt.GetDrawIndex(), round.GetDrawIndex())
	assert.Equal(t, dt.GetDealerIdx(), round.GetDealerIdx())
	assert.Equal(t, dt.GetConfig(), round.GetConfig())
	assert.Equal(t, len(dt.GetPlayers()), len(round.GetPlayers()))
}

func TestDeuceToSeven_JSON_TooManyPlayersRejected(t *testing.T) {
	payload := []byte(`{"pl":[` + replicatePlayerJSON(1001) + `]}`)
	err := json.Unmarshal(payload, &DeuceToSeven{})
	assert.Error(t, err)
}

func TestDeuceToSeven_JSON_ActedFlagsMismatchRejected(t *testing.T) {
	// 1 player but 0 acted flags → should be rejected to prevent index-out-of-bounds.
	payload := []byte(`{"pl":[{}],"rd":{"af":[],"sc":[1000]}}`)
	err := json.Unmarshal(payload, &DeuceToSeven{})
	assert.Error(t, err, "mismatched ActedFlags length should be rejected")
}

func TestDeuceToSeven_JSON_StartingChipsMismatchRejected(t *testing.T) {
	// 1 player but 0 starting chips → should be rejected.
	payload := []byte(`{"pl":[{}],"rd":{"af":[false],"sc":[]}}`)
	err := json.Unmarshal(payload, &DeuceToSeven{})
	assert.Error(t, err, "mismatched StartingChips length should be rejected")
}

// ---------------------------------------------------------------------------
// Helpers / getters
// ---------------------------------------------------------------------------

func TestDeuceToSeven_GettersRoundTrip(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetPhase(DeuceToSevenPhaseBet)
	dt.SetPot(42)
	dt.SetDealerIdx(2)
	dt.SetCurrentTurn(1)
	dt.SetLastBet(20)
	dt.SetMinRaise(40)
	dt.setRaiseCount(2)

	assert.Equal(t, DeuceToSevenPhaseBet, dt.GetPhase())
	assert.Equal(t, 42, dt.GetPot())
	assert.Equal(t, 2, dt.GetDealerIdx())
	assert.Equal(t, 1, dt.GetCurrentTurn())
	assert.Equal(t, 20, dt.GetLastBet())
	assert.Equal(t, 40, dt.GetMinRaise())
	assert.Equal(t, 2, dt.GetRaiseCount())
	assert.Equal(t, DefaultDeuceToSevenConfig().Ante, dt.GetAnte())
}

func TestDeuceToSeven_GetActionLog(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	assert.Empty(t, dt.GetActionLog())
	require.NoError(t, dt.Reset())
	assert.NotEmpty(t, dt.GetActionLog(), "ante entries should be logged")
}

// ---------------------------------------------------------------------------
// Full hand smoke test — three complete draws then showdown
// ---------------------------------------------------------------------------

func TestDeuceToSeven_FullHandProgressesToShowdownOrFold(t *testing.T) {
	dt, players := newTestDeuceToSeven()
	require.NoError(t, dt.Reset())

	const maxSteps = 16
	for step := 0; step < maxSteps && !dt.GetGameEndFlag(); step++ {
		phase := dt.GetPhase()
		if !players[dt.GetCurrentTurn()].GetIsHuman() {
			break
		}
		switch phase {
		case DeuceToSevenPhaseDeal, DeuceToSevenPhaseBet:
			callAmt := dt.GetLastBet() - players[0].GetCurrentBet()
			if callAmt == 0 {
				if err := dt.PlayerAction(DeuceToSevenActionCheck, 0, 0); err != nil {
					require.NoError(t, dt.PlayerAction(DeuceToSevenActionCall, 0, 0))
				}
			} else {
				require.NoError(t, dt.PlayerAction(DeuceToSevenActionCall, 0, 0))
			}
		case DeuceToSevenPhaseDraw:
			require.NoError(t, dt.PlayerStand())
		default:
			dt.SetGameEndFlag(true)
		}
	}
	assert.True(t, dt.GetGameEndFlag(), "hand should resolve within %d steps", maxSteps)
}

// ---------------------------------------------------------------------------
// Profile import/export
// ---------------------------------------------------------------------------

func TestDeuceToSeven_ProfileExportNilWhenMissing(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	assert.Nil(t, dt.ExportProfile())
}

func TestDeuceToSeven_ProfileImportExportRoundTrip(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	cfg := dt.GetConfig()
	cfg.CpuMetaAI = true
	dt.SetConfig(cfg)
	require.NoError(t, dt.Reset())
	exported := dt.ExportProfile()
	require.NotNil(t, exported)
	data, err := json.Marshal(exported)
	require.NoError(t, err)

	other, _ := newTestDeuceToSeven()
	require.NoError(t, other.ImportProfile(data))
	assert.NotNil(t, other.GetHumanProfile())

	require.NoError(t, other.ImportProfile(nil))
}

func TestDeuceToSeven_ResetProfileClears(t *testing.T) {
	dt, _ := newTestDeuceToSeven()
	dt.SetHumanProfile(&BettingHumanProfile{})
	dt.ResetProfile()
	assert.Nil(t, dt.GetHumanProfile())
}

// ---------------------------------------------------------------------------
// Player-level helpers
// ---------------------------------------------------------------------------

func TestDeuceToSevenPlayer_EvalHandAndNames(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	pl := NewDeuceToSevenPlayer(false, DeuceToSevenStyleBalanced)
	giveDeuceToSevenHand(pl, []*Card{
		NewCard(S, 2, false), NewCard(H, 2, false), NewCard(D, 4, false),
		NewCard(C, 6, false), NewCard(S, 8, false),
	})
	assert.Equal(t, PokerHandOnePair, pl.EvalHand())
	assert.Equal(t, "One Pair", pl.GetHandName())
	assert.Equal(t, 1, pl.GetStrength())
	assert.Equal(t, "Balanced", pl.GetPlayStyleName())

	// JSON round-trip.
	data, err := json.Marshal(pl)
	require.NoError(t, err)
	other := &DeuceToSevenPlayer{}
	require.NoError(t, json.Unmarshal(data, other))
	assert.Equal(t, pl.GetPlayStyle(), other.GetPlayStyle())
}

func TestDeuceToSevenPlayer_GetHandNameUnknown(t *testing.T) {
	pl := NewDeuceToSevenPlayer(false, DeuceToSevenStyleBalanced)
	pl.SetHandRank(-1)
	assert.Equal(t, "Unknown", pl.GetHandName())
}

func TestDeuceToSevenConfig_Validate(t *testing.T) {
	cfg := DefaultDeuceToSevenConfig()
	assert.NoError(t, cfg.Validate())

	cfg.CpuCount = 0
	assert.Error(t, cfg.Validate())

	cfg = DefaultDeuceToSevenConfig()
	cfg.CpuCount = DeuceToSevenCpuCountMax + 1
	assert.Error(t, cfg.Validate())
}

// **山が尽きても交換は成立する。** 52 枚・4 席だと配った時点の山は 32 枚しか
// なく、3 回のドローで全員が 5 枚引くと最大 60 枚要る。足りなくなった時点で
// 黙って break していたので、**捨てたはずの札がそのまま手元に残り**、画面には
// 何も出なかった。カジノの規則どおり、捨て札 (マック) を切り直して引く (#6233)。
func TestDeuceToSeven_ExchangeRecyclesTheMuckWhenTheStockRunsOut(t *testing.T) {
	dt, players := setupDeuceToSevenForHumanBet(1)
	dt.SetPhase(DeuceToSevenPhaseDraw)
	dt.SetCurrentTurn(0)

	// 先に何度か交換して、捨て札を積んでおく (マックの元)。
	dt.applyExchange(1, []int{0, 1, 2, 3, 4})
	dt.applyExchange(2, []int{0, 1, 2, 3, 4})

	// 山を空にする。
	for dt.trumpCards.DrawCard() != nil {
	}
	require.Equal(t, 0, dt.trumpCards.GetRemainingCount(), "山が空になっていない")

	before := make([]*Card, DeuceToSevenHandSize)
	copy(before, players[0].cards)

	dt.applyExchange(0, []int{0, 1, 2, 3, 4})

	assert.Equal(t, DeuceToSevenHandSize, players[0].GetDrawCount(),
		"引けた枚数が要求より少ない (山切れで黙って打ち切られた)")
	after := players[0].cards
	require.Len(t, after, DeuceToSevenHandSize)
	for i := range before {
		assert.NotSame(t, before[i], after[i],
			"捨てたはずの札がそのまま手元に残っている (index %d)", i)
	}

	// 負のコントロール: 引いた札が自分の捨て札から戻ってきていないこと。
	for _, b := range before {
		for _, a := range after {
			assert.NotSame(t, b, a, "自分が今捨てた札を引き直している")
		}
	}
}

// **マックは保存して読み戻しても残る。** Worker は毎リクエストで盤を復元する
// ので、捨て札を JSON に載せ忘れると、復元のたびに「切り直せる札が無い」状態に
// 戻り、山切れの交換がまた黙って打ち切られる。フィールドを見比べるのではなく、
// **復元した盤で実際に交換して**確かめる (#6233)。
func TestDeuceToSeven_MuckSurvivesSaveRestore(t *testing.T) {
	dt, players := setupDeuceToSevenForHumanBet(1)
	dt.SetPhase(DeuceToSevenPhaseDraw)
	dt.applyExchange(1, []int{0, 1, 2, 3, 4})
	dt.applyExchange(2, []int{0, 1, 2, 3, 4})
	for dt.trumpCards.DrawCard() != nil {
	}
	require.NotEmpty(t, dt.muck, "マックに札が積まれていない")

	data, err := json.Marshal(dt)
	require.NoError(t, err)
	var restored DeuceToSeven
	require.NoError(t, json.Unmarshal(data, &restored))
	require.Equal(t, 0, restored.trumpCards.GetRemainingCount(), "復元後も山は空")

	// **札を額面で見比べてはいけない。** このフィクスチャは 4 席に同じ 5 枚を
	// 配るので、引いた札がたまたま同じ額面になるのは普通に起きる。復元後は
	// ポインタも別物なので実体比較もできない。見るべきなのは「5 枚とも引けた
	// こと」と「その 5 枚がマックから出たこと」。
	muckBefore := len(restored.muck)
	require.GreaterOrEqual(t, muckBefore, DeuceToSevenHandSize, "復元後のマックが足りない")
	restored.applyExchange(0, []int{0, 1, 2, 3, 4})

	assert.Equal(t, DeuceToSevenHandSize, restored.players[0].GetDrawCount(),
		"復元後に山切れの交換が打ち切られている (マックが JSON に載っていない)")
	assert.Equal(t, 0, restored.trumpCards.GetRemainingCount(), "山から引いている")
	// 引いた 5 枚がマックから出て、捨てた 5 枚がマックに戻る。
	assert.Equal(t, muckBefore, len(restored.muck), "マックの出入りが合わない")
	_ = players
}
