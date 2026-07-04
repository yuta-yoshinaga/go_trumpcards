//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// primeroCard は指定デザイン・値のカードを生成するテストヘルパー。
func primeroCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// primeroSetHand は player の手札を 4 枚に差し替える。
func primeroSetHand(p *domain.PrimeroPlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// primeroEvalHand は 4 枚を評価するショートカット。
func primeroEvalHand(cards ...*domain.Card) (int, []int) {
	return domain.PrimeroEval(cards)
}

func TestPrimero_ResetDealsRound(t *testing.T) {
	g := domain.NewDefaultPrimero()
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	cfg := g.GetConfig()
	// After the deal the CPUs left of the dealer bet before the human acts, so
	// the pot equals the sum of every player's round bet and is at least the antes.
	sumBets := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		sumBets += g.GetPlayer(i).GetRoundBet()
	}
	assert.Equal(t, sumBets, g.GetPot())
	assert.GreaterOrEqual(t, g.GetPot(), cfg.Ante*cfg.PlayerCount)
	assert.Equal(t, domain.PrimeroHandSize, g.GetPlayer(0).GetCardsSize())
	// Human paid the ante; before any of their own bets, roundBet = ante.
	assert.Equal(t, cfg.Ante, g.GetPlayer(0).GetRoundBet())
}

func TestPrimeroEval_Categories(t *testing.T) {
	// Fluxus: four spades. 7+6+A+5 = 21+18+16+15 = 70.
	cat, tb := primeroEvalHand(
		primeroCard(domain.CardDesignSpade, 7), primeroCard(domain.CardDesignSpade, 6),
		primeroCard(domain.CardDesignSpade, 1), primeroCard(domain.CardDesignSpade, 5))
	assert.Equal(t, domain.PrimeroHandFluxus, cat)
	assert.Equal(t, []int{70}, tb)

	// Supremus: one of each suit, sum 70 (>= 50).
	catS, tbS := primeroEvalHand(
		primeroCard(domain.CardDesignSpade, 7), primeroCard(domain.CardDesignClover, 6),
		primeroCard(domain.CardDesignHeart, 1), primeroCard(domain.CardDesignDiamond, 5))
	assert.Equal(t, domain.PrimeroHandSupremus, catS)
	assert.Equal(t, []int{70}, tbS)

	// Primero: one of each suit, all faces (10 each) → sum 40 (< 50).
	catP, tbP := primeroEvalHand(
		primeroCard(domain.CardDesignSpade, 11), primeroCard(domain.CardDesignClover, 12),
		primeroCard(domain.CardDesignHeart, 13), primeroCard(domain.CardDesignDiamond, 11))
	assert.Equal(t, domain.PrimeroHandPrimero, catP)
	assert.Equal(t, []int{40}, tbP)

	// Numerus: two spades (39) + two hearts (20). max suit 39, total 59.
	catN, tbN := primeroEvalHand(
		primeroCard(domain.CardDesignSpade, 7), primeroCard(domain.CardDesignSpade, 6),
		primeroCard(domain.CardDesignHeart, 11), primeroCard(domain.CardDesignHeart, 12))
	assert.Equal(t, domain.PrimeroHandNumerus, catN)
	assert.Equal(t, []int{39, 59}, tbN)

	// Category order: Fluxus > Supremus > Primero > Numerus.
	assert.Equal(t, 1, domain.PrimeroCompare(cat, tb, catS, tbS))
	assert.Equal(t, 1, domain.PrimeroCompare(catS, tbS, catP, tbP))
	assert.Equal(t, 1, domain.PrimeroCompare(catP, tbP, catN, tbN))
	assert.Equal(t, -1, domain.PrimeroCompare(catN, tbN, catP, tbP))

	// Fluxus tie-break by total: 70 beats 69.
	_, tbLow := primeroEvalHand(
		primeroCard(domain.CardDesignHeart, 7), primeroCard(domain.CardDesignHeart, 6),
		primeroCard(domain.CardDesignHeart, 1), primeroCard(domain.CardDesignHeart, 4))
	assert.Equal(t, 1, domain.PrimeroCompare(domain.PrimeroHandFluxus, tb, domain.PrimeroHandFluxus, tbLow))

	// Numerus secondary tie-break by total when max suit is equal.
	_, tbN2 := primeroEvalHand(
		primeroCard(domain.CardDesignSpade, 7), primeroCard(domain.CardDesignSpade, 6),
		primeroCard(domain.CardDesignHeart, 11), primeroCard(domain.CardDesignHeart, 3))
	// spade 39, heart 23, total 62 → beats {39, 59}.
	assert.Equal(t, 1, domain.PrimeroCompare(domain.PrimeroHandNumerus, tbN2, domain.PrimeroHandNumerus, tbN))

	// Exact tie is 0.
	assert.Equal(t, 0, domain.PrimeroCompare(cat, tb, domain.PrimeroHandFluxus, []int{70}))

	// Invalid size.
	badCat, badTb := domain.PrimeroEval([]*domain.Card{primeroCard(domain.CardDesignSpade, 7)})
	assert.Equal(t, -1, badCat)
	assert.Nil(t, badTb)
}

