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

func TestContractRummyCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockContractRummyInteractor {
		m := new(mockUsecases.MockContractRummyInteractor)
		m.On("GetConfig").Return(domain.DefaultContractRummyConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("MeldContract", mock.Anything).Return(mockOutput)
		m.On("MeldExtra", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewContractRummyCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultContractRummyConfig())
	})

	t.Run("drawstock aliases", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			m := newMock()
			c := controller.NewContractRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromStock")
		}
	})

	t.Run("drawdiscard aliases", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			m := newMock()
			c := controller.NewContractRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromDiscard")
		}
	})

	t.Run("meldcontract parses slot args", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("mc 0,1,2 3,4,5"))
		m.AssertCalled(t, "MeldContract", [][]int{{0, 1, 2}, {3, 4, 5}})
	})

	t.Run("meldcontract usage when no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		out := c.Exec("mc")
		assert.Contains(t, out, msgUsage("usageMcABCDEFGHIOneSlotPerArg"))
		m.AssertNotCalled(t, "MeldContract", mock.Anything)
	})

	t.Run("meldcontract usage when bad arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		out := c.Exec("mc 0,X,2")
		assert.Contains(t, out, msgUsage("usageMcABCDEFGHIOneSlotPerArg"))
		m.AssertNotCalled(t, "MeldContract", mock.Anything)
	})

	t.Run("meldextra", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("me 0,1,2"))
		m.AssertCalled(t, "MeldExtra", []int{0, 1, 2})
	})

	t.Run("layoff", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 1 0 2"))
		m.AssertCalled(t, "Layoff", 1, 0, 2)
	})

	t.Run("layoff usage when missing args", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		out := c.Exec("lo 1")
		assert.Contains(t, out, msgUsage("usageLoTargetplayeridxMeldidxCardindex"))
		m.AssertNotCalled(t, "Layoff", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard usage when missing arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		out := c.Exec("d")
		assert.NotEmpty(t, out)
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})

	t.Run("nextround aliases", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			m := newMock()
			c := controller.NewContractRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "NextRound")
		}
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ContractRummyConfig) bool {
			return cfg.CpuDifficulty == domain.ContractRummyCpuDifficultyHard
		}))
	})

	t.Run("setpenalty", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 50"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ContractRummyConfig) bool {
			return cfg.FailContractPenalty == 50
		}))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewContractRummyCuiController(m)
		out := c.Exec("blarghhh")
		assert.NotEmpty(t, out)
	})
}
