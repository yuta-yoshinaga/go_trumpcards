//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newFiftyOneCuiController() (*controller.FiftyOneCuiController, *mockusecase.MockFiftyOneInteractor) {
	fiMock := new(mockusecase.MockFiftyOneInteractor)
	fiMock.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	fiMock.On("Reset", mock.Anything).Return("reset output")
	fiMock.On("ExchangeOne", 0, 1).Return("exchange one output")
	fiMock.On("ExchangeAll").Return("exchange all output")
	fiMock.On("Stop").Return("stop output")
	fiMock.On("ActionLog").Return("action log output")
	return controller.NewFiftyOneCuiController(fiMock), fiMock
}

func TestFiftyOneCuiController_Reset(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("reset")
	assert.Equal(t, "reset output", result)
}

func TestFiftyOneCuiController_Play(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("p 0 1")
	assert.Equal(t, "exchange one output", result)
}

func TestFiftyOneCuiController_Play_MissingArgs(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("p")
	assert.Contains(t, result, "Usage")
}

func TestFiftyOneCuiController_Play_InvalidHandIdx(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("p abc 0")
	assert.Contains(t, result, msgStem("invalidHandIndexRaw"))
}

func TestFiftyOneCuiController_Play_InvalidTableIdx(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("p 0 abc")
	assert.Contains(t, result, msgStem("invalidTableIndexRaw"))
}

func TestFiftyOneCuiController_All(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("a")
	assert.Equal(t, "exchange all output", result)
}

func TestFiftyOneCuiController_Stop(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("stop")
	assert.Equal(t, "stop output", result)
}

func TestFiftyOneCuiController_ActionLog(t *testing.T) {
	ctrl, _ := newFiftyOneCuiController()
	result := ctrl.Exec("log")
	assert.Equal(t, "action log output", result)
}

func TestFiftyOneCuiController_SetDifficulty(t *testing.T) {
	fiMock := new(mockusecase.MockFiftyOneInteractor)
	fiMock.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	fiMock.On("Reset", mock.MatchedBy(func(cfg domain.FiftyOneConfig) bool {
		return cfg.CpuDifficulty == domain.FiftyOneDifficultyHard
	})).Return("hard output")

	ctrl := controller.NewFiftyOneCuiController(fiMock)
	result := ctrl.Exec("sd 2")
	assert.Equal(t, "hard output", result)
}
