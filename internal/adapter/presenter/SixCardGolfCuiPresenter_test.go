//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// sixCardGolfDrawPending returns a human-turn DrawPending game whose grid is
// filled with the given face-up value and with drawnCard set.
func sixCardGolfDrawPending(gridValue, drawnValue int) *domain.SixCardGolf {
	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.SixCardGolfPhaseDrawPending)
	p := g.GetPlayer(0)
	for i := range p.Grid {
		p.Grid[i] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, gridValue, false), FaceUp: true}
	}
	g.SetDrawnCard(domain.NewCard(domain.CardDesignSpade, drawnValue, false))
	return g
}

func TestSixCardGolfCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SixCardGolfCuiPresenter)

	t.Run("swap recommended for a low drawn card", func(t *testing.T) {
		g := sixCardGolfDrawPending(11, 1) // Jacks in grid, drew an Ace
		assert.Contains(t, p.HintOutput(g), "sw")
	})

	t.Run("discard recommended for a high drawn card", func(t *testing.T) {
		g := sixCardGolfDrawPending(1, 11) // Aces in grid, drew a Jack
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintDiscard"))
	})

	t.Run("column pair is called out", func(t *testing.T) {
		g := sixCardGolfDrawPending(9, 5)
		g.GetPlayer(0).Grid[3] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}
		assert.Contains(t, p.HintOutput(g), i18n.Tf("sixcardgolf.hintSwapPair", "pos", "0"))
	})

	t.Run("draw from discard for a low top card", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhasePlayerTurn)
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)}) // King = 0
		assert.Contains(t, p.HintOutput(g), "dd")
	})

	t.Run("draw from stock for a high top card", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhasePlayerTurn)
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)}) // 9 > 3
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintDrawStock"))
	})

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		g := sixCardGolfDrawPending(11, 1)
		g.SetCurrentPlayerIdx(1) // CPU
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintNone"))
	})

	t.Run("no hint outside a decision phase", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhaseSetup)
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintNone"))
	})
}

// 席名が i18n のキーのまま画面に出ていた (#7061 でマニュアルを直そうとして発見)。
// `cuiPlayerNameHuman` / `cuiPlayerNameCPU` はどのロケールにも存在せず、
// i18n.T は未知のキーをそのまま返すので**キー名が利用者に見えていた**。
func TestSixCardGolfCuiPresenterResolvesSeatNames(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	out := new(presenter.SixCardGolfCuiPresenter).Output(g, nil)

	// **キーが漏れていないこと**が本体。負のコントロールとして、
	// 解決後の文言も見る (未翻訳ならキーが返るので両方見ないと意味がない)。
	assert.NotContains(t, out, "cuiPlayerName", "i18n キーが解決されずに出ている")
	assert.Contains(t, out, "あなた", "人間の席名が解決されていない")
	assert.Contains(t, out, "CPU 1", "CPU の席名が解決されていない")
}

func TestSixCardGolfCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SixCardGolfCuiPresenter)

	t.Run("displays column scores with pair, uncertain and determined columns during play", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhasePlayerTurn)

		human := g.GetPlayer(0)
		// Col 0: Pair of 7s (both face up)
		human.Grid[0] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 7, false), FaceUp: true}
		human.Grid[3] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignHeart, 7, false), FaceUp: true}
		// Col 1: top face up 3, bot face down (uncertain)
		human.Grid[1] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignDiamond, 3, false), FaceUp: true}
		human.Grid[4] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignClover, 8, false), FaceUp: false}
		// Col 2: top face up 4, bot face up 5 (determined, sum=9)
		human.Grid[2] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 4, false), FaceUp: true}
		human.Grid[5] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignHeart, 5, false), FaceUp: true}

		out := p.Output(g, nil)

		// Column scores row assertions
		pairStr := i18n.T("sixcardgolf.columnScorePair")
		uncertainStr := i18n.Tf("sixcardgolf.columnScoreUncertain", "score", "3")
		determinedStr := i18n.Tf("sixcardgolf.columnScore", "score", "9")

		assert.Contains(t, out, pairStr)
		assert.Contains(t, out, uncertainStr)
		assert.Contains(t, out, determinedStr)

		// Acceptance condition 1: Pair column is distinct and visible
		assert.Contains(t, out, pairStr+" "+uncertainStr+" "+determinedStr)

		// Acceptance condition 2: Not round over, so no final score line yet
		assert.NotContains(t, out, i18n.T("sixcardgolf.scoreLine"))
	})

	t.Run("displays determined column scores and final score after round over", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetPhase(domain.SixCardGolfPhaseRoundOver)

		human := g.GetPlayer(0)
		for i := range human.Grid {
			human.Grid[i].FaceUp = true
		}
		// Col 0: Pair of 7s
		human.Grid[0] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 7, false), FaceUp: true}
		human.Grid[3] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignHeart, 7, false), FaceUp: true}
		// Col 1: 3 + 8 = 11
		human.Grid[1] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignDiamond, 3, false), FaceUp: true}
		human.Grid[4] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignClover, 8, false), FaceUp: true}
		// Col 2: 4 + 5 = 9
		human.Grid[2] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 4, false), FaceUp: true}
		human.Grid[5] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignHeart, 5, false), FaceUp: true}

		out := p.Output(g, nil)

		pairStr := i18n.T("sixcardgolf.columnScorePair")
		col1Str := i18n.Tf("sixcardgolf.columnScore", "score", "11")
		col2Str := i18n.Tf("sixcardgolf.columnScore", "score", "9")

		// All columns are determined (no +?)
		assert.Contains(t, out, pairStr+" "+col1Str+" "+col2Str)
		// Total score line (0 + 11 + 9 = 20)
		assert.Contains(t, out, i18n.Tf("sixcardgolf.scoreLine", "score", "20"))
	})

	t.Run("action log output", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