// primeroShowdownGame は 4 人ゲームを組み立て、全員をアクティブにして決定的な
// ショーダウンを準備する。
func primeroShowdownGame(t *testing.T) *domain.Primero {
	t.Helper()
	g := domain.NewDefaultPrimero()
	g.SetPhase(domain.PrimeroPhaseBetting)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(false)
		g.GetPlayer(i).SetOut(false)
	}
	return g
}

// primeroFluxus は同スート 4 枚 (Fluxus) を返す。
func primeroFluxus(design int) []*domain.Card {
	return []*domain.Card{
		primeroCard(design, 7), primeroCard(design, 6),
		primeroCard(design, 1), primeroCard(design, 5),
	}
}

// primeroWeakNumerus は弱いヌメルス (2 スート、絵札のみ) を返す。
func primeroWeakNumerus() []*domain.Card {
	return []*domain.Card{
		primeroCard(domain.CardDesignSpade, 11), primeroCard(domain.CardDesignSpade, 12),
		primeroCard(domain.CardDesignHeart, 13), primeroCard(domain.CardDesignHeart, 11),
	}
}

func TestPrimero_Showdown_FluxusBeatsNumerus(t *testing.T) {
	g := primeroShowdownGame(t)
	primeroSetHand(g.GetPlayer(0), primeroFluxus(domain.CardDesignSpade)...) // Fluxus
	primeroSetHand(g.GetPlayer(1), primeroWeakNumerus()...)                  // Numerus
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(100)
	before := g.GetPlayer(0).GetChips()

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.PrimeroResultWin, g.GetResult())
	assert.Equal(t, before+100, g.GetPlayer(0).GetChips())
	assert.Equal(t, domain.PrimeroPhaseResult, g.GetPhase())
}

func TestPrimero_Showdown_HumanLoses(t *testing.T) {
	g := primeroShowdownGame(t)
	primeroSetHand(g.GetPlayer(0), primeroWeakNumerus()...)                  // Numerus (weak)
	primeroSetHand(g.GetPlayer(1), primeroFluxus(domain.CardDesignHeart)...) // Fluxus (strong)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(50)

	g.ResolveForTest()

	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, domain.PrimeroResultLose, g.GetResult())
}

func TestPrimero_Showdown_TieGoesToEarliestSeat(t *testing.T) {
	g := primeroShowdownGame(t)
	// Two identical-strength Fluxus hands (different suits, same points).
	primeroSetHand(g.GetPlayer(0), primeroFluxus(domain.CardDesignSpade)...)
	primeroSetHand(g.GetPlayer(1), primeroFluxus(domain.CardDesignHeart)...)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(40)

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx(), "exact tie goes to the earliest seat")
}

