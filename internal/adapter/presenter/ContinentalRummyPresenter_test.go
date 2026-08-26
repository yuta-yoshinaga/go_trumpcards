//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func contCardFor(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// contAtHumanDrawP は「人間が引く番」まで進んだ盤を返す。
func contAtHumanDrawP(t *testing.T) *domain.ContinentalRummy {
	t.Helper()
	for try := 0; try < 50; try++ {
		c := domain.NewDefaultContinentalRummy()
		c.Reset()
		if c.GetPhase() == domain.ContinentalRummyPhaseDraw && c.IsHumanTurn() {
			return c
		}
	}
	t.Fatal("50 回配っても人間の引く番にならなかった -- 前提が崩れている")
	return nil
}

// contWinningHand は上がれる 16 枚を人間に持たせた盤を返す。
func contWinningHand(t *testing.T) *domain.ContinentalRummy {
	t.Helper()
	c := contAtHumanDrawP(t)
	c.SetPhaseForTest(domain.ContinentalRummyPhaseDiscard)
	c.SetCurrentIdxForTest(domain.ContinentalRummyHumanIdx)
	p := c.GetPlayer(domain.ContinentalRummyHumanIdx)
	p.ResetRound()
	for _, run := range [][2]int{
		{domain.CardDesignSpade, 2}, {domain.CardDesignSpade, 7},
		{domain.CardDesignHeart, 4}, {domain.CardDesignClover, 9},
		{domain.CardDesignDiamond, 5},
	} {
		for k := 0; k < 3; k++ {
			p.AddCard(contCardFor(run[0], run[1]+k))
		}
	}
	p.AddCard(contCardFor(domain.CardDesignDiamond, 13))
	require.Equal(t, 16, p.GetCardsSize())
	return c
}

func TestContinentalRummyCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.ContinentalRummyCuiPresenter)

	// **認められた形はいつでも見えていること。** 15 枚をどう割るかが全部で、
	// 5+5+5 が入っていないのが肝。
	t.Run("always shows the legal layouts, and 5+5+5 is not one of them", func(t *testing.T) {
		out := p.Output(contAtHumanDrawP(t), nil)
		assert.Contains(t, out, "3+3+3+3+3")
		assert.Contains(t, out, "4+4+4+3")
		assert.Contains(t, out, "5+4+3+3")
		assert.NotContains(t, out, "5+5+5", "上がれない形を案内している")
	})

	t.Run("shows the round, the stock and the discard top", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		out := p.Output(c, nil)
		assert.Contains(t, out, i18n.Tf("continentalrummy.round",
			"n", itoa(c.GetRoundNumber()), "total", itoa(c.GetConfig().TotalRounds)))
		assert.Contains(t, out, i18n.Tf("continentalrummy.stock", "n", itoa(c.GetStockCount())))
		require.NotNil(t, c.GetDiscardTop())
		assert.Contains(t, out, i18n.T("continentalrummy.drawPrompt"))
	})

	t.Run("numbers the human hand and hides the other hands", func(t *testing.T) {
		out := p.Output(contAtHumanDrawP(t), nil)
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "[14]")
		assert.Contains(t, out, i18n.T("continentalrummy.you"))
		assert.Contains(t, out, i18n.Tf("continentalrummy.cpuName", "n", "1"))
	})

	// **上がれるときは黙っていない。** 15 枚の分割は目で追いきれない。
	t.Run("says so when the hand can go out, and stays quiet otherwise", func(t *testing.T) {
		c := contWinningHand(t)
		idx, ok := c.CanGoOut()
		require.True(t, ok)
		assert.Contains(t, p.Output(c, nil), i18n.Tf("continentalrummy.canGoOut", "idx", itoa(idx)))

		quiet := contAtHumanDrawP(t)
		quiet.SetPhaseForTest(domain.ContinentalRummyPhaseDiscard)
		if _, ok := quiet.CanGoOut(); !ok {
			assert.NotContains(t, p.Output(quiet, nil), i18n.Tf("continentalrummy.canGoOut", "idx", "0"))
		}
	})

	// **加点は内訳で見せる。** 合計だけだと、どう上がると得なのかが伝わらない。
	t.Run("the settlement lists every bonus and what was collected", func(t *testing.T) {
		c := contWinningHand(t)
		idx, _ := c.CanGoOut()
		require.NoError(t, c.GoOut(idx))
		res := c.GetLastResult()
		require.NotNil(t, res)
		require.NotEmpty(t, res.Bonuses)

		out := p.Output(c, nil)
		assert.Contains(t, out, i18n.T("continentalrummy.you"))
		for _, b := range res.Bonuses {
			assert.Contains(t, out, i18n.T("continentalrummy.bonus."+b.Key),
				"加点 %q が出ていない", b.Key)
		}
		assert.Contains(t, out, i18n.Tf("continentalrummy.collected",
			"per", itoa(res.PerOpponent), "total", itoa(res.Total)))
		assert.NotContains(t, out, "bonus.", "識別子が生のまま出ている")
	})

	t.Run("an error resolves to real text", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		err := c.Discard(0) // 引く前には捨てられない。
		require.Error(t, err)
		out := p.Output(c, err)
		assert.Contains(t, out, i18n.T("continentalrummy.errWrongPhase"))
		assert.NotContains(t, out, "continentalrummy.errWrongPhase")
	})

	t.Run("english renders english", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(contAtHumanDrawP(t), nil)
		assert.Contains(t, out, i18n.T("continentalrummy.you"))
		assert.NotContains(t, out, "上がれる形", "英語表示に日本語が漏れている")
	})
}

