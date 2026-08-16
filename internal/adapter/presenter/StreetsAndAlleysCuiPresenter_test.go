//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupStreetsAndAlleysCuiMockDefaults(bg *interfaces.MockStreetsAndAlleysGame) {
	bg.On("GetPhase").Return(domain.StreetsAndAlleysPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
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

func TestStreetsAndAlleysCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		p := new(StreetsAndAlleysCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Streets and Alleys")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		p := new(StreetsAndAlleysCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameClear)

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameOver)

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0).Maybe()

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		bg.On("GetTableau").Return(emptyTableau)

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty foundation", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var emptyFoundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		bg.On("GetFoundation").Return(emptyFoundation)

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestStreetsAndAlleysCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetHint").Return(&domain.StreetsAndAlleysHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetHint").Return(&domain.StreetsAndAlleysHint{
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetHint").Return((*domain.StreetsAndAlleysHint)(nil))

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestStreetsAndAlleysCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhasePlaying)

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockStreetsAndAlleysGame)
		bg.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(StreetsAndAlleysCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
