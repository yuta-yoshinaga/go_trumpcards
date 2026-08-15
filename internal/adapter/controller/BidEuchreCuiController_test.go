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

func TestBidEuchreCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBidEuchreInteractor {
		m := new(mockUsecases.MockBidEuchreInteractor)
		m.On("GetConfig").Return(domain.DefaultBidEuchreConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("PassBid").Return(mockOutput)
		m.On("ChooseTrump", mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewBidEuchreCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBidEuchreConfig())
	})

	// **最低ビッドは 3。**2 以下は入口で弾く。
	t.Run("bids a trick count", func(t *testing.T) {
		m := newMock()
		c := controller.NewBidEuchreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 3"))
		m.AssertCalled(t, "Bid", domain.BidEuchreMinBid)
		assert.Equal(t, mockOutput, c.Exec("bid 6"))
		m.AssertCalled(t, "Bid", domain.BidEuchreMaxBid)
		assert.Equal(t, mockOutput, c.Exec("ps"))
		m.AssertCalled(t, "PassBid")
	})

	t.Run("bid rejects a value outside 3-6", func(t *testing.T) {
		c := controller.NewBidEuchreCuiController(newMock())
		assert.Contains(t, c.Exec("b"), msgStem("bidValueRequired"))
		assert.Contains(t, c.Exec("b abc"), msgStem("invalidBid"))
		assert.Contains(t, c.Exec("b 2"), msgStem("invalidBid"))
		assert.Contains(t, c.Exec("b 7"), msgStem("invalidBid"))
	})

	// **切札は 6 種類。**ノートランプがハイとローで 2 つある。
	t.Run("names trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewBidEuchreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 0"))
		m.AssertCalled(t, "ChooseTrump", int(domain.BidEuchreTrumpSpade))
		assert.Equal(t, mockOutput, c.Exec("trump 5"))
		m.AssertCalled(t, "ChooseTrump", int(domain.BidEuchreTrumpNoLow))
		assert.Contains(t, c.Exec("t"), msgStem("trumpDeclarationRequired"))
		assert.Contains(t, c.Exec("t abc"), msgStem("invalidTrump"))
		assert.Contains(t, c.Exec("t 6"), msgStem("invalidTrump"))
		assert.Contains(t, c.Exec("t -1"), msgStem("invalidTrump"))
	})

	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewBidEuchreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 5"))
		m.AssertCalled(t, "PlayCard", 5)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 6"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewBidEuchreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
