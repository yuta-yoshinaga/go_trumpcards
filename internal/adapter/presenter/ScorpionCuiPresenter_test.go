//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupScorpionCuiMockDefaults(sg *interfaces.MockScorpionGame) {
	sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetStockCount").Return(3).Maybe()

	var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.ScorpionTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestScorpionCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionCuiMockDefaults(sg)
		p := new(ScorpionCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Scorpion")
		assert.Contains(t, result, "Completed")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionCuiMockDefaults(sg)
		p := new(ScorpionCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		sg.On("GetMoveCount").Return(5).Maybe()
		sg.On("IsStalemate").Return(true).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameClear).Maybe()
		sg.On("GetMoveCount").Return(42).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(4).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver).Maybe()
		sg.On("GetMoveCount").Return(10).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestScorpionCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("deal hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: -1, CardIndex: -1, ToCol: -1})

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return((*domain.ScorpionHint)(nil))

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestScorpionCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying)

		p := new(ScorpionCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(ScorpionCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})
}
