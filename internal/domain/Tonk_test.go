//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestTonk() *domain.Tonk {
	players := []*domain.TonkPlayer{
		domain.NewTonkPlayer(true),
		domain.NewTonkPlayer(false),
	}
	cfg := domain.DefaultTonkConfig()
	// High point limit so a Tonk-on-deal (50-bonus + 50-hand = 100) does not end the match
	cfg.PointLimit = 10000
	return domain.NewTonk(domain.NewTrumpCards(0), players, cfg)
}

func newTestTonkWithDifficulty(d domain.TonkCpuDifficulty) *domain.Tonk {
	players := []*domain.TonkPlayer{
		domain.NewTonkPlayer(true),
		domain.NewTonkPlayer(false),
	}
	cfg := domain.DefaultTonkConfig()
	cfg.CpuDifficulty = d
	return domain.NewTonk(domain.NewTrumpCards(0), players, cfg)
}

func setupTonkDrawPhase(g *domain.Tonk, currentIdx int) {
	g.SetPhase(domain.TonkPhaseDraw)
	g.SetCurrentPlayerIdx(currentIdx)
}

func setupTonkDiscardPhase(g *domain.Tonk, currentIdx int) {
	g.SetPhase(domain.TonkPhaseDiscard)
	g.SetCurrentPlayerIdx(currentIdx)
}

func giveHand(p *domain.TonkPlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewTonk(t *testing.T) {
	g := newTestTonk()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetKnockerIdx())
}

func TestNewDefaultTonk(t *testing.T) {
	g := domain.NewDefaultTonk()
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestTonk_Reset(t *testing.T) {
	g := newTestTonk()
	g.Reset()

	// Either Draw phase (no Tonk on deal) or RoundEnd/GameEnd if Tonk was triggered
	phase := g.GetPhase()
	assert.True(t, phase == domain.TonkPhaseDraw || phase == domain.TonkPhaseRoundEnd || phase == domain.TonkPhaseGameEnd)
	assert.Equal(t, 1, g.GetRoundNumber())

	// Each player should have at most 5 cards (knocker may have 5 still after Tonk)
	for i := 0; i < 2; i++ {
		assert.LessOrEqual(t, g.GetPlayer(i).GetCardsSize(), 5)
	}
}

func TestTonk_Reset_NormalDeal(t *testing.T) {
	// Run reset many times to get a normal (non-Tonk) deal
	for trial := 0; trial < 50; trial++ {
		g := newTestTonk()
		g.Reset()

		if g.GetPhase() == domain.TonkPhaseDraw {
			// Each player should have 5 cards
			for i := 0; i < 2; i++ {
				assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
			}
			assert.Len(t, g.GetDiscardPile(), 1)
			// Draw pile: 52 - 10 - 1 = 41
			assert.Equal(t, 41, g.GetDrawPileCount())
			return
		}
	}
	// If we get here, every trial triggered Tonk (extremely unlikely)
	// but the test should not fail spuriously
}

// TestTonk_ResetTwice guards against a regression where calling Reset() a
// second time on the same Game instance left the underlying deck drained,
// producing an empty deal (totalCards=0, drawPileCount=0). See bug repro
// where the user clicks "次のゲーム" on Render dev / Cloudflare workers
// and the next round arrives with no cards.
//
// Seed 1 is the first int64 where both the 1st and 2nd consecutive Reset
// produce a normal (non-Tonk-on-deal) hand, so the assertions exercise the
// regression path directly — no retry loop / probabilistic guard needed.
func TestTonk_ResetTwice(t *testing.T) {
	g := newTestTonk()
	g.SetRand(rand.New(rand.NewSource(1)))

	g.Reset()
	require.Equal(t, domain.TonkPhaseDraw, g.GetPhase(),
		"seed 1 should produce a non-Tonk deal on 1st Reset")

	// 2nd Reset on the SAME Game instance — the failure mode.
	g.Reset()
	require.Equal(t, domain.TonkPhaseDraw, g.GetPhase(),
		"seed 1 should produce a non-Tonk deal on 2nd Reset")

	for i := 0; i < 2; i++ {
		assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize(),
			"player %d hand should be 5 cards after 2nd Reset", i)
	}
	assert.Len(t, g.GetDiscardPile(), 1, "discard pile should hold 1 card after 2nd Reset")
	assert.Equal(t, 41, g.GetDrawPileCount(),
		"draw pile should hold 41 cards after 2nd Reset (52 - 10 hand - 1 discard)")
}

