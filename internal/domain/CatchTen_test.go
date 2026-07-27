//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestCatchTen() *domain.CatchTen {
	players := []*domain.CatchTenPlayer{
		domain.NewCatchTenPlayer(true, 0),
		domain.NewCatchTenPlayer(false, 1),
		domain.NewCatchTenPlayer(false, 0),
		domain.NewCatchTenPlayer(false, 1),
	}
	return domain.NewCatchTen(domain.NewTrumpCardsShortDeck(), players, domain.DefaultCatchTenConfig())
}

func setupCatchTenPlayPhase(g *domain.CatchTen, currentIdx, leadIdx, trickNum int) {
	g.SetPhase(domain.CatchTenPhasePlay)
	g.SetCurrentPlayerIdx(currentIdx)
	g.SetLeadPlayerIdx(leadIdx)
	g.SetTrickNumber(trickNum)
}

func TestNewCatchTen(t *testing.T) {
	g := newTestCatchTen()
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.Equal(t, 0, g.GetRoundNumber())
}

func TestCatchTen_DeckIs36Distinct(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	// 4 players × 9 cards = 36 cards dealt, all distinct (suit+value)
	seen := map[[2]int]bool{}
	total := 0
	for pi := 0; pi < 4; pi++ {
		p := g.GetPlayer(pi)
		assert.Equal(t, 9, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			c := p.GetCard(i)
			key := [2]int{c.GetDesign(), c.GetValue()}
			assert.False(t, seen[key], "duplicate card %v", key)
			seen[key] = true
			// values must be A(1),6,7,8,9,10,J(11),Q(12),K(13)
			v := c.GetValue()
			assert.True(t, v == 1 || (v >= 6 && v <= 13), "unexpected value %d", v)
			total++
		}
	}
	assert.Equal(t, 36, total)
	assert.Len(t, seen, 36)
}

func TestCatchTen_Reset(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()

	assert.Equal(t, domain.CatchTenPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 1, g.GetTrickNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))

	trumpSuit := g.GetTrumpSuit()
	assert.True(t, trumpSuit >= domain.CardDesignSpade && trumpSuit <= domain.CardDesignDiamond)
}

func TestCatchTen_Reset_ClearsAllState(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhaseGameEnd)
	g.SetTeamScore(0, 30)
	g.SetTeamScore(1, 20)

	g.Reset()
	assert.Equal(t, domain.CatchTenPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))
}

func TestCatchTen_PlayerPlay(t *testing.T) {
	t.Run("valid play", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		setupCatchTenPlayPhase(g, 0, 0, 1)
		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, 8, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("game ended", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		setupCatchTenPlayPhase(g, 0, 0, 1)
		g.SetGameEndFlag(true)
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		g.SetPhase(domain.CatchTenPhaseTrickEnd)
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		setupCatchTenPlayPhase(g, 1, 0, 1)
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid card index", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		setupCatchTenPlayPhase(g, 0, 0, 1)
		err := g.PlayerPlay(99)
		assert.Error(t, err)
	})

	t.Run("must follow suit", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		setupCatchTenPlayPhase(g, 0, 1, 1)

		leadCard := domain.NewCard(domain.CardDesignHeart, 10, false)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: leadCard}})

		p := g.GetPlayer(0)
		p.Reset()
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

		err := g.PlayerPlay(1) // spade, must-follow violation
		assert.Error(t, err)
		err = g.PlayerPlay(0) // heart, ok
		assert.NoError(t, err)
	})
}

func TestCatchTen_CpuPlay(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	setupCatchTenPlayPhase(g, 1, 1, 1)
	initial := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, initial-1, g.GetPlayer(1).GetCardsSize())
}

func TestCatchTen_CpuPlay_SkipsWhenHuman(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	setupCatchTenPlayPhase(g, 0, 0, 1)
	initial := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, initial, g.GetPlayer(0).GetCardsSize())
}

// TestCatchTen_TrumpJackIsHighest verifies the special trump ordering: J beats A.
func TestCatchTen_TrumpJackIsHighest(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	// All trump: J should beat A, K, Q
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},  // A trump
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)}, // K trump
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 11, false)}, // J trump (highest)
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 12, false)}, // Q trump
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx()) // J of trump wins
}

// TestCatchTen_TrumpAceBeatsTen verifies A > 10 within trump (rank order J,A,K,Q,10,...).
func TestCatchTen_TrumpAceBeatsTen(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}, // 10 trump
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},  // A trump (beats 10)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},  // 9 trump
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 6, false)},  // 6 trump
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

