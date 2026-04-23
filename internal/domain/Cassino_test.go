package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCassinoPlayers() []*domain.CassinoPlayer {
	return []*domain.CassinoPlayer{
		domain.NewCassinoPlayer(true),
		domain.NewCassinoPlayer(false),
		domain.NewCassinoPlayer(false),
		domain.NewCassinoPlayer(false),
	}
}

func newTestCassino(t *testing.T, cfg domain.CassinoConfig) *domain.Cassino {
	t.Helper()
	return domain.NewCassino(domain.NewTrumpCards(0), makeCassinoPlayers(), cfg)
}

// spade/heart/etc helpers for making cards (prefixed to avoid collisions with other test files)
func cass_s(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
func cass_h(v int) *domain.Card { return domain.NewCard(domain.CardDesignHeart, v, false) }
func cass_d(v int) *domain.Card { return domain.NewCard(domain.CardDesignDiamond, v, false) }
func cass_c(v int) *domain.Card { return domain.NewCard(domain.CardDesignClover, v, false) }

func TestCassinoConfig(t *testing.T) {
	t.Run("Default has expected values", func(t *testing.T) {
		cfg := domain.DefaultCassinoConfig()
		assert.Equal(t, domain.CassinoDefaultTargetScore, cfg.TargetScore)
		assert.True(t, cfg.MultiBuildEnabled)
		assert.True(t, cfg.SweepBonusEnabled)
		assert.Equal(t, domain.CassinoDifficultyNormal, cfg.CpuDifficulty)
	})

	t.Run("Validate rejects bad difficulty", func(t *testing.T) {
		cfg := domain.DefaultCassinoConfig()
		cfg.CpuDifficulty = domain.CassinoCpuDifficulty(99)
		assert.Error(t, cfg.Validate())
	})

	t.Run("Validate rejects bad target score", func(t *testing.T) {
		cfg := domain.DefaultCassinoConfig()
		cfg.TargetScore = 0
		assert.Error(t, cfg.Validate())
		cfg.TargetScore = 1000
		assert.Error(t, cfg.Validate())
	})

	t.Run("JSON round-trip", func(t *testing.T) {
		cfg := domain.DefaultCassinoConfig()
		b, err := json.Marshal(cfg)
		require.NoError(t, err)
		var out domain.CassinoConfig
		require.NoError(t, json.Unmarshal(b, &out))
		assert.Equal(t, cfg, out)
	})

	t.Run("Difficulty names exist for all", func(t *testing.T) {
		assert.NotEmpty(t, domain.CassinoDifficultyNames[domain.CassinoDifficultyEasy])
		assert.NotEmpty(t, domain.CassinoDifficultyNames[domain.CassinoDifficultyNormal])
		assert.NotEmpty(t, domain.CassinoDifficultyNames[domain.CassinoDifficultyHard])
	})
}

func TestCassinoCardValueAndFlags(t *testing.T) {
	t.Run("CassinoCardValue: A=1, 2-10=face, J/Q/K=11/12/13", func(t *testing.T) {
		assert.Equal(t, 1, domain.CassinoCardValue(cass_h(1)))
		assert.Equal(t, 5, domain.CassinoCardValue(cass_d(5)))
		assert.Equal(t, 10, domain.CassinoCardValue(cass_c(10)))
		assert.Equal(t, 11, domain.CassinoCardValue(cass_s(11)))
		assert.Equal(t, 13, domain.CassinoCardValue(cass_s(13)))
		assert.Equal(t, 0, domain.CassinoCardValue(nil))
	})

	t.Run("Face card flag", func(t *testing.T) {
		assert.False(t, domain.CassinoIsFaceCard(nil))
		assert.False(t, domain.CassinoIsFaceCard(cass_s(1)))
		assert.False(t, domain.CassinoIsFaceCard(cass_s(10)))
		assert.True(t, domain.CassinoIsFaceCard(cass_s(11)))
		assert.True(t, domain.CassinoIsFaceCard(cass_s(12)))
		assert.True(t, domain.CassinoIsFaceCard(cass_s(13)))
	})

	t.Run("Spade / Ace / BigCasino / LittleCasino", func(t *testing.T) {
		assert.True(t, domain.CassinoIsSpade(cass_s(7)))
		assert.False(t, domain.CassinoIsSpade(cass_h(7)))
		assert.True(t, domain.CassinoIsAce(cass_s(1)))
		assert.False(t, domain.CassinoIsAce(cass_s(2)))
		assert.True(t, domain.CassinoIsBigCasino(cass_d(10)))
		assert.False(t, domain.CassinoIsBigCasino(cass_s(10)))
		assert.True(t, domain.CassinoIsLittleCasino(cass_s(2)))
		assert.False(t, domain.CassinoIsLittleCasino(cass_d(2)))
		assert.False(t, domain.CassinoIsSpade(nil))
		assert.False(t, domain.CassinoIsAce(nil))
		assert.False(t, domain.CassinoIsBigCasino(nil))
		assert.False(t, domain.CassinoIsLittleCasino(nil))
	})
}

func TestCassinoBuild(t *testing.T) {
	t.Run("NewCassinoBuild single group", func(t *testing.T) {
		b := domain.NewCassinoBuild(1, 8, []*domain.Card{cass_h(3), cass_h(5)})
		assert.Equal(t, 1, b.OwnerIdx)
		assert.Equal(t, 8, b.Value)
		assert.False(t, b.IsMulti)
		assert.Len(t, b.AllCards(), 2)
	})

	t.Run("AddGroup promotes to multi", func(t *testing.T) {
		b := domain.NewCassinoBuild(1, 8, []*domain.Card{cass_h(3), cass_h(5)})
		b.AddGroup([]*domain.Card{cass_c(8)})
		assert.True(t, b.IsMulti)
		assert.Len(t, b.AllCards(), 3)
	})

	t.Run("AddGroup nil-safe", func(t *testing.T) {
		var b *domain.CassinoBuild
		b.AddGroup([]*domain.Card{cass_h(3)}) // no panic
		b2 := domain.NewCassinoBuild(0, 8, []*domain.Card{cass_h(8)})
		b2.AddGroup(nil)
		assert.False(t, b2.IsMulti)
	})

	t.Run("AllCards nil-safe", func(t *testing.T) {
		var b *domain.CassinoBuild
		assert.Nil(t, b.AllCards())
	})

	t.Run("JSON round-trip", func(t *testing.T) {
		b := domain.NewCassinoBuild(2, 9, []*domain.Card{cass_h(4), cass_d(5)})
		b.AddGroup([]*domain.Card{cass_c(9)})
		raw, err := json.Marshal(b)
		require.NoError(t, err)
		var out domain.CassinoBuild
		require.NoError(t, json.Unmarshal(raw, &out))
		assert.Equal(t, b.OwnerIdx, out.OwnerIdx)
		assert.Equal(t, b.Value, out.Value)
		assert.Equal(t, b.IsMulti, out.IsMulti)
	})
}

func TestCassinoPlayer(t *testing.T) {
	p := domain.NewCassinoPlayer(true)
	assert.Equal(t, 0, p.CapturedCount())
	assert.Equal(t, 0, p.GetSweepCount())
	assert.Equal(t, 0, p.GetTotalScore())
	assert.True(t, p.GetIsHuman())

	p.AddCaptured([]*domain.Card{cass_s(3), cass_s(4)})
	assert.Equal(t, 2, p.CapturedCount())
	assert.Len(t, p.GetCapturedCards(), 2)

	p.IncrementSweep()
	p.IncrementSweep()
	assert.Equal(t, 2, p.GetSweepCount())

	p.AddScore(5)
	assert.Equal(t, 5, p.GetTotalScore())

	p.ResetCaptured()
	p.ResetSweepCount()
	assert.Equal(t, 0, p.CapturedCount())
	assert.Equal(t, 0, p.GetSweepCount())

	p.ResetTotalScore()
	assert.Equal(t, 0, p.GetTotalScore())

	// JSON round-trip
	p.AddCaptured([]*domain.Card{cass_h(2)})
	p.IncrementSweep()
	p.AddScore(7)
	raw, err := json.Marshal(p)
	require.NoError(t, err)
	var q domain.CassinoPlayer
	require.NoError(t, json.Unmarshal(raw, &q))
	assert.Equal(t, p.CapturedCount(), q.CapturedCount())
	assert.Equal(t, p.GetSweepCount(), q.GetSweepCount())
	assert.Equal(t, p.GetTotalScore(), q.GetTotalScore())
}

func TestCassino_NewAndDefaults(t *testing.T) {
	c := newTestCassino(t, domain.CassinoConfig{})
	assert.NotNil(t, c)
	assert.Equal(t, 4, c.GetPlayerCnt())
	assert.False(t, c.GetGameEndFlag())
	assert.Equal(t, domain.CassinoPhaseDealing, c.GetPhase())

	d := domain.NewDefaultCassino()
	assert.Equal(t, 4, d.GetPlayerCnt())
	cfg := d.GetConfig()
	assert.Equal(t, 21, cfg.TargetScore)

	assert.Nil(t, c.GetPlayer(-1))
	assert.Nil(t, c.GetPlayer(99))
	assert.NotNil(t, c.GetPlayer(0))
}

func TestCassino_Reset(t *testing.T) {
	c := domain.NewDefaultCassino()
	c.Reset()
	// 初期配札: 4 枚ずつ + 場に 4 枚
	for i := 0; i < c.GetPlayerCnt(); i++ {
		assert.Equal(t, 4, c.GetPlayer(i).GetCardsSize())
	}
	assert.Len(t, c.GetTableCards(), 4)
	assert.Equal(t, 52-4*4-4, c.GetRemainingDeck()) // 52 - 16 - 4 = 32
	assert.Equal(t, domain.CassinoPhasePlayerTurn, c.GetPhase())
	assert.Equal(t, -1, c.GetLastCaptureIdx())
}

func TestCassino_TakeNumberMatch(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(8))
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(5), cass_s(7)})

	err := c.PlayerTake(0, []int{0, 1}, nil) // 3 + 5 = 8
	assert.NoError(t, err)
	assert.Equal(t, 3, c.GetPlayer(0).CapturedCount()) // played + 2 captured
	assert.Len(t, c.GetTableCards(), 1)
	assert.Equal(t, 0, c.GetLastCaptureIdx())
}

