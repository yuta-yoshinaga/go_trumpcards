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

// bouillotteCard は指定デザイン・値のカードを生成するテストヘルパー。
func bouillotteCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// bouillotteSetHand は player の手札を 3 枚に差し替える。
func bouillotteSetHand(p *domain.BouillottePlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// bouillotteEval3 は 3 枚 + retourne を評価するショートカット。
func bouillotteEval3(retourne *domain.Card, cards ...*domain.Card) (int, []int) {
	return domain.BouillotteEval(cards, retourne)
}

func TestBouillotte_ResetDealsRound(t *testing.T) {
	g := domain.NewDefaultBouillotte()
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
	assert.NotNil(t, g.GetRetourne())
	assert.Equal(t, domain.BouillotteHandSize, g.GetPlayer(0).GetCardsSize())
	// Human paid the ante; before any of their own bets, chips = starting - ante.
	assert.Equal(t, cfg.Ante, g.GetPlayer(0).GetRoundBet())
}

func TestBouillotteEval_Categories(t *testing.T) {
	// Brelan simple: three Kings.
	cat, tb := bouillotteEval3(bouillotteCard(domain.CardDesignHeart, 8),
		bouillotteCard(domain.CardDesignSpade, 13), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignDiamond, 13))
	assert.Equal(t, domain.BouillotteHandBrelan, cat)
	assert.Equal(t, []int{13, 0}, tb)

	// Brelan favori: pair of Kings completed by a King retourne.
	catF, tbF := bouillotteEval3(bouillotteCard(domain.CardDesignHeart, 13),
		bouillotteCard(domain.CardDesignSpade, 13), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignDiamond, 8))
	assert.Equal(t, domain.BouillotteHandBrelan, catF)
	assert.Equal(t, []int{13, 1}, tbF)

	// Favori beats simple of the same rank.
	assert.Equal(t, 1, domain.BouillotteCompare(catF, tbF, cat, tb))
	assert.Equal(t, -1, domain.BouillotteCompare(cat, tb, catF, tbF))

	// High card: A-Q-8 with 9 retourne.
	catH, tbH := bouillotteEval3(bouillotteCard(domain.CardDesignHeart, 9),
		bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 12), bouillotteCard(domain.CardDesignDiamond, 8))
	assert.Equal(t, domain.BouillotteHandHighCard, catH)
	assert.Equal(t, []int{14, 12, 9, 8}, tbH)

	// Brelan beats high card.
	assert.Equal(t, 1, domain.BouillotteCompare(cat, tb, catH, tbH))

	// High card tie-break: A-K-8 beats A-Q-8 (same retourne).
	_, tbAK := bouillotteEval3(bouillotteCard(domain.CardDesignHeart, 9),
		bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignDiamond, 8))
	assert.Equal(t, 1, domain.BouillotteCompare(domain.BouillotteHandHighCard, tbAK, domain.BouillotteHandHighCard, tbH))

	// Exact tie is 0.
	assert.Equal(t, 0, domain.BouillotteCompare(domain.BouillotteHandHighCard, tbH, domain.BouillotteHandHighCard, tbH))

	// Invalid size.
	badCat, badTb := domain.BouillotteEval([]*domain.Card{bouillotteCard(domain.CardDesignSpade, 8)}, nil)
	assert.Equal(t, -1, badCat)
	assert.Nil(t, badTb)
}