// TestCatchTen_PlainSuitAceHigh verifies A high in non-trump suits.
func TestCatchTen_PlainSuitAceHigh(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade) // trump = spade, lead = heart
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // K heart
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},  // A heart (highest)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 10, false)}, // 10 heart
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 6, false)},  // 6 heart
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

// TestCatchTen_TrumpBeatsLeadSuit verifies any trump beats a non-trump lead.
func TestCatchTen_TrumpBeatsLeadSuit(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // A heart lead
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 6, false)}, // 6 trump (wins)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

// TestCatchTen_HonorCaptureScoring verifies trump J=11,10=10,A=4,K=3,Q=2 capture.
func TestCatchTen_HonorCaptureScoring(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	// Trump J (11) + trump 10 (10) + trump A (4) in a trick; winner takes 25 honors
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 11, false)}, // J trump (wins) honor 11
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}, // 10 trump honor 10
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},  // A trump honor 4
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // K heart honor 0
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 0, g.GetLeadPlayerIdx())
	assert.Equal(t, 25, g.GetPlayer(0).GetRoundScore()) // 11+10+4
}

func TestCatchTen_HonorKingQueen(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(1)
	g.SetPhase(domain.CatchTenPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 13, false)}, // K trump (wins) honor 3
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 12, false)}, // Q trump honor 2
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},  // 9 trump honor 0
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 6, false)},  // 6 trump honor 0
	}
	g.SetCurrentTrick(trick)
	g.ResolveTrick()
	assert.Equal(t, 0, g.GetLeadPlayerIdx())
	assert.Equal(t, 5, g.GetPlayer(0).GetRoundScore()) // 3+2
}

func TestCatchTen_ResolveTrick_LastTrick(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(domain.CatchTenHandSize) // last trick
	g.SetPhase(domain.CatchTenPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 6, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.CatchTenPhaseRoundEnd, g.GetPhase())
}

// TestCatchTen_PlayCardDoesNotAutoResolve is a regression test: playCard must
// NOT resolve a full trick by itself. The caller resolves it exactly once (the
// interactor's human path, or the runCpuTurns loop after a CPU play). If
// playCard auto-resolved, a CPU completing a trick inside the loop — which then
// also calls ResolveTrick — would double the winner's trick and honour counts.
func TestCatchTen_PlayCardDoesNotAutoResolve(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	setupCatchTenPlayPhase(g, 3, 0, 1) // seat 3 (a CPU) plays the 4th card
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})

	g.CpuPlay() // CPU at seat 3 completes the trick

	assert.Equal(t, domain.CatchTenPhaseTrickEnd, g.GetPhase())
	tricksBefore := 0
	for i := 0; i < domain.CatchTenPlayerCnt; i++ {
		tricksBefore += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 0, tricksBefore, "playCard must not resolve the trick itself")

	g.ResolveTrick()
	tricksAfter := 0
	for i := 0; i < domain.CatchTenPlayerCnt; i++ {
		tricksAfter += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 1, tricksAfter, "exactly one trick should be awarded after a single ResolveTrick")
}

func TestCatchTen_NextTrick(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.NextTrick()
	assert.Equal(t, domain.CatchTenPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Nil(t, g.GetCurrentTrick())
}

func TestCatchTen_ScoreRound(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhaseRoundEnd)
	// team 0 honors = 20 (player 0), team 1 honors = 10 (player 1)
	g.GetPlayer(0).SetRoundScore(20)
	g.GetPlayer(1).SetRoundScore(10)
	g.ScoreRound()
	assert.Equal(t, 20, g.GetTeamScore(0))
	assert.Equal(t, 10, g.GetTeamScore(1))
}

func TestCatchTen_ScoreRound_GameEnd(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhaseRoundEnd)
	g.SetTeamScore(0, 30)
	g.GetPlayer(0).SetRoundScore(11) // 30+11 = 41 reaches default limit
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.CatchTenPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, 41, g.GetTeamScore(0))
}