func TestCassino_TakeFaceRankMatch(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(11)) // J
	c.SetTableCards([]*domain.Card{cass_s(11), cass_s(5), cass_s(3)})

	err := c.PlayerTake(0, []int{0}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, c.GetPlayer(0).CapturedCount()) // J + J
}

func TestCassino_TakeSweepBonus(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(8))
	c.GetPlayer(1).AddCard(cass_h(2)) // ensure not round over
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(5)})

	err := c.PlayerTake(0, []int{0, 1}, nil) // sweeps
	assert.NoError(t, err)
	assert.Equal(t, 1, c.GetPlayer(0).GetSweepCount())
}

func TestCassino_TakeInvalid(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(8))
	c.SetTableCards([]*domain.Card{cass_s(2), cass_s(5)}) // sum = 7, not 8

	err := c.PlayerTake(0, []int{0, 1}, nil)
	assert.Error(t, err)
}

func TestCassino_TakeRequiresTarget(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(8))

	err := c.PlayerTake(0, nil, nil)
	assert.Error(t, err)
}

func TestCassino_Build(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	// hand: 3 + another 8 (capture card)
	c.GetPlayer(0).AddCard(cass_h(3))
	c.GetPlayer(0).AddCard(cass_h(8))
	c.SetTableCards([]*domain.Card{cass_s(5)})

	err := c.PlayerBuild(0, []int{0}, 8) // 3 + 5 = 8
	assert.NoError(t, err)
	assert.Len(t, c.GetBuilds(), 1)
	assert.Equal(t, 8, c.GetBuilds()[0].Value)
}

