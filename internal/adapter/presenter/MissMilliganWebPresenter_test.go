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

func setupMissMilliganWebMockDefaults(g *interfaces.MockMissMilliganGame) {
	g.On("GetPhase").Return(domain.MissMilliganPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetWaived").Return([]*domain.Card(nil)).Maybe()
	g.On("CanWaive").Return(false).Maybe()

	var tableau [domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard
	for i := range domain.MissMilliganTableauCnt {
		tableau[i] = []*domain.MissMilliganTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.MissMilliganFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseMissMilliganOutput(t *testing.T, jsonStr string) *controller.MissMilliganWebOutput {
	t.Helper()
	var out controller.MissMilliganWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupMissMilliganOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupMissMilliganOutputMock(g *interfaces.MockMissMilliganGame) {
	setupMissMilliganWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestMissMilliganWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 96, result.StockCount)
		assert.Len(t, result.Tableau, domain.MissMilliganTableauCnt)
		assert.Len(t, result.Foundation, domain.MissMilliganFoundationCnt)
		assert.Empty(t, result.Waived)
		assert.False(t, result.CanWaive)
		assert.Equal(t, "missmilligan.playing", result.MessageCode)
	})

	// canWaive is the domain's judgement ("stock gone and nothing held"); the
	// client must not have to recompute it.
	t.Run("canWaive is surfaced", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanWaive")
		g.On("CanWaive").Return(true)

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
		assert.True(t, result.CanWaive)
	})

	// Holding cards is a distinct state: it blocks dealing and waiving, so it
	// gets its own message rather than looking like ordinary play.
	t.Run("waiving has its own message", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaived")
		g.On("GetWaived").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, true)})

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
		assert.Len(t, result.Waived, 1)
		assert.Equal(t, "missmilligan.waiving", result.MessageCode)
	})

	t.Run("stalemate outranks waiving", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaived")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("GetWaived").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, true)})
		g.On("IsStalemate").Return(true)

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
		assert.Equal(t, "missmilligan.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	// 不正な操作は文言ではなくコードで返す。以前は英語の完成文を Message に
	// 入れており、日本語ロケールでもそれがそのまま出ていた (#5556)。
	t.Run("a rejected move is reported as a code, not a phrase", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)

		err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "missmilligan.errNotDescendingRun", nil)
		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, err))

		assert.Equal(t, "missmilligan.errNotDescendingRun", result.MessageCode)
		// 完成文が残っていると、クライアントはコードより先にそれを出しうる。
		assert.Empty(t, result.Message)
	})

	t.Run("a rejected move carries its params", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganOutputMock(g)

		err := domain.NewDomainErrorCode(domain.ErrInvalidCard, "missmilligan.errBadColumn",
			map[string]string{"col": "9"})
		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, err))
		assert.Equal(t, "9", result.MessageParams["col"])
	})

	for _, tc := range []struct {
		name string
		val  domain.MissMilliganPhase
		code string
	}{
		{"game clear", domain.MissMilliganPhaseGameClear, "missmilligan.gameClear"},
		{"game over", domain.MissMilliganPhaseGameOver, "missmilligan.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockMissMilliganGame)
			setupMissMilliganOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMissMilliganWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.MissMilliganHint{FromZone: "tableau", FromCol: 2, CardIndex: 1, ToZone: "foundation", ToIdx: 0}

	g := new(interfaces.MockMissMilliganGame)
	setupMissMilliganWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).Output(g, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestMissMilliganWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganWebMockDefaults(g)
		g.On("GetHint").Return(&domain.MissMilliganHint{
			FromZone: "waived", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "waived", result.Hint.FromZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "missmilligan.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganWebMockDefaults(g)
		g.On("GetHint").Return((*domain.MissMilliganHint)(nil))

		result := parseMissMilliganOutput(t, new(MissMilliganWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "missmilligan.noHint", result.MessageCode)
	})
}

func TestMissMilliganWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		g.On("GetPhase").Return(domain.MissMilliganPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(MissMilliganWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		g.On("GetPhase").Return(domain.MissMilliganPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(MissMilliganWebPresenter).ActionLogOutput(g), "move")
	})
}