// TestCatchTen_ScoreRound_DrawTiebreak verifies the same-deal both-reach draw rule.
func TestCatchTen_ScoreRound_DrawTiebreak(t *testing.T) {
	t.Run("both reach, higher wins", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		g.SetPhase(domain.CatchTenPhaseRoundEnd)
		g.SetTeamScore(0, 40)
		g.SetTeamScore(1, 40)
		g.GetPlayer(0).SetRoundScore(11) // team0 -> 51
		g.GetPlayer(1).SetRoundScore(2)  // team1 -> 42
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, 0, g.GetWinnerTeam())
	})

	t.Run("both reach exactly equal -> draw", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		g.SetPhase(domain.CatchTenPhaseRoundEnd)
		g.SetTeamScore(0, 40)
		g.SetTeamScore(1, 40)
		g.GetPlayer(0).SetRoundScore(2)
		g.GetPlayer(1).SetRoundScore(2)
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, domain.CatchTenDrawTeam, g.GetWinnerTeam())
	})

	t.Run("team1 wins", func(t *testing.T) {
		g := newTestCatchTen()
		g.Reset()
		g.SetPhase(domain.CatchTenPhaseRoundEnd)
		g.SetTeamScore(0, 10)
		g.GetPlayer(1).SetRoundScore(31) // team1 -> 31, below; bump
		g.SetTeamScore(1, 40)
		g.GetPlayer(1).SetRoundScore(2)
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, 1, g.GetWinnerTeam())
	})
}

func TestCatchTen_NextRound(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhaseRoundEnd)
	g.SetRoundNumber(1)
	initialDealer := g.GetDealerIdx()
	g.NextRound()
	assert.Equal(t, domain.CatchTenPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, 1, g.GetTrickNumber())
	assert.Equal(t, (initialDealer+1)%4, g.GetDealerIdx())
	for i := 0; i < 4; i++ {
		assert.Equal(t, 9, g.GetPlayer(i).GetCardsSize())
	}
}

func TestCatchTen_NextRound_WrongPhase(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetPhase(domain.CatchTenPhasePlay)
	g.NextRound()
	assert.Equal(t, domain.CatchTenPhasePlay, g.GetPhase())
}

func TestCatchTen_IsHumanTurn(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())
}

func TestCatchTen_GetPlayer_OutOfBounds(t *testing.T) {
	g := newTestCatchTen()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestCatchTen_GetTeamScore_OutOfBounds(t *testing.T) {
	g := newTestCatchTen()
	assert.Equal(t, 0, g.GetTeamScore(-1))
	assert.Equal(t, 0, g.GetTeamScore(99))
}

func TestCatchTen_GetHint(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	setupCatchTenPlayPhase(g, 0, 0, 1)
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
	assert.NotEmpty(t, hint.Reason)
}

func TestCatchTen_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	setupCatchTenPlayPhase(g, 1, 1, 1)
	assert.Nil(t, g.GetHint())
}

func TestCatchTen_GetConfig_SetConfig(t *testing.T) {
	g := newTestCatchTen()
	cfg := domain.CatchTenConfig{CpuDifficulty: domain.CatchTenCpuDifficultyHard, PointLimit: 50}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
}

func TestCatchTen_GetActionLog(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	assert.NotEmpty(t, g.GetActionLog())
}

func TestCatchTen_DealerAndPlayerCntAccessors(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	assert.Equal(t, domain.CatchTenPlayerCnt, g.GetPlayerCnt())
	g.SetDealerIdx(2)
	assert.Equal(t, 2, g.GetDealerIdx())
}

func TestCatchTen_GetValidPlayIndices(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	setupCatchTenPlayPhase(g, 0, 0, 1)
	indices := g.GetValidPlayIndices(0)
	assert.Equal(t, 9, len(indices))
}

func TestCatchTen_MarshalUnmarshalJSON(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.CatchTen
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
}

// TestCatchTen_UnmarshalRejectsInvalidState tampers one field of a valid
// serialized game per case and asserts the unmarshal fails.
func TestCatchTen_UnmarshalRejectsInvalidState(t *testing.T) {
	base := func() map[string]json.RawMessage {
		g := newTestCatchTen()
		g.Reset()
		data, err := json.Marshal(g)
		require.NoError(t, err)
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &m))
		return m
	}

	cases := map[string]map[string]string{
		"bad phase":          {"ph": "99"},
		"bad trump suit":     {"ts": "9"},
		"bad current player": {"ci": "99"},
		"bad lead player":    {"li": "99"},
		"bad dealer":         {"di": "99"},
		"bad winner team":    {"wt": "99"},
		"bad config diff":    {"cf": `{"cd":99,"pl":41}`},
		"bad config limit":   {"cf": `{"cd":1,"pl":0}`},
	}

	for name, tamper := range cases {
		t.Run(name, func(t *testing.T) {
			m := base()
			for k, v := range tamper {
				m[k] = json.RawMessage(v)
			}
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored domain.CatchTen
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}
}

func TestCatchTen_UnmarshalRejectsOversizedSlice(t *testing.T) {
	// Build an action log with > catchTenMaxSliceLen entries.
	var b []byte
	b = append(b, []byte(`{"al":[`)...)
	for i := 0; i < 1001; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, []byte(`{"t":1,"p":-1,"a":"x","d":"y"}`)...)
	}
	b = append(b, []byte(`]}`)...)

	var restored domain.CatchTen
	assert.Error(t, json.Unmarshal(b, &restored))
}