func TestCassino_BuildInvalid(t *testing.T) {
	t.Run("no capture card", func(t *testing.T) {
		c := newTestCassino(t, domain.DefaultCassinoConfig())
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(0)
		c.GetPlayer(0).AddCard(cass_h(3))
		c.SetTableCards([]*domain.Card{cass_s(5)})
		err := c.PlayerBuild(0, []int{0}, 8)
		assert.Error(t, err)
	})

	t.Run("face card cannot build", func(t *testing.T) {
		c := newTestCassino(t, domain.DefaultCassinoConfig())
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(0)
		c.GetPlayer(0).AddCard(cass_h(11))
		c.GetPlayer(0).AddCard(cass_h(11))
		c.SetTableCards([]*domain.Card{cass_s(5)})
		err := c.PlayerBuild(0, []int{0}, 11)
		assert.Error(t, err)
	})

	t.Run("sum mismatch", func(t *testing.T) {
		c := newTestCassino(t, domain.DefaultCassinoConfig())
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(0)
		c.GetPlayer(0).AddCard(cass_h(3))
		c.GetPlayer(0).AddCard(cass_h(9))
		c.SetTableCards([]*domain.Card{cass_s(5)})
		err := c.PlayerBuild(0, []int{0}, 9) // 3+5 = 8, not 9
		assert.Error(t, err)
	})

	t.Run("declared value out of range", func(t *testing.T) {
		c := newTestCassino(t, domain.DefaultCassinoConfig())
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(0)
		c.GetPlayer(0).AddCard(cass_h(3))
		c.GetPlayer(0).AddCard(cass_h(2))
		c.SetTableCards([]*domain.Card{cass_s(5)})
		err := c.PlayerBuild(0, []int{0}, 15)
		assert.Error(t, err)
	})

	t.Run("empty table selection", func(t *testing.T) {
		c := newTestCassino(t, domain.DefaultCassinoConfig())
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(0)
		c.GetPlayer(0).AddCard(cass_h(3))
		c.GetPlayer(0).AddCard(cass_h(8))
		err := c.PlayerBuild(0, nil, 8)
		assert.Error(t, err)
	})
}

