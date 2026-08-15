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

func TestVintCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockVintInteractor {
		m := new(mockUsecases.MockVintInteractor)
		m.On("GetConfig").Return(domain.DefaultVintConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("PassBid").Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewVintCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultVintConfig())
	})

	// **denom は 0=♠ 1=♣ 2=♦ 3=♥ 4=NT。**ブリッジと序列が違うので番号で指す。
	t.Run("bids a level and a denomination", func(t *testing.T) {
		m := newMock()
		c := controller.NewVintCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 3 4"))
		m.AssertCalled(t, "Bid", 3, domain.VintDenomNoTrump)
		assert.Equal(t, mockOutput, c.Exec("bid 1 0"))
		m.AssertCalled(t, "Bid", 1, domain.VintDenomSpade)
		assert.Equal(t, mockOutput, c.Exec("ps"))
		m.AssertCalled(t, "PassBid")
	})

	t.Run("bid rejects a bad level or denomination", func(t *testing.T) {
		c := controller.NewVintCuiController(newMock())
		assert.True(t, msgRejected(c.Exec("b")))
		assert.True(t, msgRejected(c.Exec("b 3")))
		assert.Contains(t, c.Exec("b abc 1"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 0 1"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 8 1"), "Invalid bid level")
		assert.Contains(t, c.Exec("b 3 5"), "Invalid denomination")
		assert.Contains(t, c.Exec("b 3 -1"), "Invalid denomination")
	})

	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewVintCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 12"))
		m.AssertCalled(t, "PlayCard", 12)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 13"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewVintCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
