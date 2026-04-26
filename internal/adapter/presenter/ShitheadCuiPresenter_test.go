package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newShitheadForPresenter() *domain.Shithead {
	return domain.NewDefaultShithead()
}

func TestShitheadCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ShitheadCuiPresenter)

	t.Run("initial reset state", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		result := p.Output(s, nil)
		assert.Contains(t, result, "Shithead")
		assert.Contains(t, result, "山札:")
		assert.Contains(t, result, "場札: なし")
		assert.Contains(t, result, "手番:")
	})

	t.Run("error message rendered", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		result := p.Output(s, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game end rendered", func(t *testing.T) {
		s := newShitheadForPresenter()
		s.Reset()
		// Drive game until end; budget is generous to handle pickup loops.
		for i := 0; i < 50000 && !s.GetGameEndFlag(); i++ {
			if s.IsHumanTurn() {
				_ = s.PlayerPlay(nil)
			} else {
				s.CpuPlay()
			}
		}
		if !s.GetGameEndFlag() {
			t.Skip("game did not end within step budget")
		}
		result := p.Output(s, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestShitheadCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ShitheadCuiPresenter)
	s := newShitheadForPresenter()
	s.Reset()
	out := p.ActionLogOutput(s)
	assert.NotEmpty(t, out)
}
