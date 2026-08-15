//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDurakCuiController_Exec(t *testing.T) {
	mockOutput := "test output"
	diMock := new(usecase.MockDurakInteractor)
	diMock.On("Reset").Return(mockOutput)
	diMock.On("Attack", mock.Anything).Return(mockOutput)
	diMock.On("Defend", mock.Anything, mock.Anything).Return(mockOutput)
	diMock.On("Pass").Return(mockOutput)
	diMock.On("TakeCards").Return(mockOutput)
	diMock.On("Sort", mock.Anything).Return(mockOutput)
	diMock.On("GetConfig").Return(domain.DefaultDurakConfig())
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("ActionLog").Return("log output")

	c := controller.NewDurakCuiController(diMock)

	t.Run("reset", func(t *testing.T) {
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("attack", func(t *testing.T) {
		result := c.Exec("a 0")
		assert.Equal(t, mockOutput, result)
	})

	// **打ち間違いを 0 番として実行しない。** `a zz` が `a 0` に化けていたとき、
	// 40 配り中 4 配りで実際に札が出ていた (issue #5390)。
	t.Run("attack refuses an unparseable index", func(t *testing.T) {
		refuseMock := new(usecase.MockDurakInteractor)
		refuseMock.On("Attack", mock.Anything).Return(mockOutput)
		rc := controller.NewDurakCuiController(refuseMock)

		result := rc.Exec("a zz")

		assert.Equal(t, msgInvalidCardIndex("zz"), result)
		refuseMock.AssertNotCalled(t, "Attack", mock.Anything)
	})

	t.Run("defend", func(t *testing.T) {
		result := c.Exec("d 0 1")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("pass", func(t *testing.T) {
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("take", func(t *testing.T) {
		result := c.Exec("t")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("sort", func(t *testing.T) {
		result := c.Exec("sort 1")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("set difficulty", func(t *testing.T) {
		result := c.Exec("sd 1")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("log", func(t *testing.T) {
		result := c.Exec("log")
		assert.Equal(t, "log output", result)
	})
}