func TestCatchTenConfig_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert.NoError(t, domain.DefaultCatchTenConfig().Validate())
	})
	t.Run("invalid difficulty", func(t *testing.T) {
		cfg := domain.CatchTenConfig{CpuDifficulty: 99, PointLimit: 41}
		assert.Error(t, cfg.Validate())
	})
	t.Run("invalid point limit", func(t *testing.T) {
		cfg := domain.CatchTenConfig{CpuDifficulty: domain.CatchTenCpuDifficultyNormal, PointLimit: 0}
		assert.Error(t, cfg.Validate())
	})
}

func TestCatchTenPlayer_Team(t *testing.T) {
	p := domain.NewCatchTenPlayer(true, 1)
	assert.Equal(t, 1, p.GetTeam())
	assert.True(t, p.GetIsHuman())
}

func TestCatchTenPlayer_ResetRound(t *testing.T) {
	p := domain.NewCatchTenPlayer(false, 0)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	p.SetRoundScore(10)
	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetRoundScore())
}

func TestCatchTenPlayer_MarshalUnmarshalJSON(t *testing.T) {
	p := domain.NewCatchTenPlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored domain.CatchTenPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 1, restored.GetTeam())
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Equal(t, 1, restored.GetTrickCount())
}

func TestCatchTen_FullRound(t *testing.T) {
	g := newTestCatchTen()
	g.Reset()
	playCatchTenRound(t, g)
	assert.Equal(t, domain.CatchTenPhaseRoundEnd, g.GetPhase())
	total := 0
	for i := 0; i < 4; i++ {
		total += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, domain.CatchTenHandSize, total)
}

func TestCatchTen_CpuDifficulties(t *testing.T) {
	for _, diff := range []domain.CatchTenCpuDifficulty{
		domain.CatchTenCpuDifficultyEasy,
		domain.CatchTenCpuDifficultyNormal,
		domain.CatchTenCpuDifficultyHard,
	} {
		g := newTestCatchTen()
		cfg := domain.DefaultCatchTenConfig()
		cfg.CpuDifficulty = diff
		g.SetConfig(cfg)
		g.Reset()
		setupCatchTenPlayPhase(g, 1, 1, 1)
		initial := g.GetPlayer(1).GetCardsSize()
		g.CpuPlay()
		assert.Equal(t, initial-1, g.GetPlayer(1).GetCardsSize())
	}
}

func TestCatchTen_FullGame_AllDifficulties(t *testing.T) {
	for _, diff := range []domain.CatchTenCpuDifficulty{
		domain.CatchTenCpuDifficultyEasy,
		domain.CatchTenCpuDifficultyNormal,
		domain.CatchTenCpuDifficultyHard,
	} {
		for iter := 0; iter < 20; iter++ {
			g := newTestCatchTen()
			cfg := domain.DefaultCatchTenConfig()
			cfg.CpuDifficulty = diff
			cfg.PointLimit = 10 // shorten so game ends quickly
			g.SetConfig(cfg)
			g.Reset()

			for round := 0; round < 50 && !g.GetGameEndFlag(); round++ {
				playCatchTenRound(t, g) // drives the deal to RoundEnd
				g.ScoreRound()
				if g.GetGameEndFlag() {
					break
				}
				g.NextRound()
			}
			assert.True(t, g.GetGameEndFlag())
		}
	}
}

// playCatchTenRound drives a full 9-trick round until the domain reaches
// RoundEnd (or GameEnd). Tricks resolve and the round scores automatically
// inside the domain as the last card of each trick / deal is played.
func playCatchTenRound(t *testing.T, g *domain.CatchTen) {
	t.Helper()
	for {
		switch g.GetPhase() {
		case domain.CatchTenPhasePlay:
			cur := g.GetCurrentPlayerIdx()
			if g.GetPlayer(cur).GetIsHuman() {
				for ci := 0; ci < g.GetPlayer(cur).GetCardsSize(); ci++ {
					if g.PlayerPlay(ci) == nil {
						break
					}
				}
			} else {
				g.CpuPlay()
			}
		case domain.CatchTenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		default: // RoundEnd or GameEnd
			return
		}
	}
}
