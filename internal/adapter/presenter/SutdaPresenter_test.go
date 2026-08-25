//go:build test

package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// sutdaGame は配り終えた卓を返す (席 0 が人間で最初に動く)。
func sutdaGame() *domain.Sutda {
	s := domain.NewDefaultSutda()
	s.Reset()
	return s
}

// sutdaToShowdown は現在のハンドをショーダウンまで進める。
func sutdaToShowdown(t *testing.T, s *domain.Sutda) {
	t.Helper()
	for i := 0; i < 500 && s.GetPhase() == domain.SutdaPhaseBet; i++ {
		if s.IsHumanTurn() {
			require.NoError(t, s.PlayerAction(domain.SutdaActionCall))
			continue
		}
		s.CpuAction()
	}
}

func TestSutdaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.SutdaCuiPresenter)

	t.Run("shows the hand, the pot and every seat", func(t *testing.T) {
		out := p.Output(sutdaGame(), nil)
		// **訳が引けていることまで見る。** キー一致だけを見ると、ロケールが
		// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
		assert.Contains(t, out, "ソッタ")
		assert.NotContains(t, out, "sutda.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("sutda.potLine"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("sutda.playerLine"), "{{", 2)[0])
	})

	// **伏せているうちは自分のぶんだけ。** 相手の役が見えたら賭ける意味が無い。
	t.Run("shows only the human hand before the showdown", func(t *testing.T) {
		s := sutdaGame()
		out := p.Output(s, nil)
		human := s.GetHandOf(0)
		assert.Contains(t, out, i18n.T("sutda.handName."+human.Name))
		// **手札の行はちょうど 1 本。** 役名を数えると同じ役のときに数え違える
		// ので、行そのものを数える (手札行だけが 2 スペース + 数字で始まる)。
		assert.Equal(t, 1, sutdaCountHandLines(out), "自分以外の手札行が出ている")
	})

	t.Run("opens every remaining hand at the showdown", func(t *testing.T) {
		s := sutdaGame()
		sutdaToShowdown(t, s)
		require.Equal(t, domain.SutdaPhaseShowdown, s.GetPhase())
		out := p.Output(s, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("sutda.showdown"), "{{", 2)[0])
	})

	// **配り終えたポットは 0 になる。** ショーダウンで 0 と出すと、隣の
	// 「誰が何を取った」と噛み合わない。
	t.Run("shows the pot that was won at the showdown", func(t *testing.T) {
		s := sutdaGame()
		sutdaToShowdown(t, s)
		res := s.GetLastResult()
		require.NotNil(t, res)
		require.Positive(t, res.Pot)
		out := p.Output(s, nil)
		assert.NotContains(t, out, i18n.Tf("sutda.potLine", "pot", "0", "bet", "10"),
			"ショーダウンでポットが 0 と出ている")
	})

	// **光札の印を出す。** 月だけでは 13光ッタンかただの 4ット かが読めない。
	t.Run("marks the gwang cards", func(t *testing.T) {
		s := sutdaGame()
		p0 := s.GetPlayer(0)
		p0.Reset()
		p0.AddCard(domain.NewCard(1, 1, false)) // 1 月の光
		p0.AddCard(domain.NewCard(3, 1, false)) // 3 月の光
		out := p.Output(s, nil)
		assert.Contains(t, out, i18n.T("sutda.gwangMark"))
		assert.Contains(t, out, i18n.T("sutda.handName.gwang13"))
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(sutdaGame(), assert.AnError), assert.AnError.Error())
	})

	// **英語も訳が引ける。**
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(sutdaGame(), nil)
		assert.NotContains(t, out, "sutda.")
		assert.NotContains(t, out, "ソッタ", "日本語が漏れている")
	})
}

func TestSutdaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.SutdaCuiPresenter)

	out := p.HintOutput(sutdaGame())
	assert.NotContains(t, out, "sutda.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("sutda.hintAction"), "{{", 2)[0])
}

func TestSutdaWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.SutdaWebPresenter)
	s := sutdaGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, nil)), &res))

	assert.Equal(t, domain.SutdaPhaseBet, res["phase"])
	assert.Equal(t, float64(1), res["handNumber"])
	assert.Len(t, res["players"], domain.SutdaDefaultSeats)
	assert.Equal(t, float64(domain.SutdaMaxRaises), res["maxRaises"])
	assert.Equal(t, float64(domain.SutdaBetUnit), res["betUnit"])
	assert.NotEmpty(t, res["humanHandName"], "自分の役が空になっている")

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.Len(t, human["cards"], domain.SutdaHandSize)
	assert.NotEmpty(t, human["handName"])
	// **相手の札はショーダウンまで伏せる。**
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
	assert.Empty(t, cpu["handName"], "CPU の役が見えている")

	// **75 枚と同じで、20 枚も 52 枚デッキの記号では書けない。** 手続き描画の
	// 自己記述子が付いていないと画面が白札になる。
	first := human["cards"].([]any)[0].(map[string]any)
	assert.Equal(t, "hanafuda", first["deck"])
	assert.NotEmpty(t, first["glyph"])
	assert.NotEmpty(t, first["label"])
}

// 光札はラベルで見分けられる。**月だけでは役が決まらない。**
func TestSutdaWebPresenter_GwangIsVisibleInTheLabel(t *testing.T) {
	p := new(presenter.SutdaWebPresenter)
	s := sutdaGame()
	p0 := s.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(3, 1, false)) // 3 月の光
	p0.AddCard(domain.NewCard(3, 2, false)) // 3 月の並

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, nil)), &res))
	cards := res["players"].([]any)[0].(map[string]any)["cards"].([]any)
	labels := []string{cards[0].(map[string]any)["label"].(string), cards[1].(map[string]any)["label"].(string)}
	assert.NotEqual(t, labels[0], labels[1], "光と並が同じラベルになっている")
	assert.Contains(t, strings.Join(labels, ""), "光")
}

func TestSutdaWebPresenter_ShowsEveryHandAtTheShowdown(t *testing.T) {
	p := new(presenter.SutdaWebPresenter)
	s := sutdaGame()
	sutdaToShowdown(t, s)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, nil)), &res))
	assert.Equal(t, true, res["isShowdown"])
	require.NotNil(t, res["lastResult"])
	last := res["lastResult"].(map[string]any)
	assert.Positive(t, last["pot"])
	assert.NotEmpty(t, last["winners"])
	assert.Len(t, last["handNames"], domain.SutdaDefaultSeats)
	// 降りていない席は開いている。
	for _, raw := range res["players"].([]any) {
		row := raw.(map[string]any)
		if row["folded"] == true {
			continue
		}
		assert.Equal(t, true, row["revealed"], "降りていない席が伏せたまま")
		assert.NotEmpty(t, row["handName"])
	}
}

func TestSutdaWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.SutdaWebPresenter)
	s := sutdaGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, nil)), &res))
	assert.Contains(t, []any{"sutda.betPhase", "sutda.betPhase.toCall"}, res["messageCode"])

	sutdaToShowdown(t, s)
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, nil)), &res))
	assert.Contains(t, []any{"sutda.showdown.won", "sutda.showdown.lost",
		"sutda.result.humanWin", "sutda.result.cpuWin"}, res["messageCode"])

	// **ヒントは頼まれたときだけ名乗る。**
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(sutdaGame())), &res))
	assert.Contains(t, []any{"sutda.hintRequested", "sutda.noHint"}, res["messageCode"])
}

func TestSutdaWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.SutdaWebPresenter)
	s := sutdaGame()
	err := s.PlayerAction("zzz")
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(s, err)), &res))
	assert.Equal(t, "sutda.errUnknownAction", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestSutdaPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	s := sutdaGame()
	sutdaToShowdown(t, s)

	cui := new(presenter.SutdaCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(s))

	web := new(presenter.SutdaWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(s)), &res))
	assert.Contains(t, res, "entries")
}

// sutdaCountHandLines は手札の行数を数える (2 スペース + 数字で始まる行)。
func sutdaCountHandLines(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if len(line) > 2 && line[0] == ' ' && line[1] == ' ' && line[2] >= '0' && line[2] <= '9' {
			n++
		}
	}
	return n
}
