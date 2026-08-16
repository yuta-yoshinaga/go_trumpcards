//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupMonteCarloCuiMockDefaults(g *interfaces.MockMonteCarloGame) {
	g.On("GetPhase").Return(domain.MonteCarloPhasePlaying).Maybe()
	g.On("GetStockCount").Return(27).Maybe()
	g.On("GetRemovedCount").Return(0).Maybe()
	g.On("GetDealCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	board[0][0] = domain.NewCard(domain.CardDesignSpade, 7, false)
	g.On("GetBoard").Return(board).Maybe()
}

func TestMonteCarloCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		setupMonteCarloCuiMockDefaults(g)
		p := new(MonteCarloCuiPresenter)
		result := p.Output(g, nil)
		assert.Contains(t, result, "Monte Carlo")
		assert.Contains(t, result, "(0,0)")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		setupMonteCarloCuiMockDefaults(g)
		p := new(MonteCarloCuiPresenter)
		result := p.Output(g, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("GetPhase").Return(domain.MonteCarloPhasePlaying).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetRemovedCount").Return(40).Maybe()
		g.On("GetDealCount").Return(3).Maybe()
		g.On("IsStalemate").Return(true).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
		g.On("GetBoard").Return(board).Maybe()

		p := new(MonteCarloCuiPresenter)
		result := p.Output(g, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("GetPhase").Return(domain.MonteCarloPhaseGameClear).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetRemovedCount").Return(52).Maybe()
		g.On("GetDealCount").Return(5).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
		g.On("GetBoard").Return(board).Maybe()

		p := new(MonteCarloCuiPresenter)
		result := p.Output(g, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("GetPhase").Return(domain.MonteCarloPhaseGameOver).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetRemovedCount").Return(20).Maybe()
		g.On("GetDealCount").Return(2).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
		g.On("GetBoard").Return(board).Maybe()

		p := new(MonteCarloCuiPresenter)
		result := p.Output(g, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestMonteCarloCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with remove hint includes both card faces", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("Hint").Return(&domain.MonteCarloHint{
			Action: domain.MonteCarloHintActionRemove,
			FromR:  0, FromC: 1, ToR: 1, ToC: 2,
		})
		var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
		board[0][1] = domain.NewCard(domain.CardDesignSpade, 7, false)
		board[1][2] = domain.NewCard(domain.CardDesignHeart, 7, false)
		g.On("GetBoard").Return(board)

		p := new(MonteCarloCuiPresenter)
		result := p.HintOutput(g)
		assert.Contains(t, result, "(0,1)")
		assert.Contains(t, result, "(1,2)")
		assert.Contains(t, result, cuiCardStr(board[0][1]))
		assert.Contains(t, result, cuiCardStr(board[1][2]))
	})

	t.Run("with remove hint falls back to coordinates when a cell is empty", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("Hint").Return(&domain.MonteCarloHint{
			Action: domain.MonteCarloHintActionRemove,
			FromR:  0, FromC: 1, ToR: 1, ToC: 2,
		})
		var board [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card // all nil
		g.On("GetBoard").Return(board)

		p := new(MonteCarloCuiPresenter)
		result := p.HintOutput(g)
		assert.Contains(t, result, "(0,1)")
		assert.Contains(t, result, "(1,2)")
	})

	t.Run("with deal hint", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("Hint").Return(&domain.MonteCarloHint{Action: domain.MonteCarloHintActionDeal})

		p := new(MonteCarloCuiPresenter)
		result := p.HintOutput(g)
		assert.Contains(t, result, "山札")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("Hint").Return((*domain.MonteCarloHint)(nil))

		p := new(MonteCarloCuiPresenter)
		result := p.HintOutput(g)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestMonteCarloCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("GetPhase").Return(domain.MonteCarloPhasePlaying)

		p := new(MonteCarloCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockMonteCarloGame)
		g.On("GetPhase").Return(domain.MonteCarloPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "remove", Detail: "test"},
		})

		p := new(MonteCarloCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