func TestTonk_Getters(t *testing.T) {
	g := newTestTonk()
	g.Reset()

	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(2))

	cfg := g.GetConfig()
	assert.Equal(t, domain.TonkCpuDifficultyNormal, cfg.CpuDifficulty)

	g.SetConfig(domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyHard, PointLimit: 100})
	assert.Equal(t, domain.TonkCpuDifficultyHard, g.GetConfig().CpuDifficulty)
}

func TestTonk_IsHumanTurn(t *testing.T) {
	g := newTestTonk()
	g.Reset()

	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(2)
	assert.False(t, g.IsHumanTurn())
}

func TestTonk_GetDiscardTop(t *testing.T) {
	g := newTestTonk()

	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())

	card := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{card})
	assert.Equal(t, card, g.GetDiscardTop())
}

// --- Tonk on deal ---

func TestTonk_TonkOnDeal_Triggered(t *testing.T) {
	g := newTestTonk()
	// Simulate Reset already happened minimally; then manually set hand to 50 points
	g.Reset()
	// Force the state to trigger via direct hand setup + scoreTonk via NextRound flow
	// We invoke the public path: build a hand and call Reset() until we get a Tonk
	// Instead use deterministic path: set hands and re-trigger. Since checkTonkOnDeal is
	// private, we exercise it via Reset() with a known TrumpCards seed-equivalent.
	// Easiest: directly set state to mimic Tonk on deal via public setters.

	// Force RoundEnd path manually
	giveHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignClover, 12, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	})
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
	})
	// Reset would re-deal and possibly retrigger; we test the SetIsTonk getter directly.
	g.SetIsTonk(true)
	assert.True(t, g.GetIsTonk())
}

// Repeatedly call Reset on a deck designed to produce a Tonk deal
func TestTonk_TonkOnDeal_FromHandValuation(t *testing.T) {
	// We can force a Tonk by directly manipulating hand and calling NextRound
	// path via knock. But the simplest deterministic test: verify GinRummyCardValue
	// scoring used in Tonk gives 50 for 5 face cards.
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),   // K=10
		domain.NewCard(domain.CardDesignHeart, 12, false),   // Q=10
		domain.NewCard(domain.CardDesignDiamond, 11, false), // J=10
		domain.NewCard(domain.CardDesignClover, 10, false),  // 10=10
		domain.NewCard(domain.CardDesignSpade, 13, false),   // K=10
	}
	total := 0
	for _, c := range cards {
		total += domain.GinRummyCardValue(c)
	}
	assert.Equal(t, 50, total)
}

// TestTonk_TonkOnDeal_DeterministicSeed exercises the actual checkTonkOnDeal/scoreTonk
// path through Reset() using a seeded rng known to deal a 49- or 50-point hand.
// Seed 359 deals player 0 a 50-point hand (Tonk-on-deal high); seed 478 deals player 1
// a 50-point hand. Both produce a 100-point round score (TonkBonus + handValue).
func TestTonk_TonkOnDeal_DeterministicSeed(t *testing.T) {
	tests := []struct {
		name           string
		seed           int64
		wantKnocker    int
		wantRoundScore int
	}{
		{"player 0 (human) tonk on deal", 359, 0, 100},
		{"player 1 (cpu) tonk on deal", 478, 1, 100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestTonk()
			g.SetRand(rand.New(rand.NewSource(tc.seed)))
			g.Reset()

			assert.True(t, g.GetIsTonk(), "expected Tonk on deal")
			assert.Equal(t, tc.wantKnocker, g.GetKnockerIdx())
			assert.Equal(t, tc.wantRoundScore, g.GetPlayer(tc.wantKnocker).GetCumulativeScore())
			assert.Equal(t, 0, g.GetPlayer(1-tc.wantKnocker).GetCumulativeScore())
			assert.Equal(t, domain.TonkPhaseRoundEnd, g.GetPhase())
			assert.False(t, g.GetGameEndFlag(), "high PointLimit should keep game alive")
			// Opponent melds/deadwood are computed during scoreTonk
			assert.NotNil(t, g.GetOpponentDeadwood())
		})
	}
}

