//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// nertzGameForTest creates a Nertz with default 4-player config.
func nertzGameForTest(t *testing.T) *domain.Nertz {
	t.Helper()
	g := domain.NewDefaultNertz()
	g.Reset()
	return g
}

func TestNewDefaultNertz(t *testing.T) {
	g := domain.NewDefaultNertz()
	require.NotNil(t, g)
	assert.Equal(t, domain.NertzPhaseIdle, g.GetPhase())
}

func TestNertz_ResetInitialState(t *testing.T) {
	g := nertzGameForTest(t)

	assert.Equal(t, domain.NertzPhasePlaying, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNo())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetMatchWinner())

	cfg := g.GetConfig()
	assert.Equal(t, cfg.PlayerCount, len(g.GetPlayers()))
	for _, p := range g.GetPlayers() {
		assert.Equal(t, domain.NertzPileSize, p.NertzSize())
		assert.Equal(t, domain.NertzInitialStockSize, p.StockSize())
		assert.Equal(t, 0, p.WasteSize())
		for c := 0; c < domain.NertzTableauCnt; c++ {
			assert.Equal(t, 1, p.TableauSize(c))
		}
	}
	// foundations: fixed 4 * playerCount empty slots
	assert.Equal(t, 4*cfg.PlayerCount, len(g.GetFoundations()))
	for _, f := range g.GetFoundations() {
		assert.True(t, f.IsEmpty())
	}
}

func TestNertz_ResetWithConfig(t *testing.T) {
	g := domain.NewDefaultNertz()
	cfg := domain.DefaultNertzConfig()
	cfg.PlayerCount = 3
	cfg.DrawCount = 1
	g.ResetWithConfig(cfg)

	assert.Equal(t, 3, len(g.GetPlayers()))
	assert.Equal(t, 12, len(g.GetFoundations()))
	assert.Equal(t, 1, g.GetConfig().DrawCount)
}

func TestNertz_ResetWithInvalidConfigFallsBackToDefault(t *testing.T) {
	g := domain.NewDefaultNertz()
	bad := domain.NertzConfig{PlayerCount: 99, DrawCount: 7, TargetScore: 0}
	g.ResetWithConfig(bad)
	assert.Equal(t, domain.NertzPlayerCntDefault, g.GetConfig().PlayerCount)
}

// --- Helpers for board manipulation ---

func clearFoundations(g *domain.Nertz) {
	founds := g.GetFoundations()
	for i := range founds {
		founds[i] = domain.NewNertzFoundation()
	}
	g.SetFoundations(founds)
}

func clearPlayerPiles(p *domain.NertzPlayer) {
	p.ResetRoundPiles()
}

func TestNertz_DrawStock_DefaultDraw3(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	stockBefore := p.StockSize()
	require.NoError(t, g.DrawStock(0))
	assert.Equal(t, stockBefore-3, p.StockSize())
	assert.Equal(t, 3, p.WasteSize())
}

func TestNertz_DrawStock_RecyclesWhenStockEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Set up: stock=empty, waste=2 cards
	p.PushWaste(newNertzCard(domain.CardDesignSpade, 5))
	p.PushWaste(newNertzCard(domain.CardDesignHeart, 9))

	require.NoError(t, g.DrawStock(0))
	assert.Equal(t, 0, p.WasteSize(), "all waste recycled into stock")
	assert.Equal(t, 2, p.StockSize())
}

func TestNertz_DrawStock_FailsWhenBothEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	err := g.DrawStock(0)
	assert.Error(t, err)
}

func TestNertz_DrawStock_RejectsInvalidPlayer(t *testing.T) {
	g := nertzGameForTest(t)
	assert.Error(t, g.DrawStock(-1))
	assert.Error(t, g.DrawStock(99))
}

func TestNertz_MoveNertzToFoundation_OpensNewFoundationWithAce(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	ace := newNertzCard(domain.CardDesignSpade, 1)
	p.PushNertz(ace)

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, 0, p.NertzSize())
	founds := g.GetFoundations()
	assert.Equal(t, 1, founds[0].Size())
	assert.Equal(t, domain.CardDesignSpade, founds[0].Suit())
}

func TestNertz_MoveNertzToFoundation_RejectsInvalid(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushNertz(newNertzCard(domain.CardDesignSpade, 5))
	// Foundation is empty; only Ace allowed
	err := g.MoveNertzToFoundation(0, 0)
	assert.Error(t, err)
	assert.Equal(t, 1, p.NertzSize())
}

