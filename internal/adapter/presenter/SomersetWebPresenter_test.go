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

func setupSomersetWebMockDefaults(bg *interfaces.MockSomersetGame) {
	bg.On("GetPhase").Return(domain.SomersetPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	for i := range domain.SomersetTableauCnt {
		tableau[i] = make([]*domain.SomersetTableauCard, domain.SomersetColumnLen)
		for j := range domain.SomersetColumnLen {
			tableau[i][j] = &domain.SomersetTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.SomersetFoundationCnt][]*domain.Card
	for i := range domain.SomersetFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseSomersetOutput(t *testing.T, jsonStr string) *controller.SomersetWebOutput {
	t.Helper()
	var out controller.SomersetWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupSomersetOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupSomersetOutputMock(g *interfaces.MockSomersetGame) {
	setupSomersetWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSomersetWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		p := new(SomersetWebPresenter)

		result := parseSomersetOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.SomersetTableauCnt)
		assert.Len(t, result.Foundation, domain.SomersetFoundationCnt)
		assert.Equal(t, "somerset.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		p := new(SomersetWebPresenter)

		result := parseSomersetOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		p := new(SomersetWebPresenter)

		result := parseSomersetOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.SomersetPhaseGameClear)

		p := new(SomersetWebPresenter)
		result := parseSomersetOutput(t, p.Output(bg, nil))
		assert.Equal(t, "somerset.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.SomersetPhaseGameOver)

		p := new(SomersetWebPresenter)
		result := parseSomersetOutput(t, p.Output(bg, nil))
		assert.Equal(t, "somerset.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(SomersetWebPresenter)
		result := parseSomersetOutput(t, p.Output(bg, nil))
		assert.Equal(t, "somerset.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSomersetWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.SomersetHint{FromCol: 2, CardIndex: 2, ToZone: "tableau", ToCol: 2}

	bg := new(interfaces.MockSomersetGame)
	setupSomersetWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseSomersetOutput(t, new(SomersetWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestSomersetWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.SomersetHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(SomersetWebPresenter)
		result := parseSomersetOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "somerset.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		setupSomersetWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.SomersetHint)(nil))

		p := new(SomersetWebPresenter)
		result := parseSomersetOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "somerset.noHint", result.MessageCode)
	})
}

func TestSomersetWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		bg.On("GetPhase").Return(domain.SomersetPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(SomersetWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockSomersetGame)
		bg.On("GetPhase").Return(domain.SomersetPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(SomersetWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
