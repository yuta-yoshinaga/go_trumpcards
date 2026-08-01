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

func setupFlowerGardenWebMockDefaults(bg *interfaces.MockFlowerGardenGame) {
	bg.On("GetPhase").Return(domain.FlowerGardenPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	for i := range domain.FlowerGardenTableauCnt {
		tableau[i] = make([]*domain.FlowerGardenTableauCard, domain.FlowerGardenColumnLen)
		for j := range domain.FlowerGardenColumnLen {
			tableau[i][j] = &domain.FlowerGardenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	reserve := make([]*domain.Card, domain.FlowerGardenReserveCnt)
	for i := range reserve {
		reserve[i] = domain.NewCard(domain.CardDesignHeart, (i%13)+1, false)
	}
	bg.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.FlowerGardenFoundationCnt][]*domain.Card
	for i := range domain.FlowerGardenFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseFlowerGardenOutput(t *testing.T, jsonStr string) *controller.FlowerGardenWebOutput {
	t.Helper()
	var out controller.FlowerGardenWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupFlowerGardenOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupFlowerGardenOutputMock(g *interfaces.MockFlowerGardenGame) {
	setupFlowerGardenWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestFlowerGardenWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		p := new(FlowerGardenWebPresenter)

		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.FlowerGardenTableauCnt)
		assert.Len(t, result.Reserve, domain.FlowerGardenReserveCnt)
		assert.Len(t, result.Foundation, domain.FlowerGardenFoundationCnt)
		assert.Equal(t, "flowergarden.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		p := new(FlowerGardenWebPresenter)

		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		p := new(FlowerGardenWebPresenter)

		result := parseFlowerGardenOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameClear)

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "flowergarden.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameOver)

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "flowergarden.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		assert.Equal(t, "flowergarden.stalemate", result.MessageCode)
	})

	t.Run("depleted reserve cell serialises null", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetReserve")
		bg.On("GetReserve").Return([]*domain.Card{nil, domain.NewCard(domain.CardDesignSpade, 5, false)})

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.Output(bg, nil))
		assert.Len(t, result.Reserve, 2)
		assert.Nil(t, result.Reserve[0])
		assert.NotNil(t, result.Reserve[1])
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFlowerGardenWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.FlowerGardenHint{FromZone: "tableau", FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	bg := new(interfaces.MockFlowerGardenGame)
	setupFlowerGardenWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := parseFlowerGardenOutput(t, new(FlowerGardenWebPresenter).Output(bg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestFlowerGardenWebPresenter_HintOutput(t *testing.T) {
	t.Run("with tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.FlowerGardenHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "flowergarden.hintAvailable", result.MessageCode)
	})

	t.Run("with reserve hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.FlowerGardenHint{
			FromZone:  "reserve",
			FromCol:   2,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     1,
		})

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "reserve", result.Hint.FromZone)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.FlowerGardenHint)(nil))

		p := new(FlowerGardenWebPresenter)
		result := parseFlowerGardenOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "flowergarden.noHint", result.MessageCode)
	})
}

func TestFlowerGardenWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetPhase").Return(domain.FlowerGardenPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(FlowerGardenWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(FlowerGardenWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
