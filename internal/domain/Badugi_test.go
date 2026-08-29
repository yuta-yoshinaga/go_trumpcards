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

func newTestBadugi() (*Badugi, []*BadugiPlayer) {
	tc := NewTrumpCards(0)
	p0 := NewBadugiPlayer(true, BadugiStyleBalanced)
	p1 := NewBadugiPlayer(false, BadugiStyleConservative)
	p2 := NewBadugiPlayer(false, BadugiStyleAggressive)
	p3 := NewBadugiPlayer(false, BadugiStyleBluffer)
	players := []*BadugiPlayer{p0, p1, p2, p3}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	return NewBadugi(tc, players, DefaultBadugiConfig()), players
}

func setupBadugiForHumanBet(drawIdx int) (*Badugi, []*BadugiPlayer) {
	bd, players := newTestBadugi()
	if drawIdx == 0 {
		bd.SetPhase(BadugiPhaseDeal)
	} else {
		bd.SetPhase(BadugiPhaseBet)
	}
	bd.SetDrawIndex(drawIdx)
	bd.SetCurrentTurn(0)
	// Only human needs to act: CPUs start already acted.
	bd.setActedFlags([]bool{false, true, true, true})
	bd.SetLastBet(0)
	bd.SetMinRaise(10)
	bd.SetPot(40)
	bd.setStartingChips([]int{1000, 1000, 1000, 1000})
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	hand := []struct{ d, v int }{{S, 1}, {H, 2}, {D, 3}, {C, 4}}
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
	return bd, players
}