func TestPrimero_CleanWin_SoleActive(t *testing.T) {
	g := primeroShowdownGame(t)
	g.GetPlayer(1).SetFolded(true)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(70)
	before := g.GetPlayer(0).GetChips()

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, before+70, g.GetPlayer(0).GetChips())
}

func TestPrimero_HumanFold_ResultNone(t *testing.T) {
	g := primeroShowdownGame(t)
	g.GetPlayer(0).SetFolded(true) // human folded
	primeroSetHand(g.GetPlayer(1), primeroFluxus(domain.CardDesignHeart)...)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(30)

	g.ResolveForTest()

	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, domain.PrimeroResultNone, g.GetResult())
}

func TestPrimero_PlayerRaise_FoldsAroundToCleanWin(t *testing.T) {
	g := domain.NewDefaultPrimero()
	cfg := g.GetConfig()
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()

	// Force a deterministic betting state: human to act, everyone matched at ante.
	g.SetPhase(domain.PrimeroPhaseBetting)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentBet(cfg.Ante)
	g.SetPot(cfg.Ante * 3)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		p.SetFolded(false)
		p.SetOut(false)
		p.SetRoundBet(cfg.Ante)
		p.SetChips(500)
	}
	// Human strong (Fluxus), CPUs weak (low numerus) so they fold to the raise.
	primeroSetHand(g.GetPlayer(0), primeroFluxus(domain.CardDesignSpade)...)
	primeroSetHand(g.GetPlayer(1), primeroWeakNumerus()...)
	primeroSetHand(g.GetPlayer(2), primeroWeakNumerus()...)
	// Reset() auto-ran the CPUs' opening bets, leaving stray raise/acted counters;
	// clear them so the human's single raise is the only one counted.
	g.ClearBettingForTest()

	require.True(t, g.IsHumanTurn())
	require.True(t, g.CanRaise())
	require.NoError(t, g.PlayerRaise())

	assert.Equal(t, 1, g.GetRaiseCount())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.PrimeroResultWin, g.GetResult())
	assert.Equal(t, domain.PrimeroPhaseResult, g.GetPhase())
}

func TestPrimero_PlayerCall_RealFlowReachesResult(t *testing.T) {
	g := domain.NewDefaultPrimero()
	// Drive the human seat by always calling until the round resolves.
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting; i++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerCall())
		} else {
			break
		}
	}
	assert.Equal(t, domain.PrimeroPhaseResult, g.GetPhase())
}

func TestPrimero_PlayerActions_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := primeroShowdownGame(t)
		g.GetPlayer(1).SetFolded(true)
		g.GetPlayer(2).SetFolded(true)
		g.GetPlayer(3).SetFolded(true)
		g.ResolveForTest() // → Result phase
		err := g.PlayerCall()
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
	t.Run("not human turn", func(t *testing.T) {
		g := primeroShowdownGame(t)
		g.SetCurrentPlayerIdx(1) // a CPU
		err := g.PlayerFold()
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})
	t.Run("raise cap", func(t *testing.T) {
		g := domain.NewDefaultPrimero()
		cfg := g.GetConfig()
		cfg.PlayerCount = 3
		g.SetConfig(cfg)
		g.Reset()
		g.SetPhase(domain.PrimeroPhaseBetting)
		g.SetCurrentPlayerIdx(0)
		for i := 0; i < g.GetPlayerCnt(); i++ {
			g.GetPlayer(i).SetFolded(false)
			g.GetPlayer(i).SetOut(false)
			g.GetPlayer(i).SetChips(500)
		}
		for g.GetRaiseCount() < domain.PrimeroMaxRaises && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn() && g.CanRaise() {
			if err := g.PlayerRaise(); err != nil {
				break
			}
		}
	})
}