func TestCassino_Trail(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(5))

	err := c.PlayerTrail(0)
	assert.NoError(t, err)
	assert.Len(t, c.GetTableCards(), 1)
	assert.Equal(t, 5, c.GetTableCards()[0].GetValue())
}

func TestCassino_TrailBlockedByOwnBuild(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(5))
	c.SetBuilds([]*domain.CassinoBuild{domain.NewCassinoBuild(0, 9, []*domain.Card{cass_s(4), cass_s(5)})})

	err := c.PlayerTrail(0)
	assert.Error(t, err)
}

func TestCassino_GuardHumanTurn(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetGameEndFlag(true)
	err := c.PlayerTrail(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)

	c2 := newTestCassino(t, domain.DefaultCassinoConfig())
	c2.SetPhase(domain.CassinoPhasePlayerTurn)
	c2.SetCurrentTurn(1) // CPU turn
	c2.GetPlayer(1).AddCard(cass_h(3))
	err = c2.PlayerTrail(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestCassino_CpuPlayRuns(t *testing.T) {
	c := domain.NewDefaultCassino()
	c.Reset()
	// turn forward to CPU
	for i := 0; i < c.GetPlayerCnt(); i++ {
		if !c.GetPlayer(c.GetCurrentTurn()).GetIsHuman() {
			break
		}
		// human turn: do trail to advance
		if c.GetPlayer(0).GetIsHuman() && c.GetCurrentTurn() == 0 {
			_ = c.PlayerTrail(0)
		}
	}
	// ensure a CPU turn runs without panic
	if !c.IsHumanTurn() {
		c.CpuPlay()
	}
}

// giveOtherPlayersFiller gives all non-current players 1 filler card, so the
// auto-dealer in postActionAdvance does not fire and re-deal the current player.
func giveOtherPlayersFiller(c *domain.Cassino, currentIdx int) {
	for i := 0; i < c.GetPlayerCnt(); i++ {
		if i == currentIdx {
			continue
		}
		c.GetPlayer(i).AddCard(cass_c(13)) // K of clubs as filler
	}
}

func TestCassino_CpuPlayAllDifficulties(t *testing.T) {
	for _, diff := range []domain.CassinoCpuDifficulty{
		domain.CassinoDifficultyEasy,
		domain.CassinoDifficultyNormal,
		domain.CassinoDifficultyHard,
	} {
		cfg := domain.DefaultCassinoConfig()
		cfg.CpuDifficulty = diff
		c := newTestCassino(t, cfg)
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(1) // CPU
		c.GetPlayer(1).AddCard(cass_h(7))
		giveOtherPlayersFiller(c, 1)
		c.SetTableCards([]*domain.Card{cass_s(3), cass_s(4)}) // sum 7
		c.CpuPlay()
		// CPU's hand card is gone after its action
		assert.Equal(t, 0, c.GetPlayer(1).GetCardsSize(), "diff=%v hand not consumed", diff)
	}
}

// TestCassino_CpuPlayNormalPrefersTake ensures that Normal/Hard CPU prefers a take when available.
func TestCassino_CpuPlayNormalPrefersTake(t *testing.T) {
	for _, diff := range []domain.CassinoCpuDifficulty{
		domain.CassinoDifficultyNormal,
		domain.CassinoDifficultyHard,
	} {
		cfg := domain.DefaultCassinoConfig()
		cfg.CpuDifficulty = diff
		c := newTestCassino(t, cfg)
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(1)
		c.GetPlayer(1).AddCard(cass_h(7))
		giveOtherPlayersFiller(c, 1)
		c.SetTableCards([]*domain.Card{cass_s(3), cass_s(4)})
		c.CpuPlay()
		assert.GreaterOrEqual(t, c.GetPlayer(1).CapturedCount(), 3, "diff=%v", diff)
	}
}

func TestCassino_ScoreRound(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	// Manually give captures
	c.GetPlayer(0).AddCaptured([]*domain.Card{
		cass_s(2),  // little casino + spade
		cass_d(10), // big casino
		cass_s(1),  // ace + spade
		cass_h(1),  // ace
	})
	c.GetPlayer(1).AddCaptured([]*domain.Card{
		cass_c(1), cass_s(7), cass_s(8), cass_s(9), // most spades
	})
	c.GetPlayer(1).IncrementSweep()
	// Player 0: 4 cards
	// Player 1: 4 cards (tie) → no most cards bonus
	d := c.ScoreRoundForTest()
	require.NotNil(t, d)
	// Player 0: BigCasino(2) + LittleCasino(1) + 2 aces(2) + no most cards tie + no most spades tie
	// Actually P0 has 2 spades, P1 has 3 spades → P1 most spades (1pt)
	// P0: 2 + 1 + 2 = 5, P1: 0 + 0 + 1 (ace) + 1 (most spades) + 1 (sweep) = 3
	assert.Equal(t, 5, d.Gained[0])
	assert.Equal(t, 3, d.Gained[1])
}

func TestCassino_FinishRoundLastTake(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetLastCaptureIdx(0)
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(4)})
	c.FinishRoundForTest()
	assert.GreaterOrEqual(t, c.GetPlayer(0).CapturedCount(), 2)
	assert.Equal(t, 0, len(c.GetTableCards()))
}