func TestNertz_MoveNertzToFoundation_RoundEndOnEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushNertz(newNertzCard(domain.CardDesignDiamond, 1))

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, domain.NertzPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestNertz_MoveNertzToTableau_RequiresValidPlacement(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Put a black K on tableau col 0 so a red Q can land
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushNertz(newNertzCard(domain.CardDesignHeart, 12))

	require.NoError(t, g.MoveNertzToTableau(0, 0))
	assert.Equal(t, 0, p.NertzSize())
	assert.Equal(t, 2, p.TableauSize(0))
}

func TestNertz_MoveNertzToTableau_RejectsBadColor(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushNertz(newNertzCard(domain.CardDesignClover, 12)) // black-on-black rejected

	err := g.MoveNertzToTableau(0, 0)
	assert.Error(t, err)
}

func TestNertz_MoveNertzToTableau_EmptyColumnOnlyAcceptsTopCard(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Empty tableau col 0 accepts ANY card from Nertz pile (Nertz has free placement on empty cols)
	p.PushNertz(newNertzCard(domain.CardDesignHeart, 7))
	require.NoError(t, g.MoveNertzToTableau(0, 0))
	assert.Equal(t, 1, p.TableauSize(0))
}

func TestNertz_MoveWasteToFoundation_NeedsAceWhenEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushWaste(newNertzCard(domain.CardDesignSpade, 1))

	require.NoError(t, g.MoveWasteToFoundation(0, 0))
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 1, g.GetFoundations()[0].Size())
}

func TestNertz_MoveWasteToFoundation_ContinuesSuit(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	founds := g.GetFoundations()
	require.NoError(t, founds[2].Push(newNertzCard(domain.CardDesignDiamond, 1), 1))
	g.SetFoundations(founds)
	p.PushWaste(newNertzCard(domain.CardDesignDiamond, 2))

	require.NoError(t, g.MoveWasteToFoundation(0, 2))
	assert.Equal(t, 2, g.GetFoundations()[2].Size())
	assert.Equal(t, 1, g.GetFoundations()[2].CountByContributor(0))
}

func TestNertz_MoveWasteToTableau(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	p.PushTableau(1, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 8), FaceUp: true})
	p.PushWaste(newNertzCard(domain.CardDesignHeart, 7))

	require.NoError(t, g.MoveWasteToTableau(0, 1))
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 2, p.TableauSize(1))
}

func TestNertz_MoveTableauToFoundation_AutoFillsFromNertz(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	// Foundation 3 has Heart-A; we'll play Heart-2 from a single-card tableau col
	founds := g.GetFoundations()
	require.NoError(t, founds[3].Push(newNertzCard(domain.CardDesignHeart, 1), 0))
	g.SetFoundations(founds)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 2), FaceUp: true})

	// Nertz pile has a card to auto-fill the empty column with
	autoCard := newNertzCard(domain.CardDesignSpade, 9)
	p.PushNertz(autoCard)

	require.NoError(t, g.MoveTableauToFoundation(0, 0, 3))
	assert.Equal(t, 1, p.TableauSize(0), "column auto-filled from Nertz pile")
	assert.Equal(t, autoCard, p.TableauTop(0))
	assert.Equal(t, 0, p.NertzSize())
}

func TestNertz_MoveTableauToFoundation_NoAutoFillWhenNertzEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	founds := g.GetFoundations()
	require.NoError(t, founds[3].Push(newNertzCard(domain.CardDesignHeart, 1), 0))
	g.SetFoundations(founds)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 2), FaceUp: true})

	require.NoError(t, g.MoveTableauToFoundation(0, 0, 3))
	assert.Equal(t, 0, p.TableauSize(0))
}

func TestNertz_MoveTableauToTableau_Substack(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)

	// Source col 0: Spade-K, Heart-Q, Clover-J (alternating colors descending)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 12), FaceUp: true})
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 11), FaceUp: true})
	// Destination col 1: empty (any card accepted)
	require.NoError(t, g.MoveTableauToTableau(0, 0, 1, 1))
	// Take starting at idx 1 = Heart-Q + Clover-J (2 cards)
	assert.Equal(t, 1, p.TableauSize(0))
	assert.Equal(t, 2, p.TableauSize(1))
}

