//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestSpoonsCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.SpoonsCuiPresenter)

	// i18n resolves to the ja translations in this build, so assert on the
	// rendered Japanese text that uniquely identifies each phase.
	t.Run("initial pass state", func(t *testing.T) {
		g := setupSpoonsTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Spoons")
		assert.Contains(t, result, "ラウンド")
		assert.Contains(t, result, "あなたの番です")
	})

	t.Run("error", func(t *testing.T) {
		g := setupSpoonsTest()
		assert.Contains(t, p.Output(g, errors.New("oops")), "oops")
	})

	t.Run("grab prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseGrab, "gw": true})
		assert.Contains(t, p.Output(g, nil), "g (grab)")
	})

	t.Run("round end with loser", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseRoundEnd, "rl": 2})
		out := p.Output(g, nil)
		assert.Contains(t, out, "n (next)")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ci": 1})
		assert.Contains(t, p.Output(g, nil), "CPU")
	})

	t.Run("human win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 0, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "あなたの勝ちです")
	})

	t.Run("cpu win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 1, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "の勝ちです")
	})

	t.Run("eliminated and has-spoon shown", func(t *testing.T) {
		g := setupSpoonsTest()
		g.GetPlayer(3).SetEliminated(true)
		g.GetPlayer(1).SetHasSpoon(true)
		assert.Contains(t, p.Output(g, nil), "脱落")
	})
}

func TestSpoonsCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoonsCuiPresenter)
	g := setupSpoonsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **同ランクの組が最大の判断材料。**Web は色付きリングで示すのに、CUI は素の
// 一覧しか出さず、パスする札を暗算させていた (#4889)。
func TestSpoonsCuiPresenter_MarksRankGroups(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // 3 枚組の強調まで見る
	defer color.SetNoColor(orig)
	p := new(presenter.SpoonsCuiPresenter)

	g := setupSpoonsTest()
	human := g.GetPlayer(0)
	setHand := func(vals []int) {
		human.Reset()
		designs := []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond}
		for i, v := range vals {
			human.AddCard(domain.NewCard(designs[i%len(designs)], v, false))
		}
	}

	// バラバラなら印は付かない。
	setHand([]int{2, 5, 9, 13})
	assert.NotContains(t, p.Output(g, nil), "(同")

	// 2 枚組には枚数が付くが、強調はしない。
	setHand([]int{7, 7, 9, 13})
	two := p.Output(g, nil)
	assert.Contains(t, two, "(同2枚)")
	assert.NotContains(t, two, color.Yellow("(同2枚)"))

	// **3 枚は 1 枚違い。**強調してフォーカードが近いことを示す。
	setHand([]int{7, 7, 7, 13})
	three := p.Output(g, nil)
	assert.Contains(t, three, color.Yellow("(同3枚)"))
	// 組に入らない札には付かない。
	assert.NotContains(t, three, "(同1枚)")
}

// #5669: 配り手 (feeder) は山札から取っては隣へ渡し続ける役で、他プレイヤーの
// 手札が増えるペースを読む材料になる。Web は badge.feeder を出しているのに、
// CUI は GetFeederIdx を一度も呼んでおらず、誰が配り手か知る手段が無かった。
func TestSpoonsCuiPresenter_MarksTheFeeder(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SpoonsCuiPresenter)

	t.Run("tags the feeder and nobody else", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"fi": 1})
		require.Equal(t, 1, g.GetFeederIdx())

		out := p.Output(g, nil)

		lines := strings.Split(out, "\n")
		var feederLines, taggedLines int
		for _, line := range lines {
			if strings.Contains(line, i18n.T("spoons.feederTag")) {
				taggedLines++
				if strings.Contains(line, "CPU 1") {
					feederLines++
				}
			}
		}
		assert.Equal(t, 1, taggedLines, "タグが付く行はちょうど1つ")
		assert.Equal(t, 1, feederLines, "その行は配り手 (CPU 1) の行")
	})

	// 人間が配り手のこともある。
	t.Run("tags the human when they feed", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"fi": 0})

		out := p.Output(g, nil)

		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, i18n.T("spoons.feederTag")) {
				assert.Contains(t, line, i18n.T("cuiPlayerYou"))
				return
			}
		}
		t.Fatal("配り手のタグが出ていない")
	})

	// 既存の letters 表示と両立する。
	t.Run("keeps the letters readout", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"fi": 0})

		out := p.Output(g, nil)

		// 文字がまだ 1 つも付いていない状態は "-" で描かれる (spoonsLetters)。
		assert.Contains(t, out, i18n.Tf("spoons.lettersLabel", "letters", "-"))
	})
}
