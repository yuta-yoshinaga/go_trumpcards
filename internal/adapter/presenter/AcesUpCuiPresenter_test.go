//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAcesUpCuiMockDefaults(g *interfaces.MockAcesUpGame) {
	g.On("GetPhase").Return(domain.AcesUpPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetDiscardCount").Return(4).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetColumns").Return(sampleAcesUpColumns()).Maybe()
}

func TestAcesUpCuiPresenterOutput_Playing(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	setupAcesUpCuiMockDefaults(g)
	p := &AcesUpCuiPresenter{}

	result := p.Output(g, nil)
	assert.Contains(t, result, "Aces Up")
	assert.Contains(t, result, "Stock: 44枚")
	assert.Contains(t, result, "Discard: 4枚")
	assert.Contains(t, result, "手数: 0")
	assert.Contains(t, result, "[空]") // col2 is empty
}

func TestAcesUpCuiPresenterOutput_Error(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	setupAcesUpCuiMockDefaults(g)
	p := &AcesUpCuiPresenter{}
	assert.Contains(t, p.Output(g, errors.New("test error")), "test error")
}

func TestAcesUpCuiPresenterOutput_Stalemate(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(10).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "手詰まり")
}

func TestAcesUpCuiPresenterOutput_GameClear(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(20).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(48).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "ゲームクリア")
}

func TestAcesUpCuiPresenterOutput_GameOver(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(2).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "ゲームオーバー")
}

func TestAcesUpCuiPresenterHintOutput(t *testing.T) {
	t.Run("remove hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return(&domain.AcesUpHint{Type: "remove", Col: 0})
		p := &AcesUpCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "カード除去")
	})

	t.Run("move hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return(&domain.AcesUpHint{Type: "move", Col: 2})
		p := &AcesUpCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "空き列")
	})

	t.Run("draw hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return(&domain.AcesUpHint{Type: "draw", Col: -1})
		p := &AcesUpCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "配って")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return((*domain.AcesUpHint)(nil))
		p := &AcesUpCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return(&domain.AcesUpHint{Type: "unknown"})
		p := &AcesUpCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "不明")
	})
}

func TestAcesUpCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetPhase").Return(domain.AcesUpPhasePlaying)
		p := &AcesUpCuiPresenter{}
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetPhase").Return(domain.AcesUpPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
		p := &AcesUpCuiPresenter{}
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
