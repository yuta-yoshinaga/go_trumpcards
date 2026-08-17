//go:build test

package presenter

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestClockSolitaireCuiPresenterOutput_Playing(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	p := &ClockSolitaireCuiPresenter{}

	result := p.Output(gg, nil)
	assert.Contains(t, result, "Clock Solitaire")
	assert.Contains(t, result, "ステップ: 0")
	// Current card is SPADE 5 → placement hint points to the 5 o'clock pile.
	assert.Contains(t, result, "5時の山へ")
}

func TestClockSolitaireCuiPresenterOutput_KingPlacement(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.ClockSolitairePhasePlaying)
	gg.On("GetStepCount").Return(3)
	gg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false))

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, domain.ClockSolitaireCardsPerPile)
		for j := range domain.ClockSolitaireCardsPerPile {
			piles[i][j] = &domain.ClockSolitaireCard{
				Card:   domain.NewCard(domain.CardDesignSpade, i+1, false),
				FaceUp: true,
			}
		}
		fuc[i] = 4
	}
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)

	p := &ClockSolitaireCuiPresenter{}
	result := p.Output(gg, nil)
	assert.Contains(t, result, "中央(K)の山へ")
	assert.NotContains(t, result, "時の山へ")
}

func TestClockSolitaireCuiPresenterOutput_Error(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	p := &ClockSolitaireCuiPresenter{}

	result := p.Output(gg, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestClockSolitaireCuiPresenterOutput_GameClear(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.ClockSolitairePhaseGameClear)
	gg.On("GetStepCount").Return(42)
	gg.On("GetCurrentCard").Return((*domain.Card)(nil))

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, domain.ClockSolitaireCardsPerPile)
		for j := range domain.ClockSolitaireCardsPerPile {
			piles[i][j] = &domain.ClockSolitaireCard{
				Card:   domain.NewCard(domain.CardDesignSpade, i+1, false),
				FaceUp: true,
			}
		}
		fuc[i] = 4
	}
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)

	p := &ClockSolitaireCuiPresenter{}
	result := p.Output(gg, nil)

	assert.Contains(t, result, "ゲームクリア")
	assert.Contains(t, result, "42")
}

func TestClockSolitaireCuiPresenterOutput_GameOver(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.ClockSolitairePhaseGameOver)
	gg.On("GetStepCount").Return(30)
	gg.On("GetCurrentCard").Return((*domain.Card)(nil))

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, domain.ClockSolitaireCardsPerPile)
		for j := range domain.ClockSolitaireCardsPerPile {
			piles[i][j] = &domain.ClockSolitaireCard{
				Card:   domain.NewCard(domain.CardDesignSpade, i+1, false),
				FaceUp: j == 0,
			}
		}
		fuc[i] = 1
	}
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)

	p := &ClockSolitaireCuiPresenter{}
	result := p.Output(gg, nil)

	assert.Contains(t, result, "ゲームオーバー")
}

func TestClockSolitaireCuiPresenterActionLog(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	gg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, Detail: "テスト"},
	})

	p := &ClockSolitaireCuiPresenter{}
	result := p.ActionLogOutput(gg)

	assert.Contains(t, result, "Action Log")
	assert.Contains(t, result, "テスト")
}

// #5523: 「あと何山で揃うか」は CLI ターミナルを開いたときだけ見える隠れた
// 計算式で、通常プレイでは CUI にも Web にも出ていなかった。
func TestClockSolitaireCuiPresenterOutput_CompletedCount(t *testing.T) {
	outputWithCompleted := func(completed int) string {
		gg := new(interfaces.MockClockSolitaireGame)
		setupClockSolitaireWebMockDefaults(gg)
		gg.ExpectedCalls = nil
		gg.On("GetPhase").Return(domain.ClockSolitairePhasePlaying)
		gg.On("GetStepCount").Return(1)
		gg.On("GetCurrentCard").Return((*domain.Card)(nil))

		var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
		var fuc [domain.ClockSolitairePileCount]int
		for i := range domain.ClockSolitairePileCount {
			piles[i] = make([]*domain.ClockSolitaireCard, domain.ClockSolitaireCardsPerPile)
			for j := range domain.ClockSolitaireCardsPerPile {
				piles[i][j] = &domain.ClockSolitaireCard{
					Card:   domain.NewCard(domain.CardDesignSpade, i+1, false),
					FaceUp: i < completed,
				}
			}
			if i < completed {
				fuc[i] = domain.ClockSolitaireCardsPerPile
			}
		}
		gg.On("GetPiles").Return(piles)
		gg.On("GetFaceUpCount").Return(fuc)
		return (&ClockSolitaireCuiPresenter{}).Output(gg, nil)
	}

	line := func(completed int) string {
		return i18n.Tf("clocksolitaire.completedPiles",
			"completed", strconv.Itoa(completed),
			"total", strconv.Itoa(domain.ClockSolitairePileCount))
	}

	assert.Contains(t, outputWithCompleted(0), line(0))
	assert.Contains(t, outputWithCompleted(5), line(5))
	// **クリア時は 13/13。**中央のKの山も数に入る。
	assert.Contains(t, outputWithCompleted(domain.ClockSolitairePileCount), line(domain.ClockSolitairePileCount))
	// 4枚に届いていない山は数えない。
	assert.NotContains(t, outputWithCompleted(5), line(6))
}
