//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCrazyEightsCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCrazyEightsInteractor {
		m := new(mockUsecases.MockCrazyEightsInteractor)
		m.On("GetConfig").Return(domain.DefaultCrazyEightsConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("ChooseSuit", mock.Anything).Return(mockOutput)
		m.On("Draw").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		// Default: assume the play did NOT trigger the choose-suit phase. Tests
		// that exercise the inline suit prompt override this expectation below.
		m.On("IsHumanChooseSuitTurn").Return(false)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCrazyEightsConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCrazyEightsConfig())
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("playing an 8 inlines a suit prompt instead of forcing 's' on the next line", func(t *testing.T) {
		m := new(mockUsecases.MockCrazyEightsInteractor)
		m.On("Play", 0).Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(true)
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("p 0")
		assert.True(t, cuiutil.IsPromptRequest(result), "expected a PROMPT response, got %q", result)
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "s {0}", tmpl)
		m.AssertCalled(t, "Play", 0)
		m.AssertCalled(t, "IsHumanChooseSuitTurn")
	})

	t.Run("playing a non-8 returns the play result unchanged (no prompt)", func(t *testing.T) {
		m := new(mockUsecases.MockCrazyEightsInteractor)
		m.On("Play", 2).Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(false)
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 2)
		m.AssertCalled(t, "IsHumanChooseSuitTurn")
	})

	// draw
	t.Run("draw command d", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("d")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Draw")
	})

	t.Run("draw command draw", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("draw")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Draw")
	})

	// suit
	t.Run("suit command s with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("s 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ChooseSuit", 1)
	})

	t.Run("suit command suit with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("suit 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ChooseSuit", 4)
	})

	t.Run("suit command s no args", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("s")
		assert.Contains(t, result, "Suit is required")
	})

	t.Run("suit command s invalid arg", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("s abc")
		assert.Contains(t, result, "Invalid suit")
	})

	t.Run("suit command s below range", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("s 0")
		assert.Contains(t, result, "Invalid suit: 0")
	})

	t.Run("suit command s above range", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("s 5")
		assert.Contains(t, result, "Invalid suit: 5")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCrazyEightsConfig()
		expected.CpuDifficulty = domain.CrazyEightsCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCrazyEightsConfig()
		expected.CpuDifficulty = domain.CrazyEightsCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("sl 300")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCrazyEightsConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("setlimit 200")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCrazyEightsConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, msgPointLimitRequired())
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, msgInvalidPointLimitPrefix())
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, msgInvalidPointLimit("0"), result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, msgInvalidPointLimit("-1"), result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCrazyEightsCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewCrazyEightsCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
