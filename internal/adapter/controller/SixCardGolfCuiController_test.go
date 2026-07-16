//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestSixCardGolfCuiController_Hint(t *testing.T) {
	m := new(mockUsecases.MockSixCardGolfInteractor)
	m.On("Hint").Return("hint result")
	c := controller.NewSixCardGolfCuiController(m)

	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	m.AssertCalled(t, "Hint")
}
