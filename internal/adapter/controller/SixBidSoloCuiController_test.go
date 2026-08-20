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

func TestSixBidSoloCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSixBidSoloInteractor {
		m := new(mockUsecases.MockSixBidSoloInteractor)
		m.On("GetConfig").Return(domain.DefaultSixBidSoloConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("PassBid").Return(mockOutput)
		m.On("Declare", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSixBidSoloConfig())
	})

	// **1=ソロ … 6=コール。**6 段階すべて指せる。
	t.Run("bids one of the six levels", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 1"))
		m.AssertCalled(t, "Bid", int(domain.SixBidSoloBidSolo))
		assert.Equal(t, mockOutput, c.Exec("bid 6"))
		m.AssertCalled(t, "Bid", int(domain.SixBidSoloBidCall))
		assert.Equal(t, mockOutput, c.Exec("ps"))
		m.AssertCalled(t, "PassBid")
	})

	t.Run("bid rejects a value outside 1-6", func(t *testing.T) {
		c := controller.NewSixBidSoloCuiController(newMock())
		assert.Contains(t, c.Exec("b"), msgStem("bidRequiredSolo"))
		assert.Contains(t, c.Exec("b abc"), msgStem("invalidBid"))
		assert.Contains(t, c.Exec("b 0"), msgStem("invalidBid"))
		assert.Contains(t, c.Exec("b 7"), msgStem("invalidBid"))
	})

	t.Run("declares a trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 1"))
		m.AssertCalled(t, "Declare", domain.CardDesignSpade, 0, 0)
		assert.Equal(t, mockOutput, c.Exec("declare 4"))
		m.AssertCalled(t, "Declare", domain.CardDesignDiamond, 0, 0)
	})

	// **コール・ソロは指名札を続ける。**
	t.Run("a call solo names a card", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 1 3 1"))
		m.AssertCalled(t, "Declare", domain.CardDesignSpade, domain.CardDesignHeart, 1)
	})

	t.Run("declare rejects bad input", func(t *testing.T) {
		c := controller.NewSixBidSoloCuiController(newMock())
		assert.True(t, msgRejected(c.Exec("d")))
		assert.Contains(t, c.Exec("d abc"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("d 0"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("d 5"), msgStem("invalidSuit14Letters"))
		// スートだけ来てランクが無い。
		assert.Contains(t, c.Exec("d 1 3"), "both the called suit")
		assert.Contains(t, c.Exec("d 1 9 1"), msgStem("invalidCalledSuit14"))
		assert.Contains(t, c.Exec("d 1 3 0"), msgStem("invalidCalledValue113"))
		assert.Contains(t, c.Exec("d 1 3 14"), msgStem("invalidCalledValue113"))
		assert.Contains(t, c.Exec("d 1 3 abc"), msgStem("invalidCalledValue113"))
		assert.Contains(t, c.Exec("d 1 abc 1"), msgStem("invalidCalledSuit14"))
	})

	// **手札は 11 枚。**
	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 10"))
		m.AssertCalled(t, "PlayCard", 10)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 11"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewSixBidSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
