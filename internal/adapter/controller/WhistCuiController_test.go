//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newWhistCuiMock() *mockUsecases.MockWhistInteractor {
	m := new(mockUsecases.MockWhistInteractor)
	m.On("GetConfig").Return(domain.DefaultWhistConfig())
	m.On("ResetWithConfig", mock.Anything).Return("output")
	m.On("Play", mock.Anything).Return("output")
	m.On("NextTrick").Return("output")
	m.On("NextRound").Return("output")
	m.On("Hint").Return("output")
	m.On("ActionLog").Return("output")
	return m
}

func TestWhistCuiController_Exec(t *testing.T) {
	t.Run("setlimit", func(t *testing.T) {
		m := newWhistCuiMock()
		c := controller.NewWhistCuiController(m)
		assert.Equal(t, "output", c.Exec("sl 100"))
		expected := domain.DefaultWhistConfig()
		expected.PointLimit = 100
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
}

// TestWhistCuiController_SetLimitShowsTheNewLimit は受け入れ条件2を端から端まで見る。
//
// 上の setlimit テストは ResetWithConfig に新しい値が渡ることまでしか言えない。
// 「変更した直後の出力に反映される」は本物の interactor と presenter を通さないと
// 確かめられないので、ここだけモックを使わない。目標点の行は配りに依存しない。
func TestWhistCuiController_SetLimitShowsTheNewLimit(t *testing.T) {
	oldLang := i18n.Lang()
	i18n.SetLang("ja")
	defer i18n.SetLang(oldLang)

	game := domain.NewDefaultWhist()
	interactor := usecase.NewWhistInteractor(game, new(presenter.WhistCuiPresenter))
	c := controller.NewWhistCuiController(interactor)

	before := c.Exec("r")
	assert.NotContains(t, before, "目標点: 100", "前提: 既定の目標点は 100 ではない")

	after := c.Exec("sl 100")
	assert.Contains(t, after, "目標点: 100")
}
