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

func costlyGame() *domain.CostlyColours {
	c := domain.NewDefaultCostlyColours()
	c.Reset()
	return c
}

func costlyCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

func TestCostlyColoursCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CostlyColoursCuiPresenter)

	t.Run("shows the deal, the turn-up and the count", func(t *testing.T) {
		out := p.Output(costlyGame(), nil)
		assert.Contains(t, out, "コストリー・カラーズ")
		assert.NotContains(t, out, "costlycolours.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.deal"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.count"), "{{", 2)[0])
	})

	// **表の 1 枚は札そのものを見せる。** ショーの色役も J / 2 の 4 点も
	// これ次第なので、「トランプ:」の見出しだけでは足りない。
	t.Run("shows the turn-up card itself", func(t *testing.T) {
		c := costlyGame()
		c.SetTurnUpForTest(costlyCard(domain.CardDesignDiamond, 13))
		out := p.Output(c, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.turnUp"), "{{", 2)[0])
		assert.Contains(t, out, "♦K", "表の 1 枚が出ていない")
		assert.NotContains(t, out, i18n.T("costlycolours.noTurnUp"))
	})

	// **開幕は交換の案内。** ここを飛ばすと mog の無いゲームに見える。
	t.Run("asks about the mog first", func(t *testing.T) {
		out := p.Output(costlyGame(), nil)
		assert.Contains(t, out, i18n.T("costlycolours.mogPrompt"))
		assert.Contains(t, out, i18n.T("costlycolours.mogHint"))
	})

	// **J と 2 は持っているだけで点になる。** 印が無いと気軽に切ってしまう。
	t.Run("marks the jacks and deuces in the hand", func(t *testing.T) {
		c := costlyGame()
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignSpade, 11))
		c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignHeart, 2))
		c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignClover, 7))
		out := p.Output(c, nil)
		assert.Equal(t, 2, strings.Count(out, i18n.T("costlycolours.pegMark")),
			"J と 2 のちょうど 2 枚に印が付いていない")
	})

	t.Run("lists the playable cards on the human's turn", func(t *testing.T) {
		c := costlyGame()
		require.NoError(t, c.PlayerMog(false))
		c.SetCurrentForTest(0)
		out := p.Output(c, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.playable"), "{{", 2)[0])
	})

	// **31 を超えずに出せる札が無いなら「ゴー」。** 探させない。
	t.Run("says go when nothing fits under 31", func(t *testing.T) {
		c := costlyGame()
		require.NoError(t, c.PlayerMog(false))
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignSpade, 10))
		c.SetTotalForTest(30)
		c.SetCurrentForTest(0)
		require.Empty(t, c.PlayableIdxs(0))
		assert.Contains(t, p.Output(c, nil), i18n.T("costlycolours.mustGo"))
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(costlyGame(), assert.AnError), assert.AnError.Error())
	})

	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(costlyGame(), nil)
		assert.NotContains(t, out, "costlycolours.")
		assert.NotContains(t, out, "コストリー", "日本語が漏れている")
	})
}

// **ショーは 3 項目 + 名指しの色役。** 点だけだと梯子のどこに乗ったのか
// 分からない。
func TestCostlyColoursCuiPresenter_Show(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")

	c := costlyGame()
	c.SetPhaseForTest(domain.CostlyColoursPhasePlay)
	c.SetTurnUpForTest(costlyCard(domain.CardDesignSpade, 12))
	for i := 0; i < domain.CostlyColoursPlayerCnt; i++ {
		c.GetPlayer(i).ResetDeal()
	}
	for _, v := range []int{3, 5, 9} {
		c.GetPlayer(0).AddPlayed(costlyCard(domain.CardDesignSpade, v))
	}
	c.GetPlayer(1).AddPlayed(costlyCard(domain.CardDesignHeart, 4))
	c.GetPlayer(1).AddPlayed(costlyCard(domain.CardDesignClover, 6))
	c.GetPlayer(1).AddPlayed(costlyCard(domain.CardDesignDiamond, 8))
	c.FinishDealForTest()

	out := new(presenter.CostlyColoursCuiPresenter).Output(c, nil)
	assert.NotContains(t, out, "costlycolours.", "生キーが出ている")
	assert.Contains(t, out, i18n.T("costlycolours.showTitle"))
	for _, key := range []string{"jackDeuce", "rank", "colour"} {
		assert.Contains(t, out, i18n.T("costlycolours.score."+key), "得点項目 %s が出ていない", key)
	}
	// 名指しの色役。
	assert.Contains(t, out, i18n.T("costlycolours.combo."+domain.CostlyComboCostlyColours),
		"4 枚同スートの役名が出ていない")
}