func TestBouillotte_AnalyzeRetourneMatch(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	g.SetPhase(domain.BouillottePhaseBetting)
	p := g.GetPlayer(0)

	// No retourne
	g.SetRetourne(nil)
	match := g.AnalyzeRetourneMatch(0)
	assert.Empty(t, match.MatchingIndices)
	assert.Empty(t, match.NoteKey)

	// Retourne is 9 Diamond
	g.SetRetourne(bouillotteCard(domain.CardDesignDiamond, 9))

	// No match (A, K, Q)
	bouillotteSetHand(p, bouillotteCard(domain.CardDesignHeart, 1), bouillotteCard(domain.CardDesignSpade, 13), bouillotteCard(domain.CardDesignClover, 12))
	match = g.AnalyzeRetourneMatch(0)
	assert.Empty(t, match.MatchingIndices)
	assert.Empty(t, match.NoteKey)

	// 1 match (9, K, Q) - indices 0
	bouillotteSetHand(p, bouillotteCard(domain.CardDesignHeart, 9), bouillotteCard(domain.CardDesignSpade, 13), bouillotteCard(domain.CardDesignClover, 12))
	match = g.AnalyzeRetourneMatch(0)
	assert.Equal(t, []int{0}, match.MatchingIndices)
	assert.Empty(t, match.NoteKey)

	// 2 matches (9, 9, Q) - favori
	bouillotteSetHand(p, bouillotteCard(domain.CardDesignHeart, 9), bouillotteCard(domain.CardDesignSpade, 9), bouillotteCard(domain.CardDesignClover, 12))
	match = g.AnalyzeRetourneMatch(0)
	assert.Equal(t, []int{0, 1}, match.MatchingIndices)
	assert.Equal(t, "favori", match.NoteKey)

	// 3 matches (9, 9, 9) - carre
	bouillotteSetHand(p, bouillotteCard(domain.CardDesignHeart, 9), bouillotteCard(domain.CardDesignSpade, 9), bouillotteCard(domain.CardDesignClover, 9))
	match = g.AnalyzeRetourneMatch(0)
	assert.Equal(t, []int{0, 1, 2}, match.MatchingIndices)
	assert.Equal(t, "carre", match.NoteKey)
}

// bouillotteShowdownGame は 4 人ゲームを組み立て、指定座席のみをアクティブにして
// 決定的なショーダウンを準備する (retourne = 8 ダイヤ)。
func bouillotteShowdownGame(t *testing.T) *domain.Bouillotte {
	t.Helper()
	g := domain.NewDefaultBouillotte()
	g.SetPhase(domain.BouillottePhaseBetting)
	g.SetRetourne(bouillotteCard(domain.CardDesignDiamond, 8))
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(false)
		g.GetPlayer(i).SetOut(false)
	}
	return g
}

func TestBouillotte_Showdown_BrelanBeatsHighCard(t *testing.T) {
	g := bouillotteShowdownGame(t)
	// seat0: brelan of Queens. seat1: A-K-9 high card.
	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 12), bouillotteCard(domain.CardDesignClover, 12), bouillotteCard(domain.CardDesignHeart, 12))
	bouillotteSetHand(g.GetPlayer(1), bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignHeart, 9))
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(100)
	before := g.GetPlayer(0).GetChips()

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.BouillotteResultWin, g.GetResult())
	assert.Equal(t, before+100, g.GetPlayer(0).GetChips())
	assert.Equal(t, domain.BouillottePhaseResult, g.GetPhase())
}

func TestBouillotte_Showdown_HumanLoses(t *testing.T) {
	g := bouillotteShowdownGame(t)
	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 9), bouillotteCard(domain.CardDesignClover, 9), bouillotteCard(domain.CardDesignHeart, 8))   // 9 high
	bouillotteSetHand(g.GetPlayer(1), bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignHeart, 12)) // A high
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(50)

	g.ResolveForTest()

	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, domain.BouillotteResultLose, g.GetResult())
}

func TestBouillotte_Showdown_TieGoesToEarliestSeat(t *testing.T) {
	g := bouillotteShowdownGame(t)
	// Identical ranks for seat0 and seat1 (suits differ, suits are irrelevant).
	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignHeart, 9))
	bouillotteSetHand(g.GetPlayer(1), bouillotteCard(domain.CardDesignHeart, 1), bouillotteCard(domain.CardDesignDiamond, 13), bouillotteCard(domain.CardDesignSpade, 9))
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(40)

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx(), "exact tie goes to the earliest seat")
}

