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

func setupFortyThievesWebMockDefaults(fg *interfaces.MockFortyThievesGame) {
	fg.On("GetPhase").Return(domain.FortyThievesPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(0).Maybe()
	fg.On("GetStockCount").Return(64).Maybe()
	fg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	fg.On("CanUndo").Return(false).Maybe()
	fg.On("IsStalemate").Return(false).Maybe()
	fg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
	for i := range domain.FortyThievesTableauCnt {
		tableau[i] = make([]*domain.FortyThievesTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.FortyThievesTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	fg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
}

func parseFortyThievesOutput(t *testing.T, jsonStr string) *controller.FortyThievesWebOutput {
	t.Helper()
	var out controller.FortyThievesWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupFortyThievesOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupFortyThievesOutputMock(g *interfaces.MockFortyThievesGame) {
	setupFortyThievesWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestFortyThievesWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		p := new(FortyThievesWebPresenter)

		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 64, result.StockCount)
		assert.Empty(t, result.Waste)
		assert.Len(t, result.Tableau, domain.FortyThievesTableauCnt)
		assert.Len(t, result.Foundation, domain.FortyThievesFoundationCnt)
		assert.Equal(t, "fortythieves.playing", result.MessageCode)
	})

	t.Run("waste with cards", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetWaste")
		fg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "HEART", result.Waste[0].Design)
		assert.Equal(t, 5, result.Waste[0].Value)
	})

	t.Run("all face up", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		p := new(FortyThievesWebPresenter)

		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		// All cards should be face up
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		p := new(FortyThievesWebPresenter)

		result := parseFortyThievesOutput(t, p.Output(fg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameClear)

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortythieves.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameOver)

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortythieves.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesOutputMock(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "IsStalemate")
		fg.On("IsStalemate").Return(true)

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortythieves.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFortyThievesWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		ftg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesWebMockDefaults(ftg)
		ftg.On("GetHint").Return(&domain.FortyThievesHint{FromZone: "tableau", FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(FortyThievesWebPresenter).Output(ftg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		ftg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesWebMockDefaults(ftg)
		ftg.ExpectedCalls = filterCalls(ftg.ExpectedCalls, "IsStalemate")
		ftg.On("IsStalemate").Return(true)
		ftg.On("GetHint").Return(&domain.FortyThievesHint{FromZone: "tableau", FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(FortyThievesWebPresenter).Output(ftg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestFortyThievesWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesWebMockDefaults(fg)
		fg.On("GetHint").Return(&domain.FortyThievesHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.HintOutput(fg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, "fortythieves.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesWebMockDefaults(fg)
		fg.On("GetHint").Return((*domain.FortyThievesHint)(nil))

		p := new(FortyThievesWebPresenter)
		result := parseFortyThievesOutput(t, p.HintOutput(fg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "fortythieves.noHint", result.MessageCode)
	})
}

func TestFortyThievesWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetPhase").Return(domain.FortyThievesPhasePlaying)

		fg.On("GetGameEndFlag").Return(false)
		p := new(FortyThievesWebPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameOver)
		fg.On("GetGameEndFlag").Return(true)
		fg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(FortyThievesWebPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "draw")
	})
}
