package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
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
	// **Web は引いた札と判定を見せている。**CUI は CPU の行動履歴しか出さず、
	// 自分が引いた札もペナルティかどうかも分からなかった (#4864)。
	t.Run("last drawn card", func(t *testing.T) {
		orig := color.NoColor()
		color.SetNoColor(true)
		defer color.SetNoColor(orig)

		pt := newTestPigsTailForPresenter()
		// 誰も引いていないうちは出さない。
		assert.NotContains(t, p.Output(pt, nil), "直前に引いた札")

		for !pt.IsHumanTurn() {
			assert.NoError(t, pt.CpuAction())
		}
		assert.NoError(t, pt.PlayerAction(0))
		assert.NotNil(t, pt.GetLastDrawCard())
		if pt.GetLastPenalty() {
			output := p.Output(pt, nil)
			assert.Contains(t, output, "→ ペナルティ！")
			assert.Contains(t, output, "枚引き取り")
		} else {
			assert.Contains(t, p.Output(pt, nil), "→ セーフ")
		}

		// ペナルティ側の文言も踏む。引いた札は保ったままフラグだけ立てる。
		pt.SetLastPenalty(true)

		output := p.Output(pt, nil)
		assert.Contains(t, output, "→ ペナルティ！")
		assert.Contains(t, output, "枚引き取り")
		assert.NotContains(t, output, "→ セーフ")
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
	t.Run("human penalty includes the number of collected cards", func(t *testing.T) {
		orig := color.NoColor()
		color.SetNoColor(true)
		defer color.SetNoColor(orig)

		for _, tc := range []struct {
			count    int
			expected string
		}{
			{count: 2, expected: "直前に引いた札: SPADE 1 → ペナルティ！ 2枚引き取り"},
			{count: 6, expected: "直前に引いた札: SPADE 1 → ペナルティ！ 6枚引き取り"},
		} {
			pt := newTestPigsTailForPresenter()
			pt.SetLastDrawCard(domain.NewCard(domain.CardDesignSpade, 1, false))
			pt.SetLastPenalty(true)
			pt.SetHumanAction(&domain.PigsTailCpuAction{PenaltyFlag: true, PenaltyCount: tc.count})
			output := p.Output(pt, nil)
			assert.Contains(t, output, tc.expected)
			assert.NotContains(t, output, "{{")
		}
	})
	t.Run("human penalty without an action keeps the legacy message", func(t *testing.T) {
		orig := color.NoColor()
		color.SetNoColor(true)
		defer color.SetNoColor(orig)

		pt := newTestPigsTailForPresenter()
		pt.SetLastDrawCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		pt.SetLastPenalty(true)
		output := p.Output(pt, nil)
		assert.Contains(t, output, "直前に引いた札: SPADE 1 → ペナルティ！ 場札を全て引き取り")
		assert.NotContains(t, output, "{{")
		assert.NotContains(t, output, "枚引き取り")
	})
}

func TestPigsTailCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.PigsTailCuiPresenter{}
	pt := newTestPigsTailForPresenter()
	output := p.ActionLogOutput(pt)
	assert.NotEmpty(t, output)
}
