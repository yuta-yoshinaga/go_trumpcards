//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupColoradoWebMockDefaults(g *interfaces.MockColoradoGame) {
	g.On("GetPhase").Return(domain.ColoradoPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.ColoradoTableauCnt][]*domain.Card
	for i := range domain.ColoradoTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.ColoradoFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()

	// The first half build up from the Ace, the second half down from the King.
	for i := range domain.ColoradoFoundationCnt {
		g.On("IsAscendingFoundation", i).Return(i < domain.ColoradoAscendingCnt).Maybe()
	}
}

func parseColoradoOutput(t *testing.T, jsonStr string) *controller.ColoradoWebOutput {
	t.Helper()
	var out controller.ColoradoWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupColoradoOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupColoradoOutputMock(g *interfaces.MockColoradoGame) {
	setupColoradoWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestColoradoWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoOutputMock(g)

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 96, result.StockCount)
		assert.Len(t, result.Tableau, domain.ColoradoTableauCnt)
		assert.Len(t, result.Foundation, domain.ColoradoFoundationCnt)
		// The build direction has to be on the wire: the page cannot label the
		// piles from the index without silently duplicating the rule.
		assert.Equal(t,
			[]bool{true, true, true, true, false, false, false, false},
			result.FoundationAscending)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "colorado.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).Output(g, nil))
		assert.Equal(t, "colorado.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoOutputMock(g)

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.ColoradoPhase
		code string
	}{
		{"game clear", domain.ColoradoPhaseGameClear, "colorado.gameClear"},
		{"game over", domain.ColoradoPhaseGameOver, "colorado.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockColoradoGame)
			setupColoradoOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseColoradoOutput(t, new(ColoradoWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestColoradoWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.ColoradoHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestColoradoWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoWebMockDefaults(g)
		g.On("GetHint").Return(&domain.ColoradoHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "colorado.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		setupColoradoWebMockDefaults(g)
		g.On("GetHint").Return((*domain.ColoradoHint)(nil))

		result := parseColoradoOutput(t, new(ColoradoWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "colorado.noHint", result.MessageCode)
	})
}

func TestColoradoWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		g.On("GetPhase").Return(domain.ColoradoPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(ColoradoWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockColoradoGame)
		g.On("GetPhase").Return(domain.ColoradoPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(ColoradoWebPresenter).ActionLogOutput(g), "move")
	})
}
