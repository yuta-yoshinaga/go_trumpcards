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

func setupFortressWebMockDefaults(bg *interfaces.MockFortressGame) {
	bg.On("GetPhase").Return(domain.FortressPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
	for i := range domain.FortressTableauCnt {
		tableau[i] = make([]*domain.FortressTableauCard, domain.FortressColumnLen)
		for j := range domain.FortressColumnLen {
			tableau[i][j] = &domain.FortressTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortressFoundationCnt][]*domain.Card
	for i := range domain.FortressFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseFortressOutput(t *testing.T, jsonStr string) *controller.FortressWebOutput {
	t.Helper()
	var out controller.FortressWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupFortressOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupFortressOutputMock(g *interfaces.MockFortressGame) {
	setupFortressWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestFortressWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		p := new(FortressWebPresenter)

		result := parseFortressOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.FortressTableauCnt)
		assert.Len(t, result.Foundation, domain.FortressFoundationCnt)
		assert.Equal(t, "fortress.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		p := new(FortressWebPresenter)

		result := parseFortressOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		p := new(FortressWebPresenter)

		result := parseFortressOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FortressPhaseGameClear)

		p := new(FortressWebPresenter)
		result := parseFortressOutput(t, p.Output(bg, nil))
		assert.Equal(t, "fortress.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FortressPhaseGameOver)

		p := new(FortressWebPresenter)
		result := parseFortressOutput(t, p.Output(bg, nil))
		assert.Equal(t, "fortress.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(FortressWebPresenter)
		result := parseFortressOutput(t, p.Output(bg, nil))
		assert.Equal(t, "fortress.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFortressWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.FortressHint{FromCol: 2, CardIndex: 2, ToZone: "tableau", ToCol: 2}

	bg := new(interfaces.MockFortressGame)
	setupFortressWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseFortressOutput(t, new(FortressWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestFortressWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.FortressHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortressWebPresenter)
		result := parseFortressOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "fortress.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.FortressHint)(nil))

		p := new(FortressWebPresenter)
		result := parseFortressOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "fortress.noHint", result.MessageCode)
	})
}

func TestFortressWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetPhase").Return(domain.FortressPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(FortressWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetPhase").Return(domain.FortressPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(FortressWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