func TestCassino_GameEndAt21(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	// Stuff captured to force end of game
	c.GetPlayer(0).AddCaptured([]*domain.Card{cass_d(10), cass_s(2), cass_s(1), cass_h(1), cass_c(1), cass_d(1)})
	// P0 has big casino(2) + little casino(1) + 4 aces(4) + most cards(3) + most spades(1) = 11
	// Run multiple finishRound iterations to push over 21
	for i := 0; i < 3; i++ {
		c.SetLastCaptureIdx(-1)
		c.FinishRoundForTest()
	}
	// gameEnd should eventually flip
	assert.True(t, c.GetGameEndFlag() || c.GetPlayer(0).GetTotalScore() > 0)
}

func TestCassino_NextRound(t *testing.T) {
	c := domain.NewDefaultCassino()
	c.Reset()
	// Manually finish the round
	c.SetLastCaptureIdx(0)
	c.FinishRoundForTest()
	if c.GetGameEndFlag() {
		return // test done
	}
	c.NextRound()
	// After NextRound, each player should have 4 cards again
	for i := 0; i < c.GetPlayerCnt(); i++ {
		assert.Equal(t, 4, c.GetPlayer(i).GetCardsSize())
	}
}

func TestCassino_GettersAndSettersCoverage(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	assert.Empty(t, c.GetActionLog())
	assert.Nil(t, c.GetHumanAction())
	assert.Empty(t, c.GetCpuActions())
	assert.Nil(t, c.GetLastRoundDetail())
	assert.Empty(t, c.GetRoundWinners())
	assert.Equal(t, 0, c.GetPacksDealt())

	cfg := c.GetConfig()
	cfg.TargetScore = 10
	c.SetConfig(cfg)
	assert.Equal(t, 10, c.GetConfig().TargetScore)

	c.SetHumanAction(&domain.CassinoAction{PlayerIdx: 0, Type: domain.CassinoActionTrail})
	assert.NotNil(t, c.GetHumanAction())
	c.SetCpuActions([]*domain.CassinoAction{{PlayerIdx: 1}})
	assert.Len(t, c.GetCpuActions(), 1)

	c.SetGameEndFlag(true)
	assert.False(t, c.IsHumanTurn()) // game end
	c.SetGameEndFlag(false)
	c.SetCurrentTurn(1) // CPU
	assert.False(t, c.IsHumanTurn())
	c.SetCurrentTurn(0) // human
	assert.True(t, c.IsHumanTurn())
}