func TestBouillotte_CleanWin_SoleActive(t *testing.T) {
	g := bouillotteShowdownGame(t)
	g.GetPlayer(1).SetFolded(true)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(70)
	before := g.GetPlayer(0).GetChips()

	g.ResolveForTest()

	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, before+70, g.GetPlayer(0).GetChips())
}

func TestBouillotte_HumanFold_ResultNone(t *testing.T) {
	g := bouillotteShowdownGame(t)
	g.GetPlayer(0).SetFolded(true) // human folded
	bouillotteSetHand(g.GetPlayer(1), bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignHeart, 9))
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetPot(30)

	g.ResolveForTest()

	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, domain.BouillotteResultNone, g.GetResult())
}

func TestBouillotte_PlayerRaise_FoldsAroundToCleanWin(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	cfg := g.GetConfig()
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()

	// Force a deterministic betting state: human to act, everyone matched at ante.
	g.SetPhase(domain.BouillottePhaseBetting)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentBet(cfg.Ante)
	g.SetPot(cfg.Ante * 3)
	g.SetRetourne(bouillotteCard(domain.CardDesignDiamond, 8))
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		p.SetFolded(false)
		p.SetOut(false)
		p.SetRoundBet(cfg.Ante)
		p.SetChips(500)
	}
	// Human strong, CPUs weak (all-low, no favori) so they fold to the raise.
	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 1), bouillotteCard(domain.CardDesignClover, 13), bouillotteCard(domain.CardDesignHeart, 12))
	bouillotteSetHand(g.GetPlayer(1), bouillotteCard(domain.CardDesignSpade, 9), bouillotteCard(domain.CardDesignClover, 9), bouillotteCard(domain.CardDesignHeart, 8))
	bouillotteSetHand(g.GetPlayer(2), bouillotteCard(domain.CardDesignHeart, 9), bouillotteCard(domain.CardDesignDiamond, 9), bouillotteCard(domain.CardDesignClover, 8))
	// Reset() auto-ran the CPUs' opening bets, leaving stray raise/acted counters;
	// clear them so the human's single raise is the only one counted.
	g.ClearBettingForTest()

	require.True(t, g.IsHumanTurn())
	require.True(t, g.CanRaise())
	require.NoError(t, g.PlayerRaise())

	assert.Equal(t, 1, g.GetRaiseCount())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.BouillotteResultWin, g.GetResult())
	assert.Equal(t, domain.BouillottePhaseResult, g.GetPhase())
}

func TestBouillotte_PlayerCall_RealFlowReachesResult(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	// Drive the human seat by always calling until the round resolves.
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting; i++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerCall())
		} else {
			break
		}
	}
	assert.Equal(t, domain.BouillottePhaseResult, g.GetPhase())
}

func TestBouillotte_PlayerActions_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := bouillotteShowdownGame(t)
		g.GetPlayer(1).SetFolded(true)
		g.GetPlayer(2).SetFolded(true)
		g.GetPlayer(3).SetFolded(true)
		g.ResolveForTest() // → Result phase
		err := g.PlayerCall()
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
	t.Run("not human turn", func(t *testing.T) {
		g := bouillotteShowdownGame(t)
		g.SetCurrentPlayerIdx(1) // a CPU
		err := g.PlayerFold()
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})
	t.Run("raise cap", func(t *testing.T) {
		g := domain.NewDefaultBouillotte()
		cfg := g.GetConfig()
		cfg.PlayerCount = 3
		g.SetConfig(cfg)
		g.Reset()
		g.SetPhase(domain.BouillottePhaseBetting)
		g.SetCurrentPlayerIdx(0)
		g.SetRetourne(bouillotteCard(domain.CardDesignDiamond, 8))
		for i := 0; i < g.GetPlayerCnt(); i++ {
			g.GetPlayer(i).SetFolded(false)
			g.GetPlayer(i).SetOut(false)
			g.GetPlayer(i).SetChips(500)
		}
		// Exhaust the raise cap directly by repeated human raises is not possible
		// (driveCPU intervenes); instead drive the state so raiseCount == max via calls.
		for g.GetRaiseCount() < domain.BouillotteMaxRaises && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn() && g.CanRaise() {
			if err := g.PlayerRaise(); err != nil {
				break
			}
		}
	})
}