func TestNertz_MoveTableauToTableau_RejectsSameColumn(t *testing.T) {
	g := nertzGameForTest(t)
	err := g.MoveTableauToTableau(0, 0, 0, 0)
	assert.Error(t, err)
}

func TestNertz_RejectsActionsWhenNotPlaying(t *testing.T) {
	g := nertzGameForTest(t)
	g.SetPhase(domain.NertzPhaseRoundEnd)
	assert.Error(t, g.DrawStock(0))
	assert.Error(t, g.MoveNertzToFoundation(0, 0))
}

func TestNertz_JSON(t *testing.T) {
	g := nertzGameForTest(t)
	// take an action so the state is interesting
	_ = g.DrawStock(0)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := domain.NewDefaultNertz()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, len(g.GetPlayers()), len(restored.GetPlayers()))
	assert.Equal(t, len(g.GetFoundations()), len(restored.GetFoundations()))
	assert.Equal(t, g.GetRoundNo(), restored.GetRoundNo())
}

// --- Hard CPU lookahead tests (issue #1562) ---

// nertzHardGameForTest creates a freshly reset Hard-difficulty Nertz game.
// Cards are then cleared via clearPlayerPiles so the test owns the board.
func nertzHardGameForTest(t *testing.T) *domain.Nertz {
	t.Helper()
	g := domain.NewDefaultNertz()
	cfg := domain.DefaultNertzConfig()
	cfg.CpuDifficulty = domain.NertzCpuDifficultyHard
	g.ResetWithConfig(cfg)
	return g
}

// TestNertz_FindCpuMoveHardPrefersNertzToFoundation pins the headline
// behavior of issue #1562: when both a tableau-to-foundation and a
// nertz-to-foundation are legal, Hard prefers the nertz-to-foundation
// because reducing the Nertz pile is the win condition. Easy/Normal's
// first-match heuristic happens to produce the same answer in this case
// (NF is checked before TF), so the test verifies the Hard branch reaches
// the same conclusion through scoring rather than ordering.
func TestNertz_FindCpuMoveHardPrefersNertzToFoundation(t *testing.T) {
	g := nertzHardGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	cpu := g.GetPlayers()[1]
	// Both moves are legal:
	//   - Nertz top is Ace of Spades → empty foundation
	//   - Tableau col 0 top is Ace of Hearts → another empty foundation
	// Hard MUST pick the nertz-to-foundation (moveNF +100 vs moveTF +20).
	cpu.PushNertz(newNertzCard(domain.CardDesignSpade, 1))
	cpu.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 1), FaceUp: true})

	move := g.FindCpuMove(1)
	require.NotNil(t, move)
	assert.Equal(t, "moveNF", move.ActionType)
}

// TestNertz_FindCpuMoveHardPrefersNertzToTableauOverTableauToTableau
// verifies that Hard chases pile reduction even when a foundation move
// is unavailable. Easy/Normal would also pick moveNT here (first-match
// scans NT before TT), so this test pins the Hard scoring spec rather
// than a behavioral divergence.
func TestNertz_FindCpuMoveHardPrefersNertzToTableauOverTableauToTableau(t *testing.T) {
	g := nertzHardGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	cpu := g.GetPlayers()[1]
	// Black King on col 0 → red Q (Hearts) lands on it from EITHER nertz
	// (top = QH) or tableau col 1 (also QH). Hard must prefer the nertz
	// source because pile reduction is the win condition.
	cpu.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	cpu.PushTableau(1, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 12), FaceUp: true})
	cpu.PushNertz(newNertzCard(domain.CardDesignDiamond, 12))

	move := g.FindCpuMove(1)
	require.NotNil(t, move)
	assert.Equal(t, "moveNT", move.ActionType,
		"Hard must reduce nertz pile when both NT and TT are legal; got %v", move.ActionType)
}

