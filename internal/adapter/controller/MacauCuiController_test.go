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

func TestMacauCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockMacauInteractor {
		m := new(mockUsecases.MockMacauInteractor)
		m.On("GetConfig").Return(domain.DefaultMacauConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("ChooseSuit", mock.Anything).Return(mockOutput)
		m.On("Draw").Return(mockOutput)
		m.On("Declare").Return(mockOutput)
		m.On("SkipDeclare").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(false)
		m.On("IsHumanDeclareTurn").Return(false)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewMacauCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMacauConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("hint command", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewMacauCuiController(newMock()).Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid arg", func(t *testing.T) {
		assert.Contains(t, controller.NewMacauCuiController(newMock()).Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("playing an 8 inlines a suit prompt", func(t *testing.T) {
		m := new(mockUsecases.MockMacauInteractor)
		m.On("Play", 0).Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(true)
		c := controller.NewMacauCuiController(m)
		result := c.Exec("p 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "s {0}", tmpl)
	})

	t.Run("reaching one card inlines a declare prompt", func(t *testing.T) {
		m := new(mockUsecases.MockMacauInteractor)
		m.On("Play", 1).Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(false)
		m.On("IsHumanDeclareTurn").Return(true)
		c := controller.NewMacauCuiController(m)
		result := c.Exec("p 1")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "dc", tmpl)
	})

	t.Run("playing a normal card returns play result unchanged", func(t *testing.T) {
		m := new(mockUsecases.MockMacauInteractor)
		m.On("Play", 2).Return(mockOutput)
		m.On("IsHumanChooseSuitTurn").Return(false)
		m.On("IsHumanDeclareTurn").Return(false)
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
	})

	t.Run("draw", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d"))
		m.AssertCalled(t, "Draw")
	})

	t.Run("suit with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("s 1"))
		m.AssertCalled(t, "ChooseSuit", 1)
	})

	t.Run("suit below range", func(t *testing.T) {
		assert.Contains(t, controller.NewMacauCuiController(newMock()).Exec("s 0"), msgKey("invalidSuitRange", "val", "0"))
	})

	t.Run("declare dc", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dc"))
		m.AssertCalled(t, "Declare")
	})

	t.Run("declare long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("declare"))
		m.AssertCalled(t, "Declare")
	})

	t.Run("skipdeclare sk", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sk"))
		m.AssertCalled(t, "SkipDeclare")
	})

	t.Run("skipdeclare long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("skipdeclare"))
		m.AssertCalled(t, "SkipDeclare")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultMacauConfig()
		expected.CpuDifficulty = domain.MacauCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty out of range", func(t *testing.T) {
		assert.Equal(t, msgInvalidCpuDifficulty("3"), controller.NewMacauCuiController(newMock()).Exec("sd 3"))
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 300"))
		expected := domain.DefaultMacauConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit zero", func(t *testing.T) {
		assert.Equal(t, msgKey("invalidPointLimit11000", "val", "0"), controller.NewMacauCuiController(newMock()).Exec("sl 0"))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewMacauCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewMacauCuiController(newMock()).Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		assert.Contains(t, controller.NewMacauCuiController(newMock()).Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