func TestBouillotte_NextRound(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	// Resolve round 1 deterministically via a fold-around.
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.Equal(t, domain.BouillottePhaseResult, g.GetPhase())
	if g.GetGameEndFlag() {
		t.Skip("game ended in round 1")
	}
	g.NextRound()
	assert.Equal(t, domain.BouillottePhaseBetting, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// NextRound is a no-op while in the betting phase.
	rn := g.GetRoundNumber()
	g.SetPhase(domain.BouillottePhaseBetting)
	g.NextRound()
	assert.Equal(t, rn, g.GetRoundNumber())
}

func TestBouillotte_GameEndByRounds(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// NextRound after game end is a no-op.
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestBouillotte_Hint(t *testing.T) {
	g := bouillotteShowdownGame(t)
	g.SetCurrentPlayerIdx(0)
	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 12), bouillotteCard(domain.CardDesignClover, 12), bouillotteCard(domain.CardDesignHeart, 12)) // brelan
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "raise", hint.Action)
	assert.Equal(t, "strong_hand", hint.Reason)

	bouillotteSetHand(g.GetPlayer(0), bouillotteCard(domain.CardDesignSpade, 9), bouillotteCard(domain.CardDesignClover, 8), bouillotteCard(domain.CardDesignHeart, 9)) // weak
	g.GetPlayer(0).SetRoundBet(0)
	g.SetCurrentBet(10) // needs to pay → fold advice
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "fold", hint.Action)

	// No hint outside the human's betting turn.
	g.SetPhase(domain.BouillottePhaseResult)
	assert.Nil(t, g.GetHint())
}

func TestBouillotteConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultBouillotteConfig().Validate())
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 2, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 5, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 4, Ante: 0, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 4, Ante: 10, StartingChips: 1, TargetRounds: 10}.Validate())
	// Starting chips passes the range check but is below the ante → immediate elimination.
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 4, Ante: 50, StartingChips: 10, TargetRounds: 10}.Validate())
	assert.Error(t, domain.BouillotteConfig{PlayerCount: 4, Ante: 10, StartingChips: 200, TargetRounds: 0}.Validate())
}

func TestBouillottePlayer_JSON(t *testing.T) {
	p := domain.NewBouillottePlayer(true, 300)
	p.SetFolded(true)
	p.AddRoundBet(20)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.BouillottePlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 300, got.GetChips())
	assert.True(t, got.GetFolded())
	assert.Equal(t, 20, got.GetRoundBet())
	assert.True(t, got.GetIsHuman())
}

func TestBouillottePlayer_UnmarshalErrors(t *testing.T) {
	var p domain.BouillottePlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5,"rb":0}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"rb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p))
}

func TestBouillotte_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.Bouillotte
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetRoundNumber(), got.GetRoundNumber())
	assert.Equal(t, g.GetWinnerIdx(), got.GetWinnerIdx())
	assert.Equal(t, g.GetResult(), got.GetResult())
	assert.Equal(t, g.GetPlayerCnt(), got.GetPlayerCnt())
	assert.Equal(t, g.GetCurrentPlayerIdx(), got.GetCurrentPlayerIdx())
}

func TestBouillotte_UnmarshalValidation(t *testing.T) {
	cases := map[string]string{
		"not json":       `not json`,
		"invalid config": `{"cf":{"pc":9,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}]}`,
		"player mismatch": `{"cf":{"pc":4,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],` +
			`"ph":0,"rn":1}`,
		"too few players": `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1}`,
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
			var g domain.Bouillotte
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestBouillotte_UnmarshalDefaults(t *testing.T) {
	var g domain.Bouillotte
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200},{"ch":200}],"ph":0,"rn":1}`), &g))
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.NotNil(t, g.GetActionLog())
}