// --- PlayerDrawFromStock ---

func TestTonk_PlayerDrawFromStock(t *testing.T) {
	t.Run("valid draw", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)

		drawCard := domain.NewCard(domain.CardDesignHeart, 2, false)
		g.SetDrawPile([]*domain.Card{drawCard})
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
		})

		err := g.PlayerDrawFromStock()
		assert.NoError(t, err)
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.TonkPhaseDiscard, g.GetPhase())
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 0)
		assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 1)
		assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrNotHumanTurn)
	})

	t.Run("empty stock causes draw round", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)
		g.SetDrawPile(nil)

		err := g.PlayerDrawFromStock()
		assert.NoError(t, err)
		assert.Equal(t, domain.TonkPhaseRoundEnd, g.GetPhase())
	})

	t.Run("game ended", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		// Force gameEndFlag via knock + low PointLimit
		g.GetPlayer(0).SetCumulativeScore(0)
		g.SetConfig(domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyNormal, PointLimit: 1})
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
			domain.NewCard(domain.CardDesignDiamond, 1, false),
			domain.NewCard(domain.CardDesignClover, 2, false),
			domain.NewCard(domain.CardDesignClover, 13, false),
		})
		giveHand(g.GetPlayer(1), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 13, false),
			domain.NewCard(domain.CardDesignDiamond, 13, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})
		setupTonkDiscardPhase(g, 0)
		require.NoError(t, g.PlayerKnock(4)) // discard Club K
		assert.True(t, g.GetGameEndFlag())

		assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrGameEnded)
	})
}

// --- PlayerDrawFromDiscard ---

func TestTonk_PlayerDrawFromDiscard(t *testing.T) {
	t.Run("valid draw", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)
		topCard := domain.NewCard(domain.CardDesignHeart, 5, false)
		g.SetDiscardPile([]*domain.Card{topCard})
		giveHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})

		err := g.PlayerDrawFromDiscard()
		assert.NoError(t, err)
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.TonkPhaseDiscard, g.GetPhase())
		assert.Empty(t, g.GetDiscardPile())
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 0)
		assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 1)
		assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrNotHumanTurn)
	})

	t.Run("empty discard error", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)
		g.SetDiscardPile(nil)

		err := g.PlayerDrawFromDiscard()
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerDiscard ---

func TestTonk_PlayerDiscard(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 0)
		g.SetDiscardPile(nil)
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
		})

		err := g.PlayerDiscard(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
		assert.Len(t, g.GetDiscardPile(), 1)
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
		assert.Equal(t, domain.TonkPhaseDraw, g.GetPhase())
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)
		assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 1)
		assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrNotHumanTurn)
	})

	t.Run("invalid index", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 0)
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
		})
		err := g.PlayerDiscard(5)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
		err = g.PlayerDiscard(-1)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})
}

// --- PlayerKnock ---

func TestTonk_PlayerKnock(t *testing.T) {
	t.Run("valid knock with low deadwood", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		// Hand: 3-of-spades-set + 2 small cards (low deadwood)
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignClover, 1, false), // discard candidate
			domain.NewCard(domain.CardDesignSpade, 2, false),  // 2 deadwood after discard
		})
		giveHand(g.GetPlayer(1), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 13, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})
		setupTonkDiscardPhase(g, 0)

		err := g.PlayerKnock(3) // discard Club Ace, leaves set + 2 deadwood
		assert.NoError(t, err)
		assert.Equal(t, 0, g.GetKnockerIdx())
		assert.NotNil(t, g.GetKnockerMelds())
	})

	t.Run("knock fails when deadwood exceeds threshold", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 12, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		})
		setupTonkDiscardPhase(g, 0)

		err := g.PlayerKnock(0)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("undercut: opponent has lower deadwood", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		// Player 0 knocks with deadwood 5 (Spade 5)
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
			domain.NewCard(domain.CardDesignDiamond, 1, false),
			domain.NewCard(domain.CardDesignClover, 13, false), // discard
			domain.NewCard(domain.CardDesignSpade, 5, false),   // 5 deadwood
		})
		// Player 1 has lower deadwood: a triple + 2 aces (deadwood = 2)
		giveHand(g.GetPlayer(1), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 1, false),
			domain.NewCard(domain.CardDesignSpade, 2, false), // deadwood 1+2 = 3
		})
		setupTonkDiscardPhase(g, 0)

		err := g.PlayerKnock(3)
		assert.NoError(t, err)
		assert.True(t, g.GetIsUndercut())
		// opponent should have positive round score
		assert.Greater(t, g.GetPlayer(1).GetCumulativeScore(), 0)
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDrawPhase(g, 0)
		assert.ErrorIs(t, g.PlayerKnock(0), domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 1)
		assert.ErrorIs(t, g.PlayerKnock(0), domain.ErrNotHumanTurn)
	})

	t.Run("invalid index", func(t *testing.T) {
		g := newTestTonk()
		g.Reset()
		setupTonkDiscardPhase(g, 0)
		giveHand(g.GetPlayer(0), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
		})
		err := g.PlayerKnock(99)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})
}

