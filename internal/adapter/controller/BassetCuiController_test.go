//go:build test
// +build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestBassetCuiController_ParoliCommand(t *testing.T) {
	m := new(uc.MockBassetInteractor)
	m.On("PressParoli").Return("ok")
	c := controller.NewBassetCuiController(m)
	assert.Equal(t, "ok", c.Exec("paroli"))
	m.AssertCalled(t, "PressParoli")
}
func TestBassetCuiController_BetCommand(t *testing.T) {
	m := new(uc.MockBassetInteractor)
	m.On("PlaceBet", 7, 10).Return("ok")
	c := controller.NewBassetCuiController(m)
	assert.Equal(t, "ok", c.Exec("bet 7 10"))
	m.AssertCalled(t, "PlaceBet", 7, 10)
}
