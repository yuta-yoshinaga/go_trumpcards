//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestShengJiCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockShengJiInteractor {
		m := new(mockUsecases.MockShengJiInteractor)
		m.On("GetConfig").Return(domain.DefaultShengJiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Declare", mock.Anything).Return(mockOutput)
		m.On("BuryKitty", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewShengJiCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultShengJiConfig())
	})

	// **0 はパス。**下限を 1 にすると亮牌を降りられなくなる。
	t.Run("declares, and zero passes", func(t *testing.T) {
		m := newMock()
		c := controller.NewShengJiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Declare", domain.CardDesignHeart)
		assert.Equal(t, mockOutput, c.Exec("declare 0"))
		m.AssertCalled(t, "Declare", domain.ShengJiNoTrump)
	})

	t.Run("declare rejects bad input", func(t *testing.T) {
		c := controller.NewShengJiCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgStem("suitRequiredOrPass"))
		assert.Contains(t, c.Exec("d abc"), msgStem("invalidSuit"))
		assert.Contains(t, c.Exec("d 5"), msgStem("invalidSuit"))
		assert.Contains(t, c.Exec("d -1"), msgStem("invalidSuit"))
	})

	// **底牌はちょうど 8 枚。**
	t.Run("buries exactly eight", func(t *testing.T) {
		m := newMock()
		c := controller.NewShengJiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 0 1 2 3 4 5 6 7"))
		m.AssertCalled(t, "BuryKitty", []int{0, 1, 2, 3, 4, 5, 6, 7})
		// **底牌を拾った直後は 25 + 8 枚**あるので、埋め戻しの上限は広い。
		assert.Equal(t, mockOutput, c.Exec("b 25 26 27 28 29 30 31 32"))
		assert.Contains(t, c.Exec("b 33 0 1 2 3 4 5 6"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("b 0 1"), "exactly 8")
		assert.Contains(t, c.Exec("b"), msgStem("cardIndexesRequiredPair"))
	})

	t.Run("plays any number of cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewShengJiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", []int{3})
		assert.Equal(t, mockOutput, c.Exec("play 0 1"))
		m.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("play rejects bad input", func(t *testing.T) {
		c := controller.NewShengJiCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgStem("cardIndexesRequiredPair"))
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p -1"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 99"), msgInvalidCardIndexPrefix())
		// **プレイ中の手札は 25 枚。**底牌を拾った直後の上限 (32) を使い回すと緩い。
		assert.Contains(t, c.Exec("p 25"), msgInvalidCardIndexPrefix())
		assert.Equal(t, mockOutput, c.Exec("p 24"))
		// **同じ札を 2 回数えられない。**通すと 1 枚から対子が作れてしまう。
		assert.Contains(t, c.Exec("p 1 1"), "twice")
	})

	t.Run("next, log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewShengJiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("next"))
		m.AssertNumberOfCalls(t, "NextHand", 2)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