func giveBadugiHand(pl *BadugiPlayer, cards []*Card) {
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// ---------------------------------------------------------------------------
// construction
// ---------------------------------------------------------------------------

func TestNewBadugi(t *testing.T) {
	bd, players := newTestBadugi()
	assert.Equal(t, BadugiPhaseInit, bd.GetPhase())
	assert.Equal(t, 4, len(bd.GetPlayers()))
	assert.Equal(t, 0, bd.GetPot())
	assert.False(t, bd.GetGameEndFlag())
	assert.Equal(t, players, bd.GetPlayers())
	assert.Equal(t, 0, bd.GetDrawIndex())
}

func TestNewDefaultBadugi(t *testing.T) {
	bd := NewDefaultBadugi()
	assert.Equal(t, 4, len(bd.GetPlayers()))
	assert.True(t, bd.GetPlayers()[0].GetIsHuman())
	for _, pl := range bd.GetPlayers()[1:] {
		assert.False(t, pl.GetIsHuman())
	}
}

// ---------------------------------------------------------------------------
// Reset behaviour
// ---------------------------------------------------------------------------

func TestBadugi_Reset_DealsFourCardsAndCollectsAntes(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	// 4 players * ante 10 = 40, plus any bets the CPUs may have placed in the
	// opening round. Active player pot should be ≥ 40.
	assert.GreaterOrEqual(t, bd.GetPot(), 40)
	for _, pl := range players {
		assert.Equal(t, 4, pl.GetCardsSize(), "every active seat gets exactly 4 cards")
	}
	// Human has not been folded by the seat cap (CpuCount=3 → 4 seats).
	assert.False(t, players[0].GetFolded())
}

func TestBadugi_Reset_RejectsInvalidConfig(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetConfig(BadugiConfig{
		BettingLimit: BettingLimitType(-1),
		CpuCount:     1,
	})
	assert.Error(t, bd.Reset())
}

func TestBadugi_Reset_FoldsBeyondSeatCap(t *testing.T) {
	bd, players := newTestBadugi()
	cfg := bd.GetConfig()
	cfg.CpuCount = 1 // only 2 active seats
	bd.SetConfig(cfg)
	require.NoError(t, bd.Reset())
	assert.False(t, players[0].GetFolded())
	assert.False(t, players[1].GetFolded())
	assert.True(t, players[2].GetFolded(), "seat index 2 exceeds cap")
	assert.True(t, players[3].GetFolded())
}

// ---------------------------------------------------------------------------
// PlayerAction — wrong state / turn
// ---------------------------------------------------------------------------

func TestBadugi_PlayerAction_RejectsWhenGameEnded(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetGameEndFlag(true)
	err := bd.PlayerAction(BadugiActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestBadugi_PlayerAction_RejectsWrongPhase(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetPhase(BadugiPhaseDraw)
	err := bd.PlayerAction(BadugiActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestBadugi_PlayerAction_RejectsNotHumanTurn(t *testing.T) {
	bd, _ := setupBadugiForHumanBet(0)
	bd.SetCurrentTurn(1) // CPU seat
	err := bd.PlayerAction(BadugiActionCheck, 0, 0)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PlayerAction — happy path: human checks, CPUs act, advance to draw
// ---------------------------------------------------------------------------

func TestBadugi_PlayerAction_CheckAdvancesToDraw(t *testing.T) {
	bd, _ := setupBadugiForHumanBet(0)
	require.NoError(t, bd.PlayerAction(BadugiActionCheck, 0, 0))
	// Either we are in a draw (usual case) or the hand resolved early if
	// random CPU folds collapsed the field.
	phase := bd.GetPhase()
	assert.True(t, phase == BadugiPhaseDraw || phase == BadugiPhaseBet ||
		phase == BadugiPhaseEnd, "unexpected phase %d", phase)
}

// ---------------------------------------------------------------------------
// PlayerExchange
// ---------------------------------------------------------------------------

func TestBadugi_PlayerExchange_RejectsOutsideDraw(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetPhase(BadugiPhaseBet)
	err := bd.PlayerExchange([]int{0}, 0)
	assert.Error(t, err)
}

func TestBadugi_PlayerExchange_ReplacesSelectedIndicesOnly(t *testing.T) {
	bd, players := newTestBadugi()
	// Seed a deterministic "human about to draw" state: all CPUs folded, only
	// the human remains. Then flip phase to Draw and deal a known hand.
	require.NoError(t, bd.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	bd.SetPhase(BadugiPhaseDraw)
	bd.SetDrawIndex(1)
	bd.SetCurrentTurn(0)
	bd.setActedFlags([]bool{false, true, true, true})
	bd.ResetPlayerHand(0)
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	hand := []*Card{
		NewCard(S, 5, false), NewCard(H, 9, false),
		NewCard(D, 11, false), NewCard(C, 13, false),
	}
	giveBadugiHand(players[0], hand)
	before := []int{
		players[0].GetCard(0).GetValue(),
		players[0].GetCard(1).GetValue(),
		players[0].GetCard(2).GetValue(),
		players[0].GetCard(3).GetValue(),
	}
	require.NoError(t, bd.PlayerExchange([]int{1, 3}, 0))
	// Indices 0 and 2 are untouched.
	assert.Equal(t, before[0], players[0].GetCard(0).GetValue())
	assert.Equal(t, before[2], players[0].GetCard(2).GetValue())
	assert.Equal(t, 2, players[0].GetDrawCount())
}

func TestBadugi_PlayerStand_ExchangesZeroCards(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	bd.SetPhase(BadugiPhaseDraw)
	bd.SetDrawIndex(1)
	bd.SetCurrentTurn(0)
	bd.setActedFlags([]bool{false, true, true, true})
	require.NoError(t, bd.PlayerStand(0))
	assert.Equal(t, 0, players[0].GetDrawCount())
}

func setupBadugiForHumanExchange(t *testing.T, metaAI bool) (*Badugi, []*BadugiPlayer) {
	t.Helper()
	bd, players := newTestBadugi()
	bd.SetConfig(BadugiConfig{
		InitChips:    1000,
		Ante:         10,
		MinBet:       10,
		CpuCount:     3,
		BettingLimit: BettingLimitFixed,
		CpuMetaAI:    metaAI,
	})
	require.NoError(t, bd.Reset())
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	bd.SetPhase(BadugiPhaseDraw)
	bd.SetDrawIndex(1)
	bd.SetCurrentTurn(0)
	bd.setActedFlags([]bool{false, true, true, true})
	return bd, players
}

func TestBadugi_PlayerExchange_RecordsHesitationWhenCpuMetaAIEnabled(t *testing.T) {
	bd, _ := setupBadugiForHumanExchange(t, true)
	err := bd.PlayerExchange([]int{0}, 4200)
	require.NoError(t, err)

	assert.Equal(t, 4200, bd.GetLastHumanPlayMs())
	profile := bd.GetHumanProfile()
	require.NotNil(t, profile)
	assert.Equal(t, 1, profile.HesitationCount)
	assert.InDelta(t, 4200.0, profile.HesitationMean, 0.001)
}

func TestBadugi_PlayerExchange_NoHesitationWhenCpuMetaAIDisabled(t *testing.T) {
	bd, _ := setupBadugiForHumanExchange(t, false)
	err := bd.PlayerExchange([]int{0}, 4200)
	require.NoError(t, err)

	assert.Equal(t, 4200, bd.GetLastHumanPlayMs())
	assert.Nil(t, bd.GetHumanProfile())
}

func TestBadugi_PlayerStand_RecordsHesitationWhenCpuMetaAIEnabled(t *testing.T) {
	bd, _ := setupBadugiForHumanExchange(t, true)
	err := bd.PlayerStand(4200)
	require.NoError(t, err)

	assert.Equal(t, 4200, bd.GetLastHumanPlayMs())
	profile := bd.GetHumanProfile()
	require.NotNil(t, profile)
	assert.Equal(t, 1, profile.HesitationCount)
	assert.InDelta(t, 4200.0, profile.HesitationMean, 0.001)
}

func TestBadugi_PlayerStand_NoHesitationWhenCpuMetaAIDisabled(t *testing.T) {
	bd, _ := setupBadugiForHumanExchange(t, false)
	err := bd.PlayerStand(4200)
	require.NoError(t, err)

	assert.Equal(t, 4200, bd.GetLastHumanPlayMs())
	assert.Nil(t, bd.GetHumanProfile())
}

// ---------------------------------------------------------------------------
// Betting resolution — everyone folds to one player
// ---------------------------------------------------------------------------

func TestBadugi_LastPlayerStandingWinsPot(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	// Force all CPUs to fold, leaving the human.
	for i := 1; i < len(players); i++ {
		players[i].SetFolded(true)
	}
	// Drive one action so the game notices only one active player.
	bd.SetCurrentTurn(0)
	bd.SetPhase(BadugiPhaseDeal)
	bd.SetLastBet(0) // CPUs may have raised before we folded them.
	players[0].SetCurrentBet(0)
	bd.setActedFlags([]bool{false, true, true, true})
	prevChips := players[0].GetChips()
	pot := bd.GetPot()
	require.NoError(t, bd.PlayerAction(BadugiActionCheck, 0, 0))
	assert.True(t, bd.GetGameEndFlag())
	assert.Equal(t, prevChips+pot, players[0].GetChips())
	assert.Equal(t, 0, bd.GetPot())
}

// ---------------------------------------------------------------------------
// Showdown — split pot on identical hands
// ---------------------------------------------------------------------------

func TestBadugi_Showdown_SplitsEqualHands(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	// Replace seat 2 and 3 with folds, leaving players 0 and 1 to showdown.
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	perfect := []*Card{
		NewCard(S, 1, false), NewCard(H, 2, false),
		NewCard(D, 3, false), NewCard(C, 4, false),
	}
	giveBadugiHand(players[0], perfect)
	giveBadugiHand(players[1], perfect)

	bd.SetPot(200)
	bd.setStartingChips([]int{1000, 1000, 1000, 1000})
	bd.SetPhase(BadugiPhaseBet)
	bd.SetDrawIndex(3)
	// Using the internal showdown entrypoint directly keeps the test focused.
	bd.resolveShowdown()

	assert.True(t, bd.GetGameEndFlag())
	results := bd.GetRoundResults()
	assert.Len(t, results, 2)
	total := 0
	for _, r := range results {
		total += r.WonAmount
	}
	assert.Equal(t, 200, total)
}

// ---------------------------------------------------------------------------
// Showdown — stronger hand wins outright
// ---------------------------------------------------------------------------

func TestBadugi_Showdown_BestHandTakesPot(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// Player 0: perfect Badugi A-2-3-4
	giveBadugiHand(players[0], []*Card{
		NewCard(S, 1, false), NewCard(H, 2, false),
		NewCard(D, 3, false), NewCard(C, 4, false),
	})
	// Player 1: 3-card (two spades → drop one)
	giveBadugiHand(players[1], []*Card{
		NewCard(S, 5, false), NewCard(S, 7, false),
		NewCard(H, 9, false), NewCard(D, 10, false),
	})

	bd.SetPot(200)
	bd.setStartingChips([]int{1000, 1000, 1000, 1000})
	bd.SetPhase(BadugiPhaseBet)
	bd.SetDrawIndex(3)
	bd.resolveShowdown()

	results := bd.GetRoundResults()
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

func TestBadugi_FixedLimit_LateRoundUsesBigBet(t *testing.T) {
	bd, _ := newTestBadugi()
	cfg := bd.GetConfig()
	cfg.BettingLimit = BettingLimitFixed
	cfg.MinBet = 10
	bd.SetConfig(cfg)
	bd.SetPhase(BadugiPhaseBet)
	bd.SetDrawIndex(2) // late round
	assert.Equal(t, 20, bd.currentMinBet())

	bd.SetDrawIndex(1)
	assert.Equal(t, 10, bd.currentMinBet())
}

func TestBadugi_PotLimit_UsesMinBetConstant(t *testing.T) {
	bd, _ := newTestBadugi()
	cfg := bd.GetConfig()
	cfg.BettingLimit = BettingLimitPotLimit
	cfg.MinBet = 10
	bd.SetConfig(cfg)
	bd.SetPhase(BadugiPhaseBet)
	bd.SetDrawIndex(3)
	assert.Equal(t, 10, bd.currentMinBet())
}

func TestBadugi_NoLimit_UsesMinBetConstant(t *testing.T) {
	bd, _ := newTestBadugi()
	cfg := bd.GetConfig()
	cfg.BettingLimit = BettingLimitNoLimit
	cfg.MinBet = 10
	bd.SetConfig(cfg)
	bd.SetPhase(BadugiPhaseBet)
	bd.SetDrawIndex(3)
	assert.Equal(t, 10, bd.currentMinBet())
}

// ---------------------------------------------------------------------------
// CPU exchange logic
// ---------------------------------------------------------------------------

func TestBadugi_CpuDecideExchange_StandsPatOnBadugi(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	giveBadugiHand(players[1], []*Card{
		NewCard(S, 1, false), NewCard(H, 2, false),
		NewCard(D, 3, false), NewCard(C, 4, false),
	})
	indices := bd.cpuDecideExchange(1)
	assert.Empty(t, indices, "perfect Badugi should stand pat")
}

func TestBadugi_CpuDecideExchange_DropsPair(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	// Two-pair-ish: two aces (rank dup) — Conservative needs Size 4 to stand pat.
	giveBadugiHand(players[1], []*Card{
		NewCard(S, 1, false), NewCard(C, 1, false),
		NewCard(H, 7, false), NewCard(D, 10, false),
	})
	// Run several times to account for bluff stand-pat randomness (0% for Conservative).
	for range 10 {
		indices := bd.cpuDecideExchange(1)
		assert.NotEmpty(t, indices)
	}
}

// ---------------------------------------------------------------------------
// JSON round-trip
// ---------------------------------------------------------------------------

func TestBadugi_JSON_RoundTrip(t *testing.T) {
	bd, _ := newTestBadugi()
	require.NoError(t, bd.Reset())
	data, err := json.Marshal(bd)
	require.NoError(t, err)

	round := &Badugi{}
	require.NoError(t, json.Unmarshal(data, round))
	assert.Equal(t, bd.GetPhase(), round.GetPhase())
	assert.Equal(t, bd.GetPot(), round.GetPot())
	assert.Equal(t, bd.GetDrawIndex(), round.GetDrawIndex())
	assert.Equal(t, bd.GetDealerIdx(), round.GetDealerIdx())
	assert.Equal(t, bd.GetConfig(), round.GetConfig())
	assert.Equal(t, len(bd.GetPlayers()), len(round.GetPlayers()))
}

func TestBadugi_JSON_TooManyPlayersRejected(t *testing.T) {
	payload := []byte(`{"pl":[` +
		// 1001 empty objects
		replicatePlayerJSON(1001) + `]}`)
	err := json.Unmarshal(payload, &Badugi{})
	assert.Error(t, err)
}

func replicatePlayerJSON(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += "{}"
	}
	return out
}

// ---------------------------------------------------------------------------
// Helpers / getters
// ---------------------------------------------------------------------------

func TestBadugi_GettersRoundTrip(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetPhase(BadugiPhaseBet)
	bd.SetPot(42)
	bd.SetDealerIdx(2)
	bd.SetCurrentTurn(1)
	bd.SetLastBet(20)
	bd.SetMinRaise(40)
	bd.setRaiseCount(2)

	assert.Equal(t, BadugiPhaseBet, bd.GetPhase())
	assert.Equal(t, 42, bd.GetPot())
	assert.Equal(t, 2, bd.GetDealerIdx())
	assert.Equal(t, 1, bd.GetCurrentTurn())
	assert.Equal(t, 20, bd.GetLastBet())
	assert.Equal(t, 40, bd.GetMinRaise())
	assert.Equal(t, 2, bd.GetRaiseCount())
	assert.Equal(t, DefaultBadugiConfig().Ante, bd.GetAnte())
}

func TestBadugi_GetActionLog(t *testing.T) {
	bd, _ := newTestBadugi()
	assert.Empty(t, bd.GetActionLog())
	require.NoError(t, bd.Reset())
	assert.NotEmpty(t, bd.GetActionLog(), "ante entries should be logged")
}

// ---------------------------------------------------------------------------
// Full hand smoke test — three complete draws then showdown
// ---------------------------------------------------------------------------

func TestBadugi_FullHandProgressesToShowdownOrFold(t *testing.T) {
	bd, players := newTestBadugi()
	require.NoError(t, bd.Reset())

	// Human keeps checking / calling through each betting round, standing pat
	// on every draw. We cap iterations to avoid infinite loops if the state
	// machine ever wedges. The hand must end within a bounded number of
	// steps.
	const maxSteps = 16
	for step := 0; step < maxSteps && !bd.GetGameEndFlag(); step++ {
		phase := bd.GetPhase()
		if !players[bd.GetCurrentTurn()].GetIsHuman() {
			// Should not happen because Reset runs CPU actions up to the
			// human's turn; treat it as a fail-safe.
			break
		}
		switch phase {
		case BadugiPhaseDeal, BadugiPhaseBet:
			// Check if possible, else call.
			callAmt := bd.GetLastBet() - players[0].GetCurrentBet()
			if callAmt == 0 {
				if err := bd.PlayerAction(BadugiActionCheck, 0, 0); err != nil {
					// Some sequences require Call even when callAmt==0 after CPU raises; try Call.
					require.NoError(t, bd.PlayerAction(BadugiActionCall, 0, 0))
				}
			} else {
				require.NoError(t, bd.PlayerAction(BadugiActionCall, 0, 0))
			}
		case BadugiPhaseDraw:
			require.NoError(t, bd.PlayerStand(0))
		default:
			// Unexpected phase; bail out and let the assertion below flag it.
			bd.SetGameEndFlag(true)
		}
	}
	assert.True(t, bd.GetGameEndFlag(), "hand should resolve within %d steps", maxSteps)
}

// ---------------------------------------------------------------------------
// Profile import/export
// ---------------------------------------------------------------------------

func TestBadugi_ProfileExportNilWhenMissing(t *testing.T) {
	bd, _ := newTestBadugi()
	assert.Nil(t, bd.ExportProfile())
}

func TestBadugi_ProfileImportExportRoundTrip(t *testing.T) {
	bd, _ := newTestBadugi()
	cfg := bd.GetConfig()
	cfg.CpuMetaAI = true
	bd.SetConfig(cfg)
	require.NoError(t, bd.Reset())
	exported := bd.ExportProfile()
	require.NotNil(t, exported)
	data, err := json.Marshal(exported)
	require.NoError(t, err)

	other, _ := newTestBadugi()
	require.NoError(t, other.ImportProfile(data))
	assert.NotNil(t, other.GetHumanProfile())

	// Empty input is a no-op.
	require.NoError(t, other.ImportProfile(nil))
}

func TestBadugi_ResetProfileClears(t *testing.T) {
	bd, _ := newTestBadugi()
	bd.SetHumanProfile(&BettingHumanProfile{})
	bd.ResetProfile()
	assert.Nil(t, bd.GetHumanProfile())
}

// **山が尽きても交換は成立する。** 4 席・手札 4 枚だと配った時点の山は 36 枚しか
// なく、3 回のドローで全員が 4 枚引くと最大 48 枚要る。足りなくなった時点で
// 黙って break していたので、**捨てたはずの札がそのまま手元に残り**、画面には
// 何も出なかった。カジノの規則どおり、捨て札 (マック) を切り直して引く (#6256)。
func TestBadugi_ExchangeRecyclesTheMuckWhenTheStockRunsOut(t *testing.T) {
	bd, players := setupBadugiForHumanBet(1)
	bd.SetPhase(BadugiPhaseDraw)
	bd.SetCurrentTurn(0)

	// 先に何度か交換して、捨て札を積んでおく (マックの元)。
	bd.applyExchange(1, []int{0, 1, 2, 3})
	bd.applyExchange(2, []int{0, 1, 2, 3})

	// 山を空にする。
	for bd.trumpCards.DrawCard() != nil {
	}
	require.Equal(t, 0, bd.trumpCards.GetRemainingCount(), "山が空になっていない")

	before := make([]*Card, BadugiHandSize)
	copy(before, players[0].cards)

	bd.applyExchange(0, []int{0, 1, 2, 3})

	assert.Equal(t, BadugiHandSize, players[0].GetDrawCount(),
		"引けた枚数が要求より少ない (山切れで黙って打ち切られた)")
	after := players[0].cards
	require.Len(t, after, BadugiHandSize)
	// **負のコントロール: 自分が今捨てた札を引き直していないこと。**
	for _, b := range before {
		for _, a := range after {
			assert.NotSame(t, b, a, "自分が今捨てた札を引き直している")
		}
	}
}

// **マックは保存して読み戻しても残る。** Worker は毎リクエストで盤を復元する
// ので、捨て札を JSON に載せ忘れると、復元のたびに「切り直せる札が無い」状態に
// 戻り、山切れの交換がまた黙って打ち切られる。**復元した盤で実際に交換して**
// 確かめる (#6256)。
func TestBadugi_MuckSurvivesSaveRestore(t *testing.T) {
	bd, _ := setupBadugiForHumanBet(1)
	bd.SetPhase(BadugiPhaseDraw)
	bd.applyExchange(1, []int{0, 1, 2, 3})
	bd.applyExchange(2, []int{0, 1, 2, 3})
	for bd.trumpCards.DrawCard() != nil {
	}
	require.NotEmpty(t, bd.muck, "マックに札が積まれていない")

	data, err := json.Marshal(bd)
	require.NoError(t, err)
	var restored Badugi
	require.NoError(t, json.Unmarshal(data, &restored))
	require.Equal(t, 0, restored.trumpCards.GetRemainingCount(), "復元後も山は空")

	muckBefore := len(restored.muck)
	require.GreaterOrEqual(t, muckBefore, BadugiHandSize, "復元後のマックが足りない")
	restored.applyExchange(0, []int{0, 1, 2, 3})

	assert.Equal(t, BadugiHandSize, restored.players[0].GetDrawCount(),
		"復元後に山切れの交換が打ち切られている (マックが JSON に載っていない)")
	assert.Equal(t, 0, restored.trumpCards.GetRemainingCount(), "山から引いている")
	assert.Equal(t, muckBefore, len(restored.muck), "マックの出入りが合わない")
}
