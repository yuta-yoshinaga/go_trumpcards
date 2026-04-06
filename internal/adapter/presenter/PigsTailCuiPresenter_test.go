package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestPigsTailForPresenter() *domain.PigsTail {
	players := []*domain.PigsTailPlayer{
		domain.NewPigsTailPlayer(true),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
	}
	pt := domain.NewPigsTail(domain.NewTrumpCards(0), players)
	pt.Reset()
	return pt
}

func TestPigsTailCuiPresenter_Output(t *testing.T) {
	p := &presenter.PigsTailCuiPresenter{}

	t.Run("initial state", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		output := p.Output(pt, nil)
		assert.Contains(t, output, "Pig's Tail (ぶたのしっぽ)")
		assert.Contains(t, output, "山札: 52枚")
		assert.Contains(t, output, "手番:")
	})
	t.Run("with error", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		output := p.Output(pt, errors.New("test error"))
		assert.Contains(t, output, "test error")
	})
	t.Run("game ended", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		// Play until end
		for !pt.GetGameEndFlag() {
			if pt.IsHumanTurn() {
				_ = pt.PlayerAction(0)
			} else {
				_ = pt.CpuAction()
			}
		}
		output := p.Output(pt, nil)
		assert.Contains(t, output, "ゲーム終了！")
		assert.Contains(t, output, "の負け！")
	})
	t.Run("with cpu actions", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		pt.SetCpuActions([]*domain.PigsTailCpuAction{
			{DrawPlayerIdx: 1, PenaltyFlag: false},
			{DrawPlayerIdx: 2, PenaltyFlag: true, PenaltyCount: 5},
		})
		output := p.Output(pt, nil)
		assert.Contains(t, output, "CPUの行動")
		assert.Contains(t, output, "セーフ")
		assert.Contains(t, output, "ペナルティ")
	})
}

func TestPigsTailCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.PigsTailCuiPresenter{}
	pt := newTestPigsTailForPresenter()
	output := p.ActionLogOutput(pt)
	assert.NotEmpty(t, output)
}
