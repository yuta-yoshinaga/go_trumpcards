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

func setupBeleagueredCastleWebMockDefaults(bg *interfaces.MockBeleagueredCastleGame) {
	bg.On("GetPhase").Return(domain.BeleagueredCastlePhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	for i := range domain.BeleagueredCastleTableauCnt {
		tableau[i] = make([]*domain.BeleagueredCastleTableauCard, domain.BeleagueredCastleColumnLen)
		for j := range domain.BeleagueredCastleColumnLen {
			tableau[i][j] = &domain.BeleagueredCastleTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
	for i := range domain.BeleagueredCastleFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseBeleagueredCastleOutput(t *testing.T, jsonStr string) *controller.BeleagueredCastleWebOutput {
	t.Helper()
	var out controller.BeleagueredCastleWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupBeleagueredCastleOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupBeleagueredCastleOutputMock(g *interfaces.MockBeleagueredCastleGame) {
	setupBeleagueredCastleWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestBeleagueredCastleWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		p := new(BeleagueredCastleWebPresenter)

		result := parseBeleagueredCastleOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.BeleagueredCastleTableauCnt)
		assert.Len(t, result.Foundation, domain.BeleagueredCastleFoundationCnt)
		assert.Equal(t, "beleagueredcastle.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		p := new(BeleagueredCastleWebPresenter)

		result := parseBeleagueredCastleOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		p := new(BeleagueredCastleWebPresenter)

		result := parseBeleagueredCastleOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameClear)

		p := new(BeleagueredCastleWebPresenter)
		result := parseBeleagueredCastleOutput(t, p.Output(bg, nil))
		assert.Equal(t, "beleagueredcastle.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameOver)

		p := new(BeleagueredCastleWebPresenter)
		result := parseBeleagueredCastleOutput(t, p.Output(bg, nil))
		assert.Equal(t, "beleagueredcastle.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(BeleagueredCastleWebPresenter)
		result := parseBeleagueredCastleOutput(t, p.Output(bg, nil))
		assert.Equal(t, "beleagueredcastle.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBeleagueredCastleWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.BeleagueredCastleHint{FromCol: 2, CardIndex: 2, ToZone: "tableau", ToCol: 2}

	bg := new(interfaces.MockBeleagueredCastleGame)
	setupBeleagueredCastleWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseBeleagueredCastleOutput(t, new(BeleagueredCastleWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestBeleagueredCastleWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.BeleagueredCastleHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(BeleagueredCastleWebPresenter)
		result := parseBeleagueredCastleOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "beleagueredcastle.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.BeleagueredCastleHint)(nil))

		p := new(BeleagueredCastleWebPresenter)
		result := parseBeleagueredCastleOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "beleagueredcastle.noHint", result.MessageCode)
	})
}

func TestBeleagueredCastleWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(BeleagueredCastleWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BeleagueredCastleWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
