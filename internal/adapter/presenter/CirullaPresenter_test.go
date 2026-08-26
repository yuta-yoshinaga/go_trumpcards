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

// cirullaGame は配り終えた卓を返す (席 0 が人間で先に打つ)。
func cirullaGame() *domain.Cirulla {
	c := domain.NewDefaultCirulla()
	c.Reset()
	return c
}

func TestCirullaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CirullaCuiPresenter)

	t.Run("shows the round, the stock and the table", func(t *testing.T) {
		out := p.Output(cirullaGame(), nil)
		// **訳が引けていることまで見る。**
		assert.Contains(t, out, "チルッラ")
		assert.NotContains(t, out, "cirulla.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("cirulla.round"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("cirulla.deck"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("cirulla.table"), "{{", 2)[0])
	})

	// **場札には番号が要る。** 取る札はこの番号で指すので、無いと組合せ捕獲が
	// 打てない。
	t.Run("numbers the table cards", func(t *testing.T) {
		out := p.Output(cirullaGame(), nil)
		for _, n := range []string{"[0]", "[1]", "[2]", "[3]"} {
			assert.Contains(t, out, n)
		}
	})

	// **絵札は A/J/Q/K で出す。** 捕獲の合計は J=8・Q=9・K=10 で計算するので、
	// 11・12・13 と出ると盤面の足し算が読めない。
	t.Run("prints court cards by their face label", func(t *testing.T) {
		c := cirullaGame()
		// **配りに絵札を期待しない。** 絵札の出ない配りだと、数値表示のまま
		// でもこの検査は素通りする。
		c.SetTableForTest([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 11, true),
			domain.NewCard(domain.CardDesignClover, 12, true),
			domain.NewCard(domain.CardDesignHeart, 13, true),
			domain.NewCard(domain.CardDesignDiamond, 1, true),
		})
		out := p.Output(c, nil)
		assert.NotRegexp(t, `[♠♣♥♦](11|12|13)\b`, out, "絵札が数値で出ている")
		for _, want := range []string{"♠J", "♣Q", "♥K", "♦A"} {
			assert.Contains(t, out, want)
		}
	})

	// **何が取れるかを出す。** 3 つの規則が混ざるので、出さないと端末から
	// 総当たりで探すことになる。
	t.Run("lists what each card can take", func(t *testing.T) {
		c := cirullaGame()
		// **手札だけ固定しても盤面は固定できない。** 場は Reset の配りのままなので、
		// 3 で取れる札が無い配りだとこの検査は落ちる ── 場も固定する。
		c.SetTableForTest([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 3, true),
			domain.NewCard(domain.CardDesignClover, 9, true),
		})
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, true))
		require.NotEmpty(t, c.GetCaptureOptions(0, 0), "固定した盤面で取れない")
		out := p.Output(c, nil)
		assert.Contains(t, out,
			strings.SplitN(i18n.T("cirulla.captureOption"), "{{", 2)[0])
	})

	t.Run("says when nothing can be taken", func(t *testing.T) {
		c := cirullaGame()
		c.GetPlayer(0).Reset()
		// 場に無い値かつ合計にもならない札を持たせる。
		c.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, true)})
		c.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, true))
		out := p.Output(c, nil)
		assert.Contains(t, out, i18n.T("cirulla.noCapture"))
	})

	// **配札ボーナスは出た瞬間に見せる。** 見せないと、点が 3 や 10 動いた
	// 理由が盤面のどこにも書かれない。
	//
	// ボーナスは配りで決まるので、出た配りを引くまで配り直す ── if で包むと、
	// 出ない配りでは何も検査しないまま緑になる。
	t.Run("shows a deal bonus when one landed", func(t *testing.T) {
		for try := 0; try < 200; try++ {
			c := cirullaGame()
			if c.GetLastBonus()[0] == "" {
				continue
			}
			out := p.Output(c, nil)
			assert.Contains(t, out, strings.SplitN(i18n.T("cirulla.bonusLine"), "{{", 2)[0])
			// **識別子ではなく訳文が出る。** キーは "cirulla.bonus."+識別子 と
			// 組み立てているので、どのガードも綴りを見ていない ── 引けなければ
			// プレイヤーは `cirulla.bonus.barsega` を読むことになる。
			assert.NotContains(t, out, "cirulla.bonus.")
			assert.Contains(t, out, i18n.T("cirulla.bonus."+c.GetLastBonus()[0]))
			return
		}
		t.Fatal("200 回配ってもボーナスが出ない — 判定が死んでいる")
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(cirullaGame(), assert.AnError), assert.AnError.Error())
	})

	// **英語も訳が引ける。**
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(cirullaGame(), nil)
		assert.NotContains(t, out, "cirulla.")
		assert.NotContains(t, out, "チルッラ", "日本語が漏れている")
	})
}

func TestCirullaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CirullaCuiPresenter)

	out := p.HintOutput(cirullaGame())
	assert.NotContains(t, out, "cirulla.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("cirulla.hintCard"), "{{", 2)[0])
}

func TestCirullaWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.CirullaWebPresenter)
	c := cirullaGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))

	assert.Equal(t, domain.CirullaPhasePlay, res["phase"])
	assert.Equal(t, float64(1), res["roundNumber"])
	assert.Len(t, res["players"], domain.CirullaPlayerCnt)
	assert.Len(t, res["table"], domain.CirullaTableSize)
	assert.Equal(t, float64(c.GetDeckRemaining()), res["deckRemaining"])
	assert.NotNil(t, res["captureOptions"])
	assert.NotNil(t, res["hintCaptureIdxs"])

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.Len(t, human["cards"], domain.CirullaHandSize)
	// **相手の手札は伏せる。**
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
	assert.Equal(t, float64(domain.CirullaHandSize), cpu["cardCount"])
}

// **取れる組はサーバが数えて渡す。** 3 つの規則が絡むので、クライアントに
// 解かせると必ずずれる。
func TestCirullaWebPresenter_CarriesTheCaptureOptions(t *testing.T) {
	p := new(presenter.CirullaWebPresenter)
	c := cirullaGame()
	// 場をアッソ抜きにし、手札に A を持たせる ── 総取りが 1 つ出るはず。
	c.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, true),
		domain.NewCard(domain.CardDesignClover, 5, true),
	})
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, true))

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	opts := res["captureOptions"].([]any)
	require.Len(t, opts, 1)
	groups := opts[0].([]any)
	require.Len(t, groups, 1, "アッソの総取りが 1 つだけのはず")
	assert.Len(t, groups[0], 2, "場の 2 枚をまとめて取る形になっていない")
}

func TestCirullaWebPresenter_NoOptionsOffTurn(t *testing.T) {
	p := new(presenter.CirullaWebPresenter)
	c := cirullaGame()
	cirullaRunToRoundEnd(t, c)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Empty(t, res["captureOptions"], "手番でないのに候補が出ている")
	require.NotNil(t, res["lastResult"])
	last := res["lastResult"].(map[string]any)
	assert.Len(t, last["lines"], 8)
	assert.Len(t, last["totals"], domain.CirullaPlayerCnt)
}

func TestCirullaWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.CirullaWebPresenter)
	c := cirullaGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Contains(t, []any{"cirulla.playPhase", "cirulla.playPhase.canCapture"}, res["messageCode"])

	cirullaRunToRoundEnd(t, c)
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Contains(t, []any{"cirulla.roundEnd", "cirulla.result.humanWin", "cirulla.result.cpuWin"},
		res["messageCode"])

	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(cirullaGame())), &res))
	assert.Contains(t, []any{"cirulla.hintRequested", "cirulla.noHint"}, res["messageCode"])
}

func TestCirullaWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.CirullaWebPresenter)
	c := cirullaGame()
	err := c.PlayerPlay(99, nil)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, err)), &res))
	assert.Equal(t, "cirulla.errCardRange", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestCirullaPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	c := cirullaGame()
	cirullaRunToRoundEnd(t, c)

	cui := new(presenter.CirullaCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(c))

	web := new(presenter.CirullaWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(c)), &res))
	assert.Contains(t, res, "entries")
}

// cirullaRunToRoundEnd はラウンドを最後まで打つ。
func cirullaRunToRoundEnd(t *testing.T, c *domain.Cirulla) {
	t.Helper()
	for step := 0; step < 500 && c.GetPhase() == domain.CirullaPhasePlay; step++ {
		if c.IsHumanTurn() {
			hint := c.GetHint()
			require.GreaterOrEqual(t, hint.HandIdx, 0)
			require.NoError(t, c.PlayerPlay(hint.HandIdx, hint.CaptureIdxs))
			continue
		}
		c.CpuPlay()
	}
	require.NotEqual(t, domain.CirullaPhasePlay, c.GetPhase(), "ラウンドが終わらない")
}