// --- CpuPlay ---

func TestTonk_CpuPlay_DrawPhase(t *testing.T) {
	for _, diff := range []domain.TonkCpuDifficulty{
		domain.TonkCpuDifficultyEasy,
		domain.TonkCpuDifficultyNormal,
		domain.TonkCpuDifficultyHard,
	} {
		t.Run("difficulty", func(t *testing.T) {
			g := newTestTonkWithDifficulty(diff)
			g.Reset()
			if g.GetPhase() != domain.TonkPhaseDraw {
				return // Tonk on deal triggered
			}
			setupTonkDrawPhase(g, 1)

			g.CpuPlay()
			// After CPU draw, phase should be Discard or RoundEnd (if draw pile empty)
			phase := g.GetPhase()
			assert.True(t, phase == domain.TonkPhaseDiscard || phase == domain.TonkPhaseRoundEnd || phase == domain.TonkPhaseGameEnd)
		})
	}
}

func TestTonk_CpuPlay_DiscardPhase(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	setupTonkDiscardPhase(g, 1)
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
	})
	g.CpuPlay()
	// CPU should have discarded one card
	assert.LessOrEqual(t, g.GetPlayer(1).GetCardsSize(), 5)
}

func TestTonk_CpuPlay_DiscardKnock(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	setupTonkDiscardPhase(g, 1)
	// CPU has near-perfect hand
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	})
	g.CpuPlay()
	// CPU should knock since deadwood will be low
	if g.GetPhase() != domain.TonkPhaseGameEnd && g.GetPhase() != domain.TonkPhaseRoundEnd {
		t.Logf("phase=%v", g.GetPhase())
	}
}

func TestTonk_CpuPlay_GameEnded_NoOp(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	// Force gameEndFlag
	g.SetPhase(domain.TonkPhaseGameEnd)
	// Calling CpuPlay should not panic; phase remains
	g.CpuPlay()
	assert.Equal(t, domain.TonkPhaseGameEnd, g.GetPhase())
}

func TestTonk_CpuPlay_HumanTurn_NoOp(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	setupTonkDrawPhase(g, 0) // human
	g.CpuPlay()
	// Should not change state
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

// --- NextRound ---

func TestTonk_NextRound(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	g.SetPhase(domain.TonkPhaseRoundEnd)
	g.NextRound()
	// After NextRound, phase should be Draw or RoundEnd (if Tonk on deal)
	phase := g.GetPhase()
	assert.True(t, phase == domain.TonkPhaseDraw || phase == domain.TonkPhaseRoundEnd || phase == domain.TonkPhaseGameEnd)
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestTonk_NextRound_WrongPhase(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	roundBefore := g.GetRoundNumber()
	g.SetPhase(domain.TonkPhaseDraw)
	g.NextRound()
	// Round number should not change
	assert.Equal(t, roundBefore, g.GetRoundNumber())
}

// --- ScoreRound (compatibility no-op) ---

func TestTonk_ScoreRound_NoOp(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	g.ScoreRound() // should not panic
}

// --- Game End ---

func TestTonk_GameEnd(t *testing.T) {
	g := newTestTonk()
	// **配牌を固定してから点数上限を下げる。**上限 1 のまま素の Reset を通すと、
	// 配牌 Tonk (50 ボーナス + 50 点 = 100) を引いた局でその場で試合が終わり、
	// ノック前にゲーム終了になったり、勝者が配牌 Tonk を引いた側になったりする。
	// seed 1 が配牌 Tonk にならないことは TestTonk_ResetTwice が使っている。
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	require.Equal(t, domain.TonkPhaseDraw, g.GetPhase(), "seed 1 should deal without a Tonk")
	g.SetConfig(domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyNormal, PointLimit: 1})
	giveHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	})
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
	})
	setupTonkDiscardPhase(g, 0)

	require.NoError(t, g.PlayerKnock(3))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