func TestCostlyColoursCuiPresenter_Hint(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.CostlyColoursCuiPresenter)

	// 交換フェーズでは札を指さず、応じるかどうかだけを言う。
	out := p.HintOutput(costlyGame())
	assert.NotContains(t, out, "costlycolours.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.hintReason"), "{{", 2)[0])
	assert.NotContains(t, out, strings.SplitN(i18n.T("costlycolours.hintCard"), "{{", 2)[0],
		"交換フェーズなのに札を指している")

	// 数え上げでは札を指す。
	c := costlyGame()
	require.NoError(t, c.PlayerMog(false))
	c.SetCurrentForTest(0)
	out = p.HintOutput(c)
	assert.Contains(t, out, strings.SplitN(i18n.T("costlycolours.hintCard"), "{{", 2)[0])
}

func TestCostlyColoursWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.CostlyColoursWebPresenter)
	c := costlyGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))

	assert.Equal(t, domain.CostlyColoursPhaseMog, res["phase"])
	assert.Equal(t, float64(1), res["dealNumber"])
	assert.Len(t, res["players"], domain.CostlyColoursPlayerCnt)
	assert.Equal(t, float64(0), res["total"])
	// **表の 1 枚は渡す。** ショーの色役も J / 2 の 4 点もこれ次第。
	require.NotNil(t, res["turnUp"], "表の 1 枚が渡っていない")

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.Len(t, human["cards"], domain.CostlyColoursHandSize)
	// **相手の手札は伏せる。**
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
	assert.Equal(t, float64(domain.CostlyColoursHandSize), cpu["cardCount"])
	// 交換フェーズでは出せる札を渡さない。
	assert.Empty(t, res["playableIdxs"])
}

// **出せる札はサーバが数える。** 31 を超える札を並べると押しても弾かれる。
func TestCostlyColoursWebPresenter_CarriesThePlayableIdxs(t *testing.T) {
	p := new(presenter.CostlyColoursWebPresenter)
	c := costlyGame()
	require.NoError(t, c.PlayerMog(false))
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignSpade, 10))
	c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignHeart, 2))
	c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignClover, 9))
	c.SetTotalForTest(29) // 29 + 2 = 31 だけが通る。
	c.SetCurrentForTest(0)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Equal(t, []any{float64(1)}, res["playableIdxs"], "31 を超える札まで渡っている")
	assert.Equal(t, float64(29), res["total"])
}

// **手番でなければ出せる札は渡さない。** 渡すと、相手の番に自分の札が
// 押せる状態で並ぶ ── ドメインはフェーズしか見ないので、席の判定はここの仕事。
func TestCostlyColoursWebPresenter_NoPlayableIdxsOnTheCpuTurn(t *testing.T) {
	p := new(presenter.CostlyColoursWebPresenter)
	c := costlyGame()
	require.NoError(t, c.PlayerMog(false))
	c.SetCurrentForTest(1) // CPU の番。フェーズはまだ play。
	require.Equal(t, domain.CostlyColoursPhasePlay, c.GetPhase())
	require.NotEmpty(t, c.PlayableIdxs(0), "ドメイン側は席 0 の札を返すはず")

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Empty(t, res["playableIdxs"], "相手の番なのに出せる札が渡っている")
	assert.Equal(t, false, res["isHumanTurn"])
}

func TestCostlyColoursWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.CostlyColoursWebPresenter)
	var res map[string]any

	require.NoError(t, json.Unmarshal([]byte(p.Output(costlyGame(), nil)), &res))
	assert.Equal(t, "costlycolours.mogPhase", res["messageCode"])

	// **出せる札が無いなら「ゴー」と伝える。**
	c := costlyGame()
	require.NoError(t, c.PlayerMog(false))
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(costlyCard(domain.CardDesignSpade, 10))
	c.SetTotalForTest(30)
	c.SetCurrentForTest(0)
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Equal(t, "costlycolours.playPhase.go", res["messageCode"])

	// 出せるならふつうの案内 (負のコントロール)。
	c.SetTotalForTest(0)
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, nil)), &res))
	assert.Equal(t, "costlycolours.playPhase", res["messageCode"])

	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(costlyGame())), &res))
	assert.Contains(t, []any{"costlycolours.hintRequested", "costlycolours.noHint"}, res["messageCode"])
}

func TestCostlyColoursWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.CostlyColoursWebPresenter)
	c := costlyGame()
	err := c.PlayerPlay(0) // 交換フェーズなので弾かれる。
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(c, err)), &res))
	assert.Equal(t, "costlycolours.errNotPlayPhase", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestCostlyColoursPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	c := costlyGame()
	require.NoError(t, c.PlayerMog(false))

	cui := new(presenter.CostlyColoursCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(c))

	web := new(presenter.CostlyColoursWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(c)), &res))
	assert.Contains(t, res, "entries")
}
