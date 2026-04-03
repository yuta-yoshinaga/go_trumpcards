//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGolfCuiMockDefaults(gg *interfaces.MockGolfGame) {
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(0).Maybe()
	gg.On("GetStockCount").Return(16).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()

	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	// Add some cards
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			layout[c][r] = &domain.GolfCard{
				Card:    domain.NewCard(domain.CardDesignSpade, (c*5+r)%13+1, false),
				Removed: false,
			}
		}
	}
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			if r == domain.GolfRowCnt-1 {
				gg.On("IsExposed", c, r).Return(true).Maybe()
			} else {
				gg.On("IsExposed", c, r).Return(false).Maybe()
			}
		}
	}
}

func TestGolfCuiPresenterOutput_Playing(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	p := &GolfCuiPresenter{}

	result := p.Output(gg, nil)
	assert.Contains(t, result, "Golf")
	assert.Contains(t, result, "Stock: 16枚")
	assert.Contains(t, result, "手数: 0")
}

func TestGolfCuiPresenterOutput_Error(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	p := &GolfCuiPresenter{}

	result := p.Output(gg, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestGolfCuiPresenterOutput_Stalemate(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(true).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "手詰まり")
}

func TestGolfCuiPresenterOutput_GameClear(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameClear).Maybe()
	gg.On("GetMoveCount").Return(10).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "ゲームクリア")
}

func TestGolfCuiPresenterOutput_GameOver(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameOver).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestGolfCuiPresenterOutput_WithWaste(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfCuiMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(1).Maybe()
	gg.On("GetStockCount").Return(15).Maybe()
	gg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)}).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "Waste:")
	assert.NotContains(t, result, "[空]")
}

func TestGolfCuiPresenterHintOutput(t *testing.T) {
	t.Run("remove hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "remove", Col: 3})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "カード除去")
	})

	t.Run("draw hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "draw", Col: -1})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return((*domain.GolfHint)(nil))

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "unknown"})

		p := &GolfCuiPresenter{}
		result := p.HintOutput(gg)
		assert.Contains(t, result, "不明")
	})
}

func TestGolfCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhasePlaying)

		p := &GolfCuiPresenter{}
		result := p.ActionLogOutput(gg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhaseGameOver)
		gg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &GolfCuiPresenter{}
		result := p.ActionLogOutput(gg)
		assert.NotEmpty(t, result)
	})
}
