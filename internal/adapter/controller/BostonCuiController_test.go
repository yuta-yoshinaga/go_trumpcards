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

func TestBostonCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBostonInteractor {
		m := new(mockUsecases.MockBostonInteractor)
		m.On("GetConfig").Return(domain.DefaultBostonConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("PassBid").Return(mockOutput)
		m.On("CallPartner", mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewBostonCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBostonConfig())
	})

	// **段の番号で指す。**ミゼールが間に挟まるのでトリック数では一意にならない。
	t.Run("bids by ladder step", func(t *testing.T) {
		m := newMock()
		c := controller.NewBostonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 4 3"))
		m.AssertCalled(t, "Bid", domain.BostonBidSeven, domain.CardDesignHeart)
		// ミゼールはスート省略。
		assert.Equal(t, mockOutput, c.Exec("b 3"))
		m.AssertCalled(t, "Bid", domain.BostonBidLittleMisere, 0)
		assert.Equal(t, mockOutput, c.Exec("ps"))
		m.AssertCalled(t, "PassBid")
	})

	t.Run("bid rejects a bad step or suit", func(t *testing.T) {
		c := controller.NewBostonCuiController(newMock())
		assert.True(t, msgRejected(c.Exec("b")))
		assert.Contains(t, c.Exec("b abc"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 0"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 99"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 1 9"), "Invalid suit")
	})

	// **-1 は「単独で戦う」という有効な選択。**
	t.Run("calls a partner or goes alone", func(t *testing.T) {
		m := newMock()
		c := controller.NewBostonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("cp 2"))
		m.AssertCalled(t, "CallPartner", 2)
		assert.Equal(t, mockOutput, c.Exec("cp -1"))
		m.AssertCalled(t, "CallPartner", -1)
		assert.True(t, msgRejected(c.Exec("cp")))
		assert.Contains(t, c.Exec("cp abc"), "Invalid partner")
		assert.Contains(t, c.Exec("cp 9"), "Invalid partner")
		assert.Contains(t, c.Exec("cp -2"), "Invalid partner")
	})

	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewBostonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 12"))
		m.AssertCalled(t, "PlayCard", 12)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 13"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewBostonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
