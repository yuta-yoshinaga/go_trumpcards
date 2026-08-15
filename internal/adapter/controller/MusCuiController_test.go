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

func TestMusCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockMusInteractor {
		m := new(mockUsecases.MockMusInteractor)
		m.On("GetConfig").Return(domain.DefaultMusConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Mus", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Bet", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewMusCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewMusCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMusConfig())
	})

	t.Run("mus", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m"))
		m.AssertCalled(t, "Mus", true)
	})

	t.Run("mus alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("mus"))
		m.AssertCalled(t, "Mus", true)
	})

	t.Run("cut / corte", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c"))
		m.AssertCalled(t, "Mus", false)
	})

	t.Run("corte alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("corte"))
		m.AssertCalled(t, "Mus", false)
	})

	t.Run("discard with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 0 2"))
		m.AssertCalled(t, "Discard", []int{0, 2})
	})

	t.Run("discard no args (keep all)", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d"))
		m.AssertCalled(t, "Discard", []int{})
	})

	t.Run("discard invalid index", func(t *testing.T) {
		result := controller.NewMusCuiController(newMock()).Exec("d abc")
		assert.Contains(t, result, "Invalid")
	})

	t.Run("paso", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("paso"))
		m.AssertCalled(t, "Bet", domain.MusActionPaso, 0)
	})

	t.Run("envido with amount", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("e 4"))
		m.AssertCalled(t, "Bet", domain.MusActionEnvido, 4)
	})

	t.Run("envido alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("envido 3"))
		m.AssertCalled(t, "Bet", domain.MusActionEnvido, 3)
	})

	t.Run("envido no args", func(t *testing.T) {
		result := controller.NewMusCuiController(newMock()).Exec("e")
		assert.Contains(t, result, msgKey("amountRequiredEGE2"))
	})

	t.Run("ordago", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ordago"))
		m.AssertCalled(t, "Bet", domain.MusActionOrdago, 0)
	})

	t.Run("quiero", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("quiero"))
		m.AssertCalled(t, "Bet", domain.MusActionQuiero, 0)
	})

	t.Run("noquiero", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nq"))
		m.AssertCalled(t, "Bet", domain.MusActionNoQuiero, 0)
	})

	t.Run("noquiero alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("noquiero"))
		m.AssertCalled(t, "Bet", domain.MusActionNoQuiero, 0)
	})

	t.Run("next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("next alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("next"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultMusConfig()
		expected.CpuDifficulty = domain.MusCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewMusCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewMusCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewMusCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
