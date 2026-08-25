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

func cometGame() *domain.Comet {
	c := domain.NewDefaultComet()
	c.Reset()
	return c
}

func cometCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

func TestCometCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CometCuiPresenter)

	t.Run("shows the round, the dead hand and the sequence", func(t *testing.T) {
		out := p.Output(cometGame(), nil)
		assert.Contains(t, out, "コメット")
		assert.NotContains(t, out, "comet.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("comet.round"), "{{", 2)[0])
		// **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。
		assert.Contains(t, out, strings.SplitN(i18n.T("comet.dead"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("comet.pile"), "{{", 2)[0])
	})

	// **先頭は何でも出せると書く。** 「次に要るランク」と別の文言でないと、
	// 何を出してよいのか分からない。
	t.Run("says the lead is free, then names the rank needed", func(t *testing.T) {
		c := cometGame()
		assert.Contains(t, p.Output(c, nil), i18n.T("comet.needAny"))

		c.SetNeedForTest(7)
		out := p.Output(c, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("comet.need"), "{{", 2)[0])
		assert.NotContains(t, out, i18n.T("comet.needAny"), "先頭の案内が残っている")
	})

	// **コメットは目立たせる。** どのランクの代わりにもなる 1 枚がただの ♦9 に
	// 見えていると、出せる手を見落とす。
	t.Run("marks the comet in the hand", func(t *testing.T) {
		c := cometGame()
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(cometCard(domain.CardDesignDiamond, 9))
		c.GetPlayer(0).AddCard(cometCard(domain.CardDesignClover, 9))
		out := p.Output(c, nil)
		assert.Contains(t, out, i18n.T("comet.wildMark"))
		// ♦ でない 9 には印が付かない (負のコントロール)。
		assert.Equal(t, 1, strings.Count(out, i18n.T("comet.wildMark")),
			"♦ 以外の 9 にも印が付いている")
	})

	// **出せる札が無いならパスしかない、と書く。** 探させない。
	t.Run("says to pass when nothing is playable", func(t *testing.T) {
		c := cometGame()
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(cometCard(domain.CardDesignSpade, 2))
		c.SetNeedForTest(7)
		c.SetCurrentForTest(0)
		require.Empty(t, c.PlayableIdxs(0))
		out := p.Output(c, nil)
		assert.Contains(t, out, i18n.T("comet.mustPass"))
	})

	t.Run("lists the playable cards on the human's turn", func(t *testing.T) {
		c := cometGame()
		c.SetNeedForTest(0)
		c.SetCurrentForTest(0)
		out := p.Output(c, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("comet.playable"), "{{", 2)[0])
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(cometGame(), assert.AnError), assert.AnError.Error())
	})

	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(cometGame(), nil)
		assert.NotContains(t, out, "comet.")
		assert.NotContains(t, out, "コメット", "日本語が漏れている")
	})
}

func TestCometCuiPresenter_RoundResultAndHint(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CometCuiPresenter)

	c := cometGame()
	cometRunToRoundEnd(t, c)
	out := p.Output(c, nil)
	assert.NotContains(t, out, "comet.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("comet.goOut"), "{{", 2)[0])
	// **出なかった K は取り分の一部なので、内訳に出す。**
	assert.Contains(t, out, strings.SplitN(i18n.T("comet.unplayedKings"), "{{", 2)[0])

	hintOut := p.HintOutput(cometGame())
	assert.NotContains(t, hintOut, "comet.", "生キーが出ている")
	assert.Contains(t, hintOut, strings.SplitN(i18n.T("comet.hintCard"), "{{", 2)[0])
}

// **パスしかないときは「勧める手が無い」ではなく「パスしろ」と言う。**
func TestCometCuiPresenter_HintSaysPass(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	c := cometGame()
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignSpade, 2))
	c.SetNeedForTest(7)
	c.SetCurrentForTest(0)
	out := new(presenter.CometCuiPresenter).HintOutput(c)
	assert.Contains(t, out, i18n.T("comet.mustPass"))
	assert.NotContains(t, out, i18n.T("comet.noHint"))
}

func TestCometWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.CometWebPresenter)
	c := cometGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))

	assert.Equal(t, domain.CometPhasePlay, res["phase"])
	assert.Equal(t, float64(1), res["roundNumber"])
	assert.Len(t, res["players"], domain.CometDefaultPlayers)
	assert.Equal(t, float64(0), res["need"], "先頭なのに要るランクが決まっている")
	assert.Equal(t, float64(c.GetDeadCount()), res["deadCount"])
	assert.NotNil(t, res["playableIdxs"])
	assert.NotEmpty(t, res["playableIdxs"], "先頭なら全部出せる")

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.NotEmpty(t, human["cards"])
	// **相手の手札は伏せる。**
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
	assert.Positive(t, cpu["cardCount"])
}

