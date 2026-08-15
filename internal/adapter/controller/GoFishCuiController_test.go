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

func newGoFishCuiController() (*controller.GoFishCuiController, *mockusecase.MockGoFishInteractor) {
	giMock := new(mockusecase.MockGoFishInteractor)
	giMock.On("GetConfig").Return(domain.DefaultGoFishConfig())
	giMock.On("Reset", mock.Anything).Return("reset output")
	giMock.On("Ask", 1, 3).Return("ask output")
	giMock.On("ActionLog").Return("action log output")
	return controller.NewGoFishCuiController(giMock), giMock
}

func TestGoFishCuiController_Reset(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("reset")
	assert.Equal(t, "reset output", result)
}

func TestGoFishCuiController_Ask(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("ask 1 3")
	assert.Equal(t, "ask output", result)
}

func TestGoFishCuiController_Ask_MissingArgs(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("ask")
	assert.Contains(t, result, "Usage")
}

func TestGoFishCuiController_Ask_InvalidTarget(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("ask abc 3")
	assert.Contains(t, result, msgStem("invalidTargetIndexRaw"))
}

func TestGoFishCuiController_Ask_InvalidRank(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("ask 1 abc")
	assert.Contains(t, result, msgStem("invalidRankRaw"))
}

func TestGoFishCuiController_SetDifficulty(t *testing.T) {
	giMock := new(mockusecase.MockGoFishInteractor)
	giMock.On("GetConfig").Return(domain.DefaultGoFishConfig())
	giMock.On("Reset", mock.MatchedBy(func(cfg domain.GoFishConfig) bool {
		return cfg.CpuDifficulty == domain.GoFishCpuDifficultyHard
	})).Return("hard output")
	ctrl := controller.NewGoFishCuiController(giMock)

	result := ctrl.Exec("sd 2")
	assert.Equal(t, "hard output", result)
}

func TestGoFishCuiController_ActionLog(t *testing.T) {
	ctrl, _ := newGoFishCuiController()
	result := ctrl.Exec("log")
	assert.Equal(t, "action log output", result)
}
