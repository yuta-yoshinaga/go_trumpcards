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

func TestLiteratureCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockLiteratureInteractor {
		m := new(mockUsecases.MockLiteratureInteractor)
		m.On("GetConfig").Return(domain.DefaultLiteratureConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Ask", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Claim", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultLiteratureConfig())
	})

	// **要求は 相手・スート・ランク の 3 つ。**
	t.Run("asks for a card", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("a 1 1 2"))
		m.AssertCalled(t, "Ask", 1, domain.CardDesignSpade, 2)
		assert.Equal(t, mockOutput, c.Exec("ask 5 4 13"))
		m.AssertCalled(t, "Ask", 5, domain.CardDesignDiamond, 13)
	})

	t.Run("ask rejects bad input", func(t *testing.T) {
		c := controller.NewLiteratureCuiController(newMock())
		assert.Contains(t, c.Exec("a"), msgStem("askNeedsSeatSuitRank"))
		assert.Contains(t, c.Exec("a 1 1"), msgStem("askNeedsSeatSuitRank"))
		assert.Contains(t, c.Exec("a 9 1 2"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("a abc 1 2"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("a 1 9 2"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("a 1 abc 2"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("a 1 1 14"), msgStem("invalidRank113"))
		assert.Contains(t, c.Exec("a 1 1 abc"), msgStem("invalidRank113"))
	})

	// **宣言は組と 6 席。**所在を 6 枚ぶん申告する。
	t.Run("claims a half-suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 0 0 0 2 2 4 4"))
		m.AssertCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})
		assert.Equal(t, mockOutput, c.Exec("claim 7 1 1 1 3 3 5"))
		m.AssertCalled(t, "Claim", 7, []int{1, 1, 1, 3, 3, 5})
	})

	t.Run("claim rejects bad input", func(t *testing.T) {
		c := controller.NewLiteratureCuiController(newMock())
		assert.Contains(t, c.Exec("c"), msgStem("claimNeedsHalfSuitAndHolders"))
		assert.Contains(t, c.Exec("c 0 0 0 2"), msgStem("claimNeedsHalfSuitAndHolders"))
		assert.Contains(t, c.Exec("c 9 0 0 2 2 4 4"), msgStem("invalidHalfSuit0Max"))
		assert.Contains(t, c.Exec("c abc 0 0 2 2 4 4"), msgStem("invalidHalfSuit0Max"))
		assert.Contains(t, c.Exec("c 0 0 0 2 2 4 9"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("c 0 0 0 2 2 4 abc"), msgStem("invalidSeat0Max"))
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