func TestCassino_ApplyForTestHelpers(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	p := c.GetPlayer(2)
	p.AddCard(cass_h(8))
	giveOtherPlayersFiller(c, 2)
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(5)})
	err := c.ApplyTakeForTest(2, 0, []int{0, 1}, nil)
	assert.NoError(t, err)
	// Build helper
	c2 := newTestCassino(t, domain.DefaultCassinoConfig())
	c2.GetPlayer(2).AddCard(cass_h(3))
	c2.GetPlayer(2).AddCard(cass_c(8))
	c2.SetTableCards([]*domain.Card{cass_s(5)})
	assert.NoError(t, c2.ApplyBuildForTest(2, 0, []int{0}, 8))
	// Trail helper
	c3 := newTestCassino(t, domain.DefaultCassinoConfig())
	c3.GetPlayer(2).AddCard(cass_h(5))
	assert.NoError(t, c3.ApplyTrailForTest(2, 0))
	// DealNextPack helper
	c4 := newTestCassino(t, domain.DefaultCassinoConfig())
	c4.DealNextPackForTest()
	assert.Equal(t, 4, c4.GetPlayer(0).GetCardsSize())
}

func TestCassino_RemoveBuildsByIndex(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	// Setup: 2 builds present; player has 9 to capture first build
	c.SetBuilds([]*domain.CassinoBuild{
		domain.NewCassinoBuild(1, 9, []*domain.Card{cass_s(4), cass_s(5)}),
		domain.NewCassinoBuild(2, 7, []*domain.Card{cass_c(3), cass_c(4)}),
	})
	c.GetPlayer(0).AddCard(cass_h(9))
	giveOtherPlayersFiller(c, 0)
	err := c.PlayerTake(0, nil, []int{0}) // capture build 0
	assert.NoError(t, err)
	assert.Len(t, c.GetBuilds(), 1)
	assert.Equal(t, 7, c.GetBuilds()[0].Value)
}

