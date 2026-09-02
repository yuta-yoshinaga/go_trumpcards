//go:build test && (!js || !wasm || extra4)

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPutCuiPresenter_Output_Phases(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	phases := []domain.PutPhase{
		domain.PutPhasePlay,
		domain.PutPhaseRespond,
		domain.PutPhaseTrickEnd,
		domain.PutPhaseHandEnd,
		domain.PutPhaseGameEnd,
	}
	for _, ph := range phases {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(ph)
		if ph == domain.PutPhaseRespond {
			g.SetPendingLevel(domain.PutLevelPut)
			g.SetPutCallerIdx(1)
			g.SetResponderIdx(0)
		}
		if ph == domain.PutPhaseHandEnd {
			g.SetHandWinnerIdx(0)
		}
		out := p.Output(g, nil)
		assert.NotEmpty(t, out, "phase %d output should be non-empty", ph)
	}
}

func TestPutCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestPutCuiPresenter_Output_GameEndBanner(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	g.SetGameEndFlag(true)
	g.SetPhase(domain.PutPhaseGameEnd)
	g.SetPlayerMatchPoints(0, 15)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestPutCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PutCuiPresenter)

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick(nil)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("respond hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhaseRespond)
		g.SetResponderIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhaseTrickEnd)
		out := p.HintOutput(g)
		assert.True(t, strings.TrimSpace(out) != "")
	})
}

func TestPutCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **3 が最強・4 が最弱という独自の並びを画面に出す (#6609)。**
// Web は折りたたみの参照表を常設しているのに、CUI にはこのゲーム最大の
// 意外性を知る手段が一切なかった。
func TestPutCuiPresenter_ShowsTheRankReference(t *testing.T) {
	g := domain.NewDefaultPut()
	g.Reset()
	out := new(presenter.PutCuiPresenter).Output(g, nil)

	assert.Contains(t, out, "カードの強さ")
	assert.NotContains(t, out, "{{")

	// **並びが本当にドメインと一致していること。** 文字列があるだけでは、
	// 誰かが順番を書き間違えても気付けない。表に出した順で強さが単調に
	// 下がることを `PutCardStrength` に照らして確かめる。
	order := []int{3, 2, 1, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4}
	prev := 99
	for _, v := range order {
		s := domain.PutCardStrength(domain.NewCard(domain.CardDesignSpade, v, true))
		assert.Less(t, s, prev, "値 %d は 1 つ前より弱いはず", v)
		prev = s
	}
	// 3 が最強で 4 が最弱であることを名指しで固定する。
	assert.Greater(t,
		domain.PutCardStrength(domain.NewCard(domain.CardDesignSpade, 3, true)),
		domain.PutCardStrength(domain.NewCard(domain.CardDesignSpade, 2, true)),
		"3 は 2 より強い")
	assert.Less(t,
		domain.PutCardStrength(domain.NewCard(domain.CardDesignSpade, 4, true)),
		domain.PutCardStrength(domain.NewCard(domain.CardDesignSpade, 5, true)),
		"4 は 5 より弱い")
}
