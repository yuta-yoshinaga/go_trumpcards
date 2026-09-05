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

func setupCitadelWebMockDefaults(bg *interfaces.MockCitadelGame) {
	bg.On("GetPhase").Return(domain.CitadelPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
	for i := range domain.CitadelTableauCnt {
		tableau[i] = make([]*domain.CitadelTableauCard, domain.CitadelMaxColumnLen)
		for j := range domain.CitadelMaxColumnLen {
			tableau[i][j] = &domain.CitadelTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CitadelFoundationCnt][]*domain.Card
	for i := range domain.CitadelFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseCitadelOutput(t *testing.T, jsonStr string) *controller.CitadelWebOutput {
	t.Helper()
	var out controller.CitadelWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupCitadelOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupCitadelOutputMock(g *interfaces.MockCitadelGame) {
	setupCitadelWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCitadelWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		p := new(CitadelWebPresenter)

		result := parseCitadelOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.CitadelTableauCnt)
		assert.Len(t, result.Foundation, domain.CitadelFoundationCnt)
		assert.Equal(t, "citadel.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		p := new(CitadelWebPresenter)

		result := parseCitadelOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		p := new(CitadelWebPresenter)

		result := parseCitadelOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.CitadelPhaseGameClear)

		p := new(CitadelWebPresenter)
		result := parseCitadelOutput(t, p.Output(bg, nil))
		assert.Equal(t, "citadel.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.CitadelPhaseGameOver)

		p := new(CitadelWebPresenter)
		result := parseCitadelOutput(t, p.Output(bg, nil))
		assert.Equal(t, "citadel.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(CitadelWebPresenter)
		result := parseCitadelOutput(t, p.Output(bg, nil))
		assert.Equal(t, "citadel.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCitadelWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CitadelHint{FromCol: 2, CardIndex: 2, ToZone: "tableau", ToCol: 2}

	bg := new(interfaces.MockCitadelGame)
	setupCitadelWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseCitadelOutput(t, new(CitadelWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestCitadelWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.CitadelHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(CitadelWebPresenter)
		result := parseCitadelOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "citadel.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		setupCitadelWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.CitadelHint)(nil))

		p := new(CitadelWebPresenter)
		result := parseCitadelOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "citadel.noHint", result.MessageCode)
	})
}

func TestCitadelWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		bg.On("GetPhase").Return(domain.CitadelPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(CitadelWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockCitadelGame)
		bg.On("GetPhase").Return(domain.CitadelPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(CitadelWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