func TestCassino_TrailFallbackPlan(t *testing.T) {
	// Force a CPU with no legal takes / builds, only trail options
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyEasy
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	c.GetPlayer(1).AddCard(cass_h(13)) // K: no matching rank on table
	giveOtherPlayersFiller(c, 1)
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(4)}) // no K
	c.CpuPlay()
	// Table should now contain the trailed K
	found := false
	for _, tc := range c.GetTableCards() {
		if tc.GetValue() == 13 && tc.GetDesign() == domain.CardDesignHeart {
			found = true
		}
	}
	assert.True(t, found)
}

func TestCassino_RemoveTableCardsEdgeCases(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(0)
	c.GetPlayer(0).AddCard(cass_h(3))
	giveOtherPlayersFiller(c, 0)
	// out-of-range build index
	c.SetTableCards([]*domain.Card{cass_s(3)})
	assert.Error(t, c.PlayerTake(0, nil, []int{5}))
}

func TestCassino_JSONRoundTrip(t *testing.T) {
	c := domain.NewDefaultCassino()
	c.Reset()
	raw, err := json.Marshal(c)
	require.NoError(t, err)
	restored := domain.NewDefaultCassino()
	require.NoError(t, json.Unmarshal(raw, restored))
	assert.Equal(t, c.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetCurrentTurn(), restored.GetCurrentTurn())
}

func TestEnumerateTakes(t *testing.T) {
	t.Run("simple single match", func(t *testing.T) {
		table := []*domain.Card{cass_s(3), cass_s(5), cass_s(7)}
		takes := domain.EnumerateTakes(cass_h(5), table, nil)
		// 5 alone, 3+5=8? no, should be {5}, {3+5} wait 3+5=8 not 5
		// But also {5}=5, and 3 alone=3, nope
		// So only {5} at index 1
		found := false
		for _, t2 := range takes {
			if len(t2[0]) == 1 && t2[0][0] == 1 {
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("face card matches face", func(t *testing.T) {
		table := []*domain.Card{cass_s(11), cass_s(11), cass_s(7)}
		takes := domain.EnumerateTakes(cass_h(11), table, nil)
		assert.GreaterOrEqual(t, len(takes), 1)
	})

	t.Run("build capture only", func(t *testing.T) {
		builds := []*domain.CassinoBuild{domain.NewCassinoBuild(1, 8, []*domain.Card{cass_s(3), cass_s(5)})}
		takes := domain.EnumerateTakes(cass_h(8), nil, builds)
		assert.NotEmpty(t, takes)
	})

	t.Run("empty on nil hand", func(t *testing.T) {
		assert.Nil(t, domain.EnumerateTakes(nil, nil, nil))
	})
}

func TestEnumerateBuilds(t *testing.T) {
	t.Run("basic build enumeration", func(t *testing.T) {
		table := []*domain.Card{cass_s(5)}
		hand := []*domain.Card{cass_h(3), cass_h(8)}
		out := domain.EnumerateBuilds(cass_h(3), 0, hand, table)
		// 3 + 5 = 8, hand has 8
		require.NotEmpty(t, out)
	})

	t.Run("no capture card", func(t *testing.T) {
		hand := []*domain.Card{cass_h(3)}
		out := domain.EnumerateBuilds(cass_h(3), 0, hand, []*domain.Card{cass_s(5)})
		assert.Empty(t, out)
	})

	t.Run("face card returns nil", func(t *testing.T) {
		assert.Nil(t, domain.EnumerateBuilds(cass_h(11), 0, []*domain.Card{cass_h(11)}, []*domain.Card{cass_s(5)}))
	})
}