func TestContinentalRummyCuiPresenter_HintAndLog(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.ContinentalRummyCuiPresenter)

	t.Run("hint resolves to real text on the draw and the discard", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		hint := c.GetHint()
		require.NotNil(t, hint)
		out := p.HintOutput(c)
		assert.Contains(t, out, i18n.T("continentalrummy.reason."+hint.Reason))
		assert.NotContains(t, out, "reason.", "識別子が生のまま出ている")

		require.NoError(t, c.DrawStock())
		hint2 := c.GetHint()
		require.NotNil(t, hint2)
		out2 := p.HintOutput(c)
		assert.Contains(t, out2, i18n.T("continentalrummy.reason."+hint2.Reason))
		assert.Contains(t, out2, itoa(hint2.DiscardIdx))
	})

	t.Run("no hint on the opponents' turn", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		c.SetCurrentIdxForTest(1)
		assert.Contains(t, p.HintOutput(c), i18n.T("continentalrummy.noHint"))
	})

	// **棋譜は終局してから。**
	t.Run("the transcript is empty until the game ends", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		assert.Contains(t, p.ActionLogOutput(c), i18n.T("cuiActionLogEmpty"))
	})
}

func TestContinentalRummyWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.ContinentalRummyWebPresenter)

	decode := func(t *testing.T, s string) *controller.ContinentalRummyWebOutput {
		t.Helper()
		var out controller.ContinentalRummyWebOutput
		require.NoError(t, json.Unmarshal([]byte(s), &out))
		return &out
	}

	// **相手の手札は枚数だけ。** レスポンスを覗いても中身が見えないこと。
	t.Run("carries only the human hand, and counts for the rest", func(t *testing.T) {
		out := decode(t, p.Output(contAtHumanDrawP(t), nil))
		require.Len(t, out.Players, domain.ContinentalRummyPlayerCnt)
		assert.True(t, out.Players[0].IsHuman)
		assert.Len(t, out.Players[0].Cards, domain.ContinentalRummyHandSize)
		for _, pl := range out.Players[1:] {
			assert.Empty(t, pl.Cards, "相手の手札が漏れている")
			assert.Equal(t, domain.ContinentalRummyHandSize, pl.CardCount)
		}
	})

	// **認められた形は毎回ドメインから返す。** ページ側に書き写さない。
	t.Run("carries the legal layouts, without 5+5+5", func(t *testing.T) {
		out := decode(t, p.Output(contAtHumanDrawP(t), nil))
		assert.Equal(t, domain.ContinentalRummyLayouts(), out.Layouts)
		for _, l := range out.Layouts {
			assert.NotEqual(t, []int{5, 5, 5}, l)
		}
	})

	t.Run("carries the round, the stock, the discard top and the phase", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		out := decode(t, p.Output(c, nil))
		assert.Equal(t, c.GetRoundNumber(), out.RoundNumber)
		assert.Equal(t, c.GetStockCount(), out.StockCount)
		assert.Equal(t, c.GetConfig().TotalRounds, out.TotalRounds)
		require.NotNil(t, out.DiscardTop)
		assert.Equal(t, domain.ContinentalRummyPhaseDraw, out.Phase)
		assert.True(t, out.IsHumanTurn)
		assert.Equal(t, "continentalrummy.drawPhase", out.MessageCode)
	})

	// **上がれるかはサーバが解く。** ページ側で 15 枚の分割を解き直さない。
	t.Run("carries the go-out index when the hand can go out", func(t *testing.T) {
		c := contWinningHand(t)
		out := decode(t, p.Output(c, nil))
		idx, ok := c.CanGoOut()
		require.True(t, ok)
		assert.Equal(t, idx, out.GoOutIdx)
		assert.Equal(t, "continentalrummy.discardPhase.canGoOut", out.MessageCode)

		// 負のコントロール: 上がれないときは -1。
		plain := contAtHumanDrawP(t)
		plain.SetPhaseForTest(domain.ContinentalRummyPhaseDiscard)
		if _, ok := plain.CanGoOut(); !ok {
			assert.Equal(t, -1, decode(t, p.Output(plain, nil)).GoOutIdx)
		}
	})

	t.Run("the settlement carries the bonuses and what was collected", func(t *testing.T) {
		c := contWinningHand(t)
		idx, _ := c.CanGoOut()
		require.NoError(t, c.GoOut(idx))
		out := decode(t, p.Output(c, nil))
		require.NotNil(t, out.LastResult)
		assert.Equal(t, domain.ContinentalRummyHumanIdx, out.LastResult.WinnerIdx)
		assert.NotEmpty(t, out.LastResult.Bonuses)
		assert.Equal(t, out.LastResult.PerOpponent*(domain.ContinentalRummyPlayerCnt-1),
			out.LastResult.Total)
		// 並べたシーケンスは公開情報。
		assert.Len(t, out.Players[0].Melds, 5)
	})

	t.Run("an error carries its message code", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		err := c.Discard(0)
		require.Error(t, err)
		out := decode(t, p.Output(c, err))
		assert.Equal(t, "continentalrummy.errWrongPhase", out.MessageCode)
		assert.NotEmpty(t, out.Message)
	})

	t.Run("hint and log", func(t *testing.T) {
		c := contAtHumanDrawP(t)
		hint := decode(t, p.HintOutput(c))
		assert.Equal(t, "continentalrummy.hintRequested", hint.MessageCode)
		assert.NotEmpty(t, hint.HintReason)

		c.SetCurrentIdxForTest(1)
		assert.Equal(t, "continentalrummy.noHint", decode(t, p.HintOutput(c)).MessageCode)
	})
}
