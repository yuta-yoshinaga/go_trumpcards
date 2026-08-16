//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTriPeaksCuiMockDefaults(tg *interfaces.MockTriPeaksGame) {
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(0).Maybe()
	tg.On("GetStockCount").Return(23).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()

	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	// Add some cards at row 3
	for c := range 10 {
		layout[3][c] = &domain.TriPeaksCard{
			Card:    domain.NewCard(domain.CardDesignSpade, c%13+1, false),
			Removed: false,
		}
		// Even columns exposed, odd columns blocked — exercises both formats.
		tg.On("IsExposed", 3, c).Return(c%2 == 0).Maybe()
	}
	tg.On("GetLayout").Return(layout).Maybe()
}

func TestTriPeaksCuiPresenterOutput_Playing(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	p := &TriPeaksCuiPresenter{}

	result := p.Output(tg, nil)
	assert.Contains(t, result, "TriPeaks")
	assert.Contains(t, result, "Stock: 23枚")
	assert.Contains(t, result, "手数: 0")
	// No waste top -> nothing playable; stock remains, so draw is recommended.
	assert.Contains(t, result, "今出せるカード: 0枚")
	assert.Contains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_PlayableAndBlocked(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(0).Maybe()
	tg.On("GetStockCount").Return(10).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	// Waste top is a 2: adjacent to Ace(1) and 3, with K-A wrap also possible.
	tg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)}).Maybe()

	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	// (3,0)=A exposed & adjacent->playable; (3,1)=5 blocked; (3,2)=9 exposed but not adjacent.
	layout[3][0] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false)}
	layout[3][1] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 5, false)}
	layout[3][2] = &domain.TriPeaksCard{Card: domain.NewCard(domain.CardDesignSpade, 9, false)}
	tg.On("GetLayout").Return(layout).Maybe()
	tg.On("IsExposed", 3, 0).Return(true).Maybe()
	tg.On("IsExposed", 3, 1).Return(false).Maybe()
	tg.On("IsExposed", 3, 2).Return(true).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	// (3,0) is exposed and ±1 from the waste top -> playable marker.
	assert.Contains(t, result, "(3,0)")
	assert.Contains(t, result, "*")
	// (3,1) is blocked -> coordinates hidden.
	assert.Contains(t, result, "[--]")
	// (3,2) is exposed but not adjacent -> plain coordinate format, no marker.
	assert.Contains(t, result, "(3,2)")
	// Exactly one exposed card is playable, so no draw recommendation.
	assert.Contains(t, result, "今出せるカード: 1枚")
	assert.NotContains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_Error(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	p := &TriPeaksCuiPresenter{}

	result := p.Output(tg, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestTriPeaksCuiPresenterOutput_Stalemate(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(true).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "手詰まり")
	// Nothing playable but the stock is empty, so no draw recommendation.
	assert.Contains(t, result, "今出せるカード: 0枚")
	assert.NotContains(t, result, "ドロー推奨")
}

func TestTriPeaksCuiPresenterOutput_GameClear(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameClear).Maybe()
	tg.On("GetMoveCount").Return(10).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "ゲームクリア")
}

func TestTriPeaksCuiPresenterOutput_GameOver(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestTriPeaksCuiPresenterOutput_WithWaste(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksCuiMockDefaults(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(1).Maybe()
	tg.On("GetStockCount").Return(22).Maybe()
	tg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)}).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()

	p := &TriPeaksCuiPresenter{}
	result := p.Output(tg, nil)
	assert.Contains(t, result, "Waste:")
	assert.NotContains(t, result, "[空]")
}

func TestTriPeaksCuiPresenterHintOutput(t *testing.T) {
	t.Run("remove hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "remove", Row: 3, Col: 0})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "カード除去")
	})

	t.Run("draw hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "draw", Row: -1, Col: -1})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return((*domain.TriPeaksHint)(nil))

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "unknown"})

		p := &TriPeaksCuiPresenter{}
		result := p.HintOutput(tg)
		assert.Contains(t, result, "不明")
	})
}

func TestTriPeaksCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying)

		p := &TriPeaksCuiPresenter{}
		result := p.ActionLogOutput(tg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver)
		tg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &TriPeaksCuiPresenter{}
		result := p.ActionLogOutput(tg)
		assert.NotEmpty(t, result)
	})
}