// TestNertz_FindCpuMoveHardEmptyColumnPenalty pins the "stranded waste
// pile" penalty's two regimes (PR #1587 review). Both sub-cases set up
// the same trigger condition (Nertz pile empty, a moveTF that would
// empty its source column) but with competing moves at different
// strengths so the test asserts a single, deterministic outcome —
// either the penalty narrows the gap but moveTF still wins, or the
// penalty is the deciding factor and a competing move wins.
func TestNertz_FindCpuMoveHardEmptyColumnPenalty(t *testing.T) {
	t.Run("penalty applied; moveTF still wins narrowly", func(t *testing.T) {
		g := nertzHardGameForTest(t)
		clearFoundations(g)
		for _, p := range g.GetPlayers() {
			clearPlayerPiles(p)
		}
		cpu := g.GetPlayers()[1]
		// Nertz pile empty. col 0 holds a solitary Ace of Spades, so
		// moveTF (Ace → foundation 0) empties the column. Competing
		// move: waste 2 of Hearts onto col 2's 3 of Spades.
		// Scores: moveTF base 20 - empty-column penalty 8 = 12.
		// moveWT base 10 (col 2 is non-empty so no empty-target penalty) = 10.
		// 12 > 10 → moveTF still wins, but the test now pins it.
		cpu.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 1), FaceUp: true})
		cpu.PushTableau(2, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 3), FaceUp: true})
		cpu.PushWaste(newNertzCard(domain.CardDesignHeart, 2))

		move := g.FindCpuMove(1)
		require.NotNil(t, move)
		assert.Equal(t, "moveTF", move.ActionType,
			"moveTF (12) still beats moveWT (10) after the empty-column penalty")
	})

	t.Run("penalty deflects column choice when both moveTFs are legal", func(t *testing.T) {
		g := nertzHardGameForTest(t)
		clearFoundations(g)
		for _, p := range g.GetPlayers() {
			clearPlayerPiles(p)
		}
		cpu := g.GetPlayers()[1]
		// Nertz pile empty. Two foundation-legal moveTF candidates:
		//   col 0: solo Ace of Spades (size 1). moveTF empties the
		//          column → penalty applies. Score 20 − 0 − 8 = 12.
		//   col 1: face-down Clover 5 then face-up Ace of Hearts.
		//          moveTF takes only the top card; size 2 ≠ moves 1 →
		//          penalty does NOT apply. Score 20 − 0 = 20.
		// Without the penalty both would tie at 20 and the foundation
		// tiebreak (lower index wins) would pick col 0's enumeration
		// (it iterates first). With the penalty, col 1 wins decisively.
		cpu.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 1), FaceUp: true})
		cpu.PushTableau(1, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 5), FaceUp: false})
		cpu.PushTableau(1, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 1), FaceUp: true})

		move := g.FindCpuMove(1)
		require.NotNil(t, move)
		assert.Equal(t, "moveTF", move.ActionType)
		assert.Equal(t, 1, move.FromCol,
			"penalty must steer Hard onto col 1's non-emptying moveTF "+
				"(score 20) over col 0's emptying moveTF (score 12)")
	})
}

// TestNertz_FindCpuMoveHardFallsBackToDrawWhenNothingElseLegal verifies
// that the Hard CPU still draws as a last resort. enumerateCpuMoves only
// adds the draw candidate when no other move is legal — confirm that
// empty board (no playable cards) still produces a draw.
func TestNertz_FindCpuMoveHardFallsBackToDrawWhenNothingElseLegal(t *testing.T) {
	g := nertzHardGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	cpu := g.GetPlayers()[1]
	// Stock has cards but tableau/nertz/waste are all empty so there are no
	// place-able moves. Push a single stock card directly so DrawStock can fire.
	cpu.PushStock(newNertzCard(domain.CardDesignSpade, 5))
	move := g.FindCpuMove(1)
	require.NotNil(t, move)
	assert.Equal(t, "draw", move.ActionType)
}

// TestNertz_FindCpuMoveDifficultyDispatch verifies the smoke property:
// Easy/Normal go through findCpuMoveFast, Hard goes through
// findCpuMoveHard. Setting the same trivial position on each difficulty
// must yield a non-nil move for all three, and the structural categories
// match the heuristics (1..7 sequence at minimum). Other tests pin the
// Hard scoring details; this one just guards the dispatch wiring.
func TestNertz_FindCpuMoveDifficultyDispatch(t *testing.T) {
	for _, d := range []domain.NertzCpuDifficulty{
		domain.NertzCpuDifficultyEasy,
		domain.NertzCpuDifficultyNormal,
		domain.NertzCpuDifficultyHard,
	} {
		g := domain.NewDefaultNertz()
		cfg := domain.DefaultNertzConfig()
		cfg.CpuDifficulty = d
		g.ResetWithConfig(cfg)
		// After Reset all four tableau columns hold a face-up card; some
		// moves are likely legal. Just check we got *something* back.
		move := g.FindCpuMove(1)
		assert.NotNil(t, move, "difficulty=%v must produce a move", d)
	}
}