func TestPrimero_NextRound(t *testing.T) {
	g := domain.NewDefaultPrimero()
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.Equal(t, domain.PrimeroPhaseResult, g.GetPhase())
	if g.GetGameEndFlag() {
		t.Skip("game ended in round 1")
	}
	g.NextRound()
	assert.Equal(t, domain.PrimeroPhaseBetting, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// NextRound is a no-op while in the betting phase.
	rn := g.GetRoundNumber()
	g.SetPhase(domain.PrimeroPhaseBetting)
	g.NextRound()
	assert.Equal(t, rn, g.GetRoundNumber())
}

func TestPrimero_GameEndByRounds(t *testing.T) {
	g := domain.NewDefaultPrimero()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// NextRound after game end is a no-op.
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestPrimero_Hint(t *testing.T) {
	g := primeroShowdownGame(t)
	g.SetCurrentPlayerIdx(0)
	primeroSetHand(g.GetPlayer(0), primeroFluxus(domain.CardDesignSpade)...) // Fluxus
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "raise", hint.Action)
	assert.Equal(t, "strong_hand", hint.Reason)

	primeroSetHand(g.GetPlayer(0), primeroWeakNumerus()...) // weak numerus
	g.GetPlayer(0).SetRoundBet(0)
	g.SetCurrentBet(10) // needs to pay → fold advice
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "fold", hint.Action)

	// No hint outside the human's betting turn.
	g.SetPhase(domain.PrimeroPhaseResult)
	assert.Nil(t, g.GetHint())
}

func TestPrimeroConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultPrimeroConfig().Validate())
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 1, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 7, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 4, Ante: 0, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 4, Ante: 10, StartingChips: 1, TargetRounds: 10}.Validate())
	// Starting chips passes the range check but is below the ante → immediate elimination.
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 4, Ante: 50, StartingChips: 10, TargetRounds: 10}.Validate())
	assert.Error(t, domain.PrimeroConfig{PlayerCount: 4, Ante: 10, StartingChips: 200, TargetRounds: 0}.Validate())
}

func TestPrimeroPlayer_JSON(t *testing.T) {
	p := domain.NewPrimeroPlayer(true, 300)
	p.SetFolded(true)
	p.AddRoundBet(20)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.PrimeroPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 300, got.GetChips())
	assert.True(t, got.GetFolded())
	assert.Equal(t, 20, got.GetRoundBet())
	assert.True(t, got.GetIsHuman())
}

func TestPrimeroPlayer_UnmarshalErrors(t *testing.T) {
	var p domain.PrimeroPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5,"rb":0}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"rb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p))
}

func TestPrimero_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultPrimero()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.Primero
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetRoundNumber(), got.GetRoundNumber())
	assert.Equal(t, g.GetWinnerIdx(), got.GetWinnerIdx())
	assert.Equal(t, g.GetResult(), got.GetResult())
	assert.Equal(t, g.GetPlayerCnt(), got.GetPlayerCnt())
	assert.Equal(t, g.GetCurrentPlayerIdx(), got.GetCurrentPlayerIdx())
}

func TestPrimero_UnmarshalValidation(t *testing.T) {
	cases := map[string]string{
		"not json":       `not json`,
		"invalid config": `{"cf":{"pc":9,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}]}`,
		"player mismatch": `{"cf":{"pc":4,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],` +
			`"ph":0,"rn":1}`,
		"too few players": `{"cf":{"pc":2,"an":10,"sc":200,"tr":10},"ps":[{"ch":1}],"ph":0,"rn":1}`,
		"invalid phase":   `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":9,"rn":1}`,
		"round zero":      `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":0}`,
		"negative pot":    `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1,"pt":-1}`,
		"dealer range":    `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1,"di":5}`,
		"current range":   `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1,"ci":5}`,
		"winner range":    `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1,"wi":5}`,
		"result range":    `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1,"re":9}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Primero
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestPrimero_UnmarshalDefaults(t *testing.T) {
	var g domain.Primero
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200},{"ch":200}],"ph":0,"rn":1}`), &g))
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.NotNil(t, g.GetActionLog())
}