// **出せる札はサーバが数える。** コメットがどのランクの代わりにもなるので、
// 画面側で組み直すと必ずずれる。
func TestCometWebPresenter_CarriesThePlayableIdxs(t *testing.T) {
	p := new(presenter.CometWebPresenter)
	c := cometGame()
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignSpade, 2))
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignDiamond, 9)) // コメット
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignHeart, 7))
	c.SetNeedForTest(7)
	c.SetCurrentForTest(0)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	idxs := res["playableIdxs"].([]any)
	require.Len(t, idxs, 2, "コメットと ♥7 の 2 枚が出せるはず")
	assert.Equal(t, []any{float64(1), float64(2)}, idxs)
}

// **手番でなければ出せる札は渡さない。** 渡すと、他人の番に自分の札が
// 押せる状態で並ぶ ── ドメインはフェーズしか見ないので、席の判定はここの仕事。
func TestCometWebPresenter_NoPlayableIdxsOnAnotherSeatsTurn(t *testing.T) {
	p := new(presenter.CometWebPresenter)
	c := cometGame()
	c.SetNeedForTest(0)
	c.SetCurrentForTest(1) // CPU の番。フェーズはまだ play。
	require.Equal(t, domain.CometPhasePlay, c.GetPhase())
	require.NotEmpty(t, c.PlayableIdxs(0), "ドメイン側は席 0 の札を返すはず")

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Empty(t, res["playableIdxs"], "他人の番なのに出せる札が渡っている")
	assert.Equal(t, false, res["isHumanTurn"])
}

func TestCometWebPresenter_NoPlayableIdxsOffTurn(t *testing.T) {
	p := new(presenter.CometWebPresenter)
	c := cometGame()
	cometRunToRoundEnd(t, c)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Empty(t, res["playableIdxs"], "手番でないのに候補が出ている")
	require.NotNil(t, res["lastResult"])
	last := res["lastResult"].(map[string]any)
	assert.Len(t, last["cardsLeft"], domain.CometDefaultPlayers)
	assert.Len(t, last["gained"], domain.CometDefaultPlayers)
	assert.NotNil(t, last["unplayedKings"])
}

func TestCometWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.CometWebPresenter)
	var res map[string]any

	// 先頭は「好きな札を出せる」。
	require.NoError(t, json.Unmarshal([]byte(p.Output(cometGame(), nil)), &res))
	assert.Equal(t, "comet.playPhase.lead", res["messageCode"])

	// **出せる札が無いならパスしかない、と伝える。**
	c := cometGame()
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignSpade, 2))
	c.SetNeedForTest(7)
	c.SetCurrentForTest(0)
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Equal(t, "comet.playPhase.mustPass", res["messageCode"])

	// 出せるならふつうの案内 (負のコントロール)。
	c.GetPlayer(0).AddCard(cometCard(domain.CardDesignHeart, 7))
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Equal(t, "comet.playPhase", res["messageCode"])

	done := cometGame()
	cometRunToRoundEnd(t, done)
	require.NoError(t, json.Unmarshal([]byte(p.Output(done, nil)), &res))
	assert.Contains(t, []any{"comet.roundEnd", "comet.result.humanWin", "comet.result.cpuWin"},
		res["messageCode"])

	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(cometGame())), &res))
	assert.Contains(t, []any{"comet.hintRequested", "comet.noHint"}, res["messageCode"])
}

func TestCometWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.CometWebPresenter)
	c := cometGame()
	err := c.PlayerPlay(99)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, err)), &res))
	assert.Equal(t, "comet.errCardRange", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestCometPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	c := cometGame()
	cometRunToRoundEnd(t, c)

	cui := new(presenter.CometCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(c))

	web := new(presenter.CometWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(c)), &res))
	assert.Contains(t, res, "entries")
}

// cometRunToRoundEnd は局を最後まで打つ。
func cometRunToRoundEnd(t *testing.T, c *domain.Comet) {
	t.Helper()
	for step := 0; step < 500 && c.GetPhase() == domain.CometPhasePlay; step++ {
		if c.IsHumanTurn() {
			h := c.GetHint()
			if h.HandIdx < 0 {
				require.NoError(t, c.PlayerPass())
				continue
			}
			require.NoError(t, c.PlayerPlay(h.HandIdx))
			continue
		}
		c.CpuPlay()
	}
	require.NotEqual(t, domain.CometPhasePlay, c.GetPhase(), "局が終わらない")
}
