//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newSpoonsCuiMock() *usecase.MockSpoonsInteractor {
	m := new(usecase.MockSpoonsInteractor)
	m.On("Reset").Return("reset-ok")
	m.On("Pass", mock.Anything).Return("pass-ok")
	m.On("Grab").Return("grab-ok")
	m.On("NextRound").Return("next-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultSpoonsConfig())
	return m
}

func TestSpoonsCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
	})
	t.Run("reset", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.Equal(t, "reset-ok", c.Exec("r"))
		assert.Equal(t, "reset-ok", c.Exec("reset"))
	})
	t.Run("pass with index", func(t *testing.T) {
		m := newSpoonsCuiMock()
		c := controller.NewSpoonsCuiController(m)
		assert.Equal(t, "pass-ok", c.Exec("p 2"))
		assert.Equal(t, "pass-ok", c.Exec("pass 0"))
		m.AssertCalled(t, "Pass", 2)
	})
	// 引数なしの 0 番既定は残す一方、打ち間違いは 0 番として実行しない (issue #5390)。
	t.Run("pass refuses an unparseable index", func(t *testing.T) {
		m := newSpoonsCuiMock()
		c := controller.NewSpoonsCuiController(m)

		assert.Equal(t, msgInvalidCardIndex("zz"), c.Exec("p zz"))

		m.AssertNotCalled(t, "Pass", mock.Anything)
	})
	t.Run("pass without index defaults to 0", func(t *testing.T) {
		m := newSpoonsCuiMock()
		c := controller.NewSpoonsCuiController(m)
		assert.Equal(t, "pass-ok", c.Exec("p"))
		m.AssertCalled(t, "Pass", 0)
	})
	t.Run("grab", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.Equal(t, "grab-ok", c.Exec("g"))
		assert.Equal(t, "grab-ok", c.Exec("grab"))
	})
	t.Run("next", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.Equal(t, "next-ok", c.Exec("n"))
		assert.Equal(t, "next-ok", c.Exec("next"))
	})
	t.Run("log", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.Equal(t, "log-ok", c.Exec("log"))
		assert.Equal(t, "log-ok", c.Exec("l"))
	})
	t.Run("empty and unknown", func(t *testing.T) {
		c := controller.NewSpoonsCuiController(newSpoonsCuiMock())
		assert.NotEmpty(t, c.Exec(""))
		assert.NotEmpty(t, c.Exec("xyz"))
	})
}