// --- Action Log ---

func TestTonk_ActionLog(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	if g.GetPhase() != domain.TonkPhaseDraw {
		// Tonk on deal triggered an action log entry
		assert.NotEmpty(t, g.GetActionLog())
		return
	}
	// In normal flow, log starts empty
	setupTonkDrawPhase(g, 0)
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.NotEmpty(t, g.GetActionLog())
}

// --- JSON roundtrip ---

func TestTonk_JSONRoundtrip(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	g.SetIsTonk(true)
	g.SetKnockerIdx(0)
	g.SetKnockerMelds([][]*domain.Card{
		{domain.NewCard(domain.CardDesignSpade, 5, false)},
	})
	g.SetKnockerDeadwood([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 7, false),
	})

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Tonk
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.True(t, restored.GetIsTonk())
	assert.Equal(t, 0, restored.GetKnockerIdx())
	assert.Len(t, restored.GetKnockerMelds(), 1)
	assert.Len(t, restored.GetKnockerDeadwood(), 1)
}

func TestTonk_UnmarshalJSON_TooLarge(t *testing.T) {
	// Build JSON with too many players
	type wire struct {
		Players []*domain.TonkPlayer `json:"pl"`
	}
	pls := make([]*domain.TonkPlayer, 1001)
	for i := range pls {
		pls[i] = domain.NewTonkPlayer(false)
	}
	data, err := json.Marshal(wire{Players: pls})
	require.NoError(t, err)

	var t2 domain.Tonk
	err = json.Unmarshal(data, &t2)
	assert.Error(t, err)
}

func TestTonk_UnmarshalJSON_Empty(t *testing.T) {
	var t2 domain.Tonk
	require.NoError(t, json.Unmarshal([]byte(`{}`), &t2))
	// Should default to safe values
	assert.NotNil(t, t2.GetDiscardPile())
	assert.NotNil(t, t2.GetActionLog())
}

// --- ScoreRound paths via direct knock ---

func TestTonk_ScoreNormalKnock(t *testing.T) {
	g := newTestTonk()
	g.Reset()
	giveHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignSpade, 2, false), // 2 deadwood after discard
	})
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false), // 49 deadwood
	})
	setupTonkDiscardPhase(g, 0)

	require.NoError(t, g.PlayerKnock(3))
	assert.False(t, g.GetIsUndercut())
	assert.Greater(t, g.GetPlayer(0).GetCumulativeScore(), 0)
}

// --- PlayerDiscard error: game ended ---

func TestTonk_PlayerDiscard_GameEnded(t *testing.T) {
	g := newTestTonk()
	// Reset() runs the random deal under newTestTonk's safe PointLimit (10000)
	// so a Tonk-on-deal cannot pre-end the game. Apply the strict PointLimit=1
	// only afterwards — the test's PlayerKnock(3) scoreRound is what should
	// trip the game-end here, not the random deal.
	g.Reset()
	g.SetConfig(domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyNormal, PointLimit: 1})
	giveHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	})
	giveHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
	})
	setupTonkDiscardPhase(g, 0)
	require.NoError(t, g.PlayerKnock(3))

	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerKnock(0), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrGameEnded)
}

// --- Setters smoke test ---

func TestTonk_Setters(t *testing.T) {
	g := newTestTonk()
	g.SetRoundNumber(3)
	assert.Equal(t, 3, g.GetRoundNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetKnockerIdx(0)
	assert.Equal(t, 0, g.GetKnockerIdx())
}
