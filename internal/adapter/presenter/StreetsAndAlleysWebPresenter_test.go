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

func setupStreetsAndAlleysWebMockDefaults(bg *interfaces.MockStreetsAndAlleysGame) {
	bg.On("GetPhase").Return(domain.StreetsAndAlleysPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
	for i := range domain.StreetsAndAlleysTableauCnt {
		tableau[i] = make([]*domain.StreetsAndAlleysTableauCard, domain.StreetsAndAlleysColumnLen)
		for j := range domain.StreetsAndAlleysColumnLen {
			tableau[i][j] = &domain.StreetsAndAlleysTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
	for i := range domain.StreetsAndAlleysFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func parseStreetsAndAlleysOutput(t *testing.T, jsonStr string) *controller.StreetsAndAlleysWebOutput {
	t.Helper()
	var out controller.StreetsAndAlleysWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupStreetsAndAlleysOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupStreetsAndAlleysOutputMock(g *interfaces.MockStreetsAndAlleysGame) {
	setupStreetsAndAlleysWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestStreetsAndAlleysWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		p := new(StreetsAndAlleysWebPresenter)

		result := parseStreetsAndAlleysOutput(t, p.Output(bg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Len(t, result.Tableau, domain.StreetsAndAlleysTableauCnt)
		assert.Len(t, result.Foundation, domain.StreetsAndAlleysFoundationCnt)
		assert.Equal(t, "streetsandalleys.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		p := new(StreetsAndAlleysWebPresenter)

		result := parseStreetsAndAlleysOutput(t, p.Output(bg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		p := new(StreetsAndAlleysWebPresenter)

		result := parseStreetsAndAlleysOutput(t, p.Output(bg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameClear)

		p := new(StreetsAndAlleysWebPresenter)
		result := parseStreetsAndAlleysOutput(t, p.Output(bg, nil))
		assert.Equal(t, "streetsandalleys.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameOver)

		p := new(StreetsAndAlleysWebPresenter)
		result := parseStreetsAndAlleysOutput(t, p.Output(bg, nil))
		assert.Equal(t, "streetsandalleys.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(StreetsAndAlleysWebPresenter)
		result := parseStreetsAndAlleysOutput(t, p.Output(bg, nil))
		assert.Equal(t, "streetsandalleys.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestStreetsAndAlleysWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.StreetsAndAlleysHint{FromCol: 3, CardIndex: 0, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(StreetsAndAlleysWebPresenter).Output(sg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		sg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysWebMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("GetHint").Return(&domain.StreetsAndAlleysHint{FromCol: 3, CardIndex: 0, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(StreetsAndAlleysWebPresenter).Output(sg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestStreetsAndAlleysWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.StreetsAndAlleysHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(StreetsAndAlleysWebPresenter)
		result := parseStreetsAndAlleysOutput(t, p.HintOutput(bg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "streetsandalleys.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.StreetsAndAlleysHint)(nil))

		p := new(StreetsAndAlleysWebPresenter)
		result := parseStreetsAndAlleysOutput(t, p.HintOutput(bg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "streetsandalleys.noHint", result.MessageCode)
	})
}

func TestStreetsAndAlleysWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)

		p := new(StreetsAndAlleysWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameOver)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(StreetsAndAlleysWebPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
