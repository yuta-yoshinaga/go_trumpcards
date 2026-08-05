//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestKempsCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.KempsCuiPresenter)

	// i18n resolves to the ja translations in this build, so assert on the
	// rendered Japanese text that uniquely identifies each phase.
	t.Run("initial exchange state", func(t *testing.T) {
		g := setupKempsTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Kemps")
		assert.Contains(t, result, "ラウンド")
		assert.Contains(t, result, "フィールド")
		assert.Contains(t, result, "あなたの番です")
	})

	t.Run("error", func(t *testing.T) {
		g := setupKempsTest()
		assert.Contains(t, p.Output(g, errors.New("oops")), "oops")
	})

	t.Run("partner signal prompt", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ph": domain.KempsPhaseDeclare, "fh": 0})
		out := p.Output(g, nil)
		assert.Contains(t, out, "シグナル")
		assert.Contains(t, out, "k で Kemps")
	})

	t.Run("opponent signal prompt", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ph": domain.KempsPhaseDeclare, "fh": 1})
		assert.Contains(t, p.Output(g, nil), "気配")
	})

	t.Run("round end prompt", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ph": domain.KempsPhaseRoundEnd})
		assert.Contains(t, p.Output(g, nil), "n (next)")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ci": 1})
		assert.Contains(t, p.Output(g, nil), "CPU")
	})

	t.Run("human win", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ge": true, "wt": 0, "ph": domain.KempsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "あなたのチームの勝ちです")
	})

	t.Run("cpu win", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ge": true, "wt": 1, "ph": domain.KempsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "相手チームの勝ちです")
	})

	t.Run("signal type blink shown", func(t *testing.T) {
		g := setupKempsTest()
		g.PlayerSetSignal(int(domain.SignalBlink))
		assert.Contains(t, p.Output(g, nil), "瞬き")
	})
}

func TestKempsCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KempsCuiPresenter)
	g := setupKempsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **揃ったことに気づかないと宣言する権利ごと失う。**Web は交換中からバナーと
// ボタン強調で知らせるのに、CUI は手札を自分で見て判断するしかなかった (#4890)。
func TestKempsCuiPresenter_AnnouncesFourOfAKind(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KempsCuiPresenter)

	g := setupKempsTest()
	human := g.GetPlayer(0)
	// 1 枚だけ別ランクなら成立しない。
	human.Reset()
	for i, d := range []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond} {
		v := 7
		if i == 3 {
			v = 8
		}
		human.AddCard(domain.NewCard(d, v, false))
	}
	assert.NotContains(t, p.Output(g, nil), "4枚が同ランク")

	// 4 枚そろえば出る。
	human.Reset()
	for _, d := range []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond} {
		human.AddCard(domain.NewCard(d, 7, false))
	}
	out := p.Output(g, nil)
	assert.Contains(t, out, "4枚が同ランク")
	assert.Contains(t, out, "ケンプスを宣言できます")

	// **終局後は出さない。**宣言できないので案内する意味がなく、Web も
	// `humanHasFour && !isGameEnd` で隠している（レビュー指摘）。
	g.SetGameEndFlagForTest(true)
	assert.NotContains(t, p.Output(g, nil), "4枚が同ランク")
}
