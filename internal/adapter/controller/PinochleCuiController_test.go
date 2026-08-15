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

func newPinochleMock() *mockUsecases.MockPinochleInteractor {
	m := new(mockUsecases.MockPinochleInteractor)
	m.On("GetConfig").Return(domain.DefaultPinochleConfig())
	m.On("ResetWithConfig", mock.Anything).Return(`{"phase":0}`)
	m.On("Bid", mock.Anything).Return(`{"phase":0}`)
	m.On("Pass").Return(`{"phase":0}`)
	m.On("CallTrump", mock.Anything).Return(`{"phase":0}`)
	m.On("ConfirmMelds").Return(`{"phase":0}`)
	m.On("Play", mock.Anything).Return(`{"phase":0}`)
	m.On("NextTrick").Return(`{"phase":0}`)
	m.On("NextRound").Return(`{"phase":0}`)
	m.On("Hint").Return(`{"phase":0}`)
	m.On("ActionLog").Return(`[]`)
	return m
}

func TestPinochleCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultPinochleConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
	})

	// bid
	t.Run("bid command b with amount", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("b 25")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 25)
	})

	t.Run("bid command bid with amount", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("bid 30")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 30)
	})

	t.Run("bid command no args", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("b")
		assert.Contains(t, result, "Bid amount is required")
	})

	t.Run("bid command invalid arg", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("b abc")
		assert.Contains(t, result, "Invalid bid amount")
	})

	t.Run("bid command below min", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("b 10")
		assert.Contains(t, result, "Invalid bid amount: 10")
	})

	// pass
	t.Run("pass command pa", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("pa")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Pass")
	})

	t.Run("pass command pass", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("pass")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Pass")
	})

	// trump
	t.Run("trump command t with suit", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("t 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 3)
	})

	t.Run("trump command trump with suit", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("trump 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 1)
	})

	t.Run("trump command no args", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("t")
		assert.Contains(t, result, "Suit is required")
	})

	t.Run("trump command invalid arg", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("t abc")
		assert.Contains(t, result, "Invalid suit")
	})

	t.Run("trump command out of range", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("t 5")
		assert.Contains(t, result, "Invalid suit: 5")
	})

	// meld
	t.Run("meld command m", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("m")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ConfirmMelds")
	})

	t.Run("meld command meld", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("meld")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ConfirmMelds")
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command no args", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command invalid arg", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultPinochleConfig()
		expected.CpuDifficulty = domain.PinochleCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultPinochleConfig()
		expected.CpuDifficulty = domain.PinochleCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty out of range", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sd 3")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("sl 2000")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultPinochleConfig()
		expected.PointLimit = 2000
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("setlimit 500")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultPinochleConfig()
		expected.PointLimit = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sl")
		assert.Contains(t, result, "required")
	})

	t.Run("setlimit invalid", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("sl 0")
		assert.Contains(t, result, "Invalid point limit: 0")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// log
	t.Run("log command l", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, `[]`, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("log command log", func(t *testing.T) {
		m := newPinochleMock()
		c := controller.NewPinochleCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, `[]`, result)
		m.AssertCalled(t, "ActionLog")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewPinochleCuiController(newPinochleMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
