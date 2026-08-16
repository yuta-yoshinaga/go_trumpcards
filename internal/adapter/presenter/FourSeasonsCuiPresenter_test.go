//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupFourSeasonsCuiMockDefaults(g *interfaces.MockFourSeasonsGame) {
	g.On("GetPhase").Return(domain.FourSeasonsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetBaseRank").Return(7).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)}).Maybe()

	var fnd [domain.FourSeasonsFoundationCnt][]*domain.Card
	fnd[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	g.On("GetFoundations").Return(fnd).Maybe()

	var tab [domain.FourSeasonsTableauCnt][]*domain.Card
	tab[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 12, false)}
	g.On("GetTableau").Return(tab).Maybe()
}

func TestFourSeasonsCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsCuiMockDefaults(g)
		out := new(FourSeasonsCuiPresenter).Output(g, nil)

		assert.Contains(t, out, "Four Seasons")
		assert.Contains(t, out, "[F0]")
		assert.Contains(t, out, "[T0]")
		assert.Contains(t, out, "ストック")
	})

	// The base rank is what every placement rule reads, so it must be on screen
	// before the board — the player cannot judge a corner without it.
	t.Run("prints the base rank", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsCuiMockDefaults(g)
		assert.Contains(t, new(FourSeasonsCuiPresenter).Output(g, nil), "ベースランク")
	})

	// An empty corner still announces the rank it wants: the base rank, not an Ace.
	t.Run("empty corner announces the base rank as next", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsCuiMockDefaults(g)
		out := new(FourSeasonsCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "(空)")
		assert.Contains(t, out, "次:")
	})

	t.Run("complete corner reads as complete", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		var fnd [domain.FourSeasonsFoundationCnt][]*domain.Card
		for v := 1; v <= domain.CardValueMax; v++ {
			fnd[0] = append(fnd[0], domain.NewCard(domain.CardDesignSpade, v, false))
		}
		g.On("GetFoundations").Return(fnd)
		setupFourSeasonsCuiMockDefaults(g)
		assert.Contains(t, new(FourSeasonsCuiPresenter).Output(g, nil), "完成")
	})

	t.Run("empty waste and tableau", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetWaste").Return([]*domain.Card{})
		g.On("GetTableau").Return([domain.FourSeasonsTableauCnt][]*domain.Card{})
		setupFourSeasonsCuiMockDefaults(g)
		assert.NotEmpty(t, new(FourSeasonsCuiPresenter).Output(g, nil))
	})

	for _, tt := range []struct {
		name  string
		phase domain.FourSeasonsPhase
	}{
		{"game clear", domain.FourSeasonsPhaseGameClear},
		{"game over", domain.FourSeasonsPhaseGameOver},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := new(interfaces.MockFourSeasonsGame)
			g.On("GetPhase").Return(tt.phase)
			setupFourSeasonsCuiMockDefaults(g)
			assert.NotEmpty(t, new(FourSeasonsCuiPresenter).Output(g, nil))
		})
	}

	t.Run("error is surfaced", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsCuiMockDefaults(g)
		assert.Contains(t, new(FourSeasonsCuiPresenter).Output(g, assert.AnError), assert.AnError.Error())
	})
}

// The CUI's own next-rank arithmetic must wrap the same way the domain's does;
// if it drifts, only the "next:" hint lies while the board stays correct.
func TestFourSeasonsCuiNextRank_Wraps(t *testing.T) {
	assert.Equal(t, 1, fourSeasonsCuiNextRank(domain.CardValueMax), "K -> A")
	assert.Equal(t, 8, fourSeasonsCuiNextRank(7))
}

func TestFourSeasonsCuiPresenter_HintOutput(t *testing.T) {
	t.Run("from a cross pile", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(&domain.FourSeasonsHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1})
		out := new(FourSeasonsCuiPresenter).HintOutput(g)
		assert.Contains(t, out, "2")
		assert.Contains(t, out, "1")
	})
	t.Run("from the waste", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(&domain.FourSeasonsHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0})
		assert.NotEmpty(t, new(FourSeasonsCuiPresenter).HintOutput(g))
	})
	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(nil)
		assert.NotEmpty(t, new(FourSeasonsCuiPresenter).HintOutput(g))
	})
}

func TestFourSeasonsCuiPresenter_ActionLogOutput(t *testing.T) {
	// While the game runs the transcript is withheld — it is a post-mortem, not
	// a live crib sheet.
	t.Run("withheld while playing", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetPhase").Return(domain.FourSeasonsPhasePlaying)
		assert.Equal(t, actionLogToText(nil), new(FourSeasonsCuiPresenter).ActionLogOutput(g))
	})
	t.Run("emitted once ended", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetPhase").Return(domain.FourSeasonsPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "引きました"}})
		assert.Contains(t, new(FourSeasonsCuiPresenter).ActionLogOutput(g), "draw")
	})
}
