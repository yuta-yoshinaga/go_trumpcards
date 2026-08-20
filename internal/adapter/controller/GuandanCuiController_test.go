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

func TestGuandanCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":1}`
	checkOutput := "combo: bomb"

	newMock := func() *mockUsecases.MockGuandanInteractor {
		m := new(mockUsecases.MockGuandanInteractor)
		m.On("GetConfig").Return(domain.DefaultGuandanConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("PlayCards", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("ReturnTribute", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Check", mock.Anything).Return(checkOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewGuandanCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGuandanConfig())
	})

	// **役は複数枚で出す。**単一の添字では足りない。
	t.Run("plays a combination", func(t *testing.T) {
		m := newMock()
		c := controller.NewGuandanCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "PlayCards", []int{3})
		assert.Equal(t, mockOutput, c.Exec("play 0 1 2"))
		m.AssertCalled(t, "PlayCards", []int{0, 1, 2})
	})

	t.Run("play rejects bad input", func(t *testing.T) {
		c := controller.NewGuandanCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgStem("cardIndexesRequiredTriple"))
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p -1"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 27"), msgInvalidCardIndexPrefix())
		// **同じ札を 2 回数えられない。**通すとペアが 1 枚から作れてしまう。
		assert.Contains(t, c.Exec("p 1 1"), "twice")
	})

	t.Run("pass, tribute and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewGuandanCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ps"))
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertNumberOfCalls(t, "Pass", 2)
		assert.Equal(t, mockOutput, c.Exec("t 4"))
		m.AssertCalled(t, "ReturnTribute", 4)
		assert.Equal(t, mockOutput, c.Exec("tribute 0"))
		m.AssertCalled(t, "ReturnTribute", 0)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("next"))
		m.AssertNumberOfCalls(t, "NextHand", 2)
	})

	t.Run("tribute rejects bad input", func(t *testing.T) {
		c := controller.NewGuandanCuiController(newMock())
		assert.Contains(t, c.Exec("t"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("t abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("t 27"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewGuandanCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}

// **出さずに役だけ調べられる** (#5734)。手札は動かない。
func TestGuandanCuiController_Check(t *testing.T) {
	mockOutput := `{"phase":0}`
	checkOutput := "combo: bomb"
	newMock := func() *mockUsecases.MockGuandanInteractor {
		m := new(mockUsecases.MockGuandanInteractor)
		m.On("GetConfig").Return(domain.DefaultGuandanConfig())
		m.On("PlayCards", mock.Anything).Return(mockOutput)
		m.On("Check", mock.Anything).Return(checkOutput)
		return m
	}

	m := newMock()
	c := controller.NewGuandanCuiController(m)
	assert.Equal(t, checkOutput, c.Exec("ch 0 1 2"))
	m.AssertCalled(t, "Check", []int{0, 1, 2})
	assert.Equal(t, checkOutput, c.Exec("check 3"))
	m.AssertCalled(t, "Check", []int{3})
	// **調べても場には出ない。**
	m.AssertNotCalled(t, "PlayCards", mock.Anything)

	// 数字でない引数は読み飛ばしたうえで警告が付く。
	warned := c.Exec("ch 0 x")
	assert.Contains(t, warned, checkOutput)
	assert.NotEqual(t, checkOutput, warned)
	m.AssertCalled(t, "Check", []int{0})
}
