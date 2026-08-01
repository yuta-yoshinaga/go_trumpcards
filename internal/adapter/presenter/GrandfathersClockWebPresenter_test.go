//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGrandfathersClockWebMockDefaults(g *interfaces.MockGrandfathersClockGame) {
	g.On("GetPhase").Return(domain.GrandfathersClockPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(false).Maybe()

	var tableau [domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard
	for i := range domain.GrandfathersClockTableauCnt {
		tableau[i] = make([]*domain.GrandfathersClockTableauCard, domain.GrandfathersClockColumnLen)
		for j := range domain.GrandfathersClockColumnLen {
			tableau[i][j] = &domain.GrandfathersClockTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.GrandfathersClockFoundationCnt][]*domain.Card
	for i := range domain.GrandfathersClockFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseGrandfathersClockOutput(t *testing.T, jsonStr string) *controller.GrandfathersClockWebOutput {
	t.Helper()
	var out controller.GrandfathersClockWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupGrandfathersClockOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupGrandfathersClockOutputMock(g *interfaces.MockGrandfathersClockGame) {
	setupGrandfathersClockWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestGrandfathersClockWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Len(t, result.Tableau, domain.GrandfathersClockTableauCnt)
		assert.Len(t, result.Foundation, domain.GrandfathersClockFoundationCnt)
		assert.Equal(t, "grandfathersclock.playing", result.MessageCode)
	})

	// The target rank is on the wire so the client never has to recompute the
	// clock ordering and drift from the domain.
	t.Run("each face carries its target rank", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
		for i, f := range result.Foundation {
			assert.Equal(t, domain.GrandfathersClockTargetRank(i), f.TargetRank, "face %d", i)
			assert.False(t, f.Complete)
		}
		assert.Equal(t, 1, result.Foundation[0].TargetRank, "1 o'clock wants an Ace")
		assert.Equal(t, 12, result.Foundation[11].TargetRank, "12 o'clock wants a Queen")
	})

	t.Run("completed faces are flagged", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(true)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
		for _, f := range result.Foundation {
			assert.True(t, f.Complete)
		}
	})

	t.Run("all face up", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.GrandfathersClockPhase
		code string
	}{
		{"game clear", domain.GrandfathersClockPhaseGameClear, "grandfathersclock.gameClear"},
		{"game over", domain.GrandfathersClockPhaseGameOver, "grandfathersclock.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockGrandfathersClockGame)
			setupGrandfathersClockOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
		assert.Equal(t, "grandfathersclock.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGrandfathersClockWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.GrandfathersClockHint{FromCol: 2, ToZone: "foundation", ToIdx: 0}

	g := new(interfaces.MockGrandfathersClockGame)
	setupGrandfathersClockWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).Output(g, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestGrandfathersClockWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockWebMockDefaults(g)
		g.On("GetHint").Return(&domain.GrandfathersClockHint{FromCol: 3, ToZone: "foundation", ToIdx: 7})

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 3, result.Hint.FromCol)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, 7, result.Hint.ToIdx)
		assert.Equal(t, "grandfathersclock.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockWebMockDefaults(g)
		g.On("GetHint").Return((*domain.GrandfathersClockHint)(nil))

		result := parseGrandfathersClockOutput(t, new(GrandfathersClockWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "grandfathersclock.noHint", result.MessageCode)
	})
}

func TestGrandfathersClockWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		g.On("GetPhase").Return(domain.GrandfathersClockPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(GrandfathersClockWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		g.On("GetPhase").Return(domain.GrandfathersClockPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(GrandfathersClockWebPresenter).ActionLogOutput(g), "move")
	})
}
