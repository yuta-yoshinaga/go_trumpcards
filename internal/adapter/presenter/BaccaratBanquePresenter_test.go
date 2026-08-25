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

func banqueGame() *domain.BaccaratBanque {
	b := domain.NewDefaultBaccaratBanque()
	b.Reset()
	return b
}

func banqueCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// banqueStack は 右・左・親 の順に 2 周配られる決め打ちのシューを返す。
func banqueStack(vals ...int) []*domain.Card {
	out := make([]*domain.Card, 0, len(vals)+20)
	for _, v := range vals {
		out = append(out, banqueCard(domain.CardDesignSpade, v))
	}
	for i := 0; i < 20; i++ {
		out = append(out, banqueCard(domain.CardDesignHeart, 10)) // 0 点
	}
	return out
}

// banqueAtBankerDecision は「親がまだ引くか決めていない」局面を作る。
func banqueAtBankerDecision(t *testing.T) *domain.BaccaratBanque {
	t.Helper()
	b := banqueGame()
	// 右 = 3 (必ず引く)、左 = 7 (必ず止まる)、親 = 6。
	b.SetShoeForTest(banqueStack(1, 3, 3, 2, 4, 3))
	b.SetPhaseForTest(domain.BaccaratBanquePhaseResult)
	b.NextCoup()
	require.Equal(t, domain.BaccaratBanquePhaseBanker, b.GetPhase(),
		"親の判断を待つ局面になっていない -- 前提が崩れている")
	return b
}

func TestBaccaratBanqueCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.BaccaratBanqueCuiPresenter)

	t.Run("names the three seats by role, not by CPU number", func(t *testing.T) {
		out := p.Output(banqueAtBankerDecision(t), nil)
		for _, role := range []string{"role.banker", "role.right", "role.left"} {
			name := i18n.T("baccaratbanque." + role)
			assert.Contains(t, out, name)
		}
		assert.NotContains(t, out, "CPU 1", "席をロール名でなく CPU 番号で呼んでいる")
	})

	// **1 回負けても続くのがこの形式の要。** 続いていることを盤に書く。
	t.Run("shows the coup number, the bank tenure and the shoe", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		out := p.Output(b, nil)
		assert.Contains(t, out, i18n.Tf("baccaratbanque.coup", "n", "2"))
		assert.Contains(t, out, i18n.Tf("baccaratbanque.bankHeld", "n", "2"))
		assert.Contains(t, out, i18n.Tf("baccaratbanque.shoe", "n",
			itoa(b.GetShoeRemaining())))
	})

	t.Run("the banker prompt says the choice is free on any total", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		out := p.Output(b, nil)
		total := b.GetPlayer(domain.BaccaratBanqueBankerIdx).GetTotal()
		assert.Contains(t, out, i18n.Tf("baccaratbanque.bankerPrompt", "total", itoa(total)))
		assert.Contains(t, out, i18n.T("baccaratbanque.commandHint"))
	})

	// **左右は別勘定。** 片方に払い片方から取る決着が 1 行ずつ出る。
	t.Run("settlement reports both tableaux separately", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		require.NoError(t, b.BankerDraw(false))
		require.Equal(t, domain.BaccaratBanquePhaseResult, b.GetPhase())
		res := b.GetLastResult()
		require.NotNil(t, res)
		require.Len(t, res.Sides, 2)

		out := p.Output(b, nil)
		assert.Contains(t, out, i18n.Tf("baccaratbanque.resultTitle",
			"total", itoa(res.BankerTotal)))
		for _, s := range res.Sides {
			assert.Contains(t, out, i18n.T("baccaratbanque.outcome."+s.Outcome))
		}
		assert.Contains(t, out, i18n.Tf("baccaratbanque.bankerDelta",
			"delta", itoa(res.BankerDelta)))
		assert.Contains(t, out, i18n.T("baccaratbanque.nextCoupHint"))
	})

	// **降りたあとに「負けても続きます」と書かない。**
	t.Run("the bank-survives line is dropped once the bank has ended", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		require.NoError(t, b.BankerDraw(false))
		require.NoError(t, b.Retire())
		out := p.Output(b, nil)
		assert.Contains(t, out, i18n.Tf("baccaratbanque.endRetired", "n", itoa(b.GetBankHeld())))
		assert.NotContains(t, out, i18n.Tf("baccaratbanque.bankHeld", "n", itoa(b.GetBankHeld())),
			"終わったあとに「負けても続きます」が残っている")
		// 負のコントロール: 終わる前は出ている。
		assert.Contains(t, p.Output(banqueAtBankerDecision(t), nil),
			i18n.Tf("baccaratbanque.bankHeld", "n", "2"))
	})

	t.Run("an error is shown", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		err := b.Retire() // 決着の外では降りられない。
		require.Error(t, err)
		out := p.Output(b, err)
		// **実際の文言に解決していることを見る。** 生の識別子が出ていたら
		// ロケールに載せ忘れている。
		assert.Contains(t, out, i18n.T("baccaratbanque.errNotResultPhase"))
		assert.NotContains(t, out, "baccaratbanque.errNotResultPhase")
	})

	t.Run("english renders english", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(banqueAtBankerDecision(t), nil)
		assert.Contains(t, out, i18n.T("baccaratbanque.role.banker"))
		assert.NotContains(t, out, "タブロー", "英語表示に日本語が漏れている")
	})
}

func TestBaccaratBanqueCuiPresenter_HintAndLog(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.BaccaratBanqueCuiPresenter)

	t.Run("hint names an action and a reason", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		hint := b.GetHint()
		require.NotNil(t, hint)
		require.NotEqual(t, "none", hint.Reason)
		out := p.HintOutput(b)
		action := i18n.T("baccaratbanque.actionStand")
		if hint.Draw {
			action = i18n.T("baccaratbanque.actionDraw")
		}
		assert.Contains(t, out, i18n.Tf("baccaratbanque.hintAction", "action", action))
		assert.Contains(t, out, i18n.T("baccaratbanque.reason."+hint.Reason))
		assert.NotContains(t, out, "hint.", "識別子が生のまま出ている")
	})

	t.Run("no hint once the bank has ended", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		require.NoError(t, b.BankerDraw(false))
		require.NoError(t, b.Retire())
		assert.Contains(t, p.HintOutput(b), i18n.T("baccaratbanque.noHint"))
	})

	// **棋譜は終局してから。** 途中で読めると相手の判断が漏れる。
	t.Run("the transcript is empty until the bank ends, then names the seats", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		assert.Contains(t, p.ActionLogOutput(b), i18n.T("cuiActionLogEmpty"))

		require.NoError(t, b.BankerDraw(false))
		require.NoError(t, b.Retire())
		out := p.ActionLogOutput(b)
		assert.Contains(t, out, i18n.T("cuiActionLogHeader"))
		assert.Contains(t, out, i18n.T("baccaratbanque.role.right"),
			"棋譜が席をロール名で呼んでいない")
		assert.NotContains(t, out, "CPU 1")
	})
}

func TestBaccaratBanqueWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.BaccaratBanqueWebPresenter)

	decode := func(t *testing.T, s string) *controller.BaccaratBanqueWebOutput {
		t.Helper()
		var out controller.BaccaratBanqueWebOutput
		require.NoError(t, json.Unmarshal([]byte(s), &out))
		return &out
	}

	t.Run("carries three seats, all face up, with roles", func(t *testing.T) {
		out := decode(t, p.Output(banqueAtBankerDecision(t), nil))
		require.Len(t, out.Players, domain.BaccaratBanquePlayerCnt)
		assert.Equal(t, []string{"banker", "right", "left"},
			[]string{out.Players[0].Role, out.Players[1].Role, out.Players[2].Role})
		assert.True(t, out.Players[0].IsHuman)
		// **バカラは全部表向き。** 伏せた札は Design が空で出るので、
		// 3 席すべてに実体のある札が並んでいることを見る。
		for _, pl := range out.Players {
			assert.NotEmpty(t, pl.Cards)
			for _, c := range pl.Cards {
				assert.NotEmpty(t, c.Design, "伏せた札が混じっている")
			}
			assert.Equal(t, len(pl.Cards) >= 2, true)
		}
	})

	t.Run("carries the coup, the tenure, the shoe and the phase", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		out := decode(t, p.Output(b, nil))
		assert.Equal(t, b.GetCoupNumber(), out.CoupNumber)
		assert.Equal(t, b.GetBankHeld(), out.BankHeld)
		assert.Equal(t, b.GetShoeRemaining(), out.ShoeRemaining)
		assert.Equal(t, domain.BaccaratBanquePhaseBanker, out.Phase)
		assert.True(t, out.IsHumanTurn)
		assert.Equal(t, "baccaratbanque.bankerPhase", out.MessageCode)
	})

	t.Run("the settlement carries both sides and the bank net", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		require.NoError(t, b.BankerDraw(false))
		out := decode(t, p.Output(b, nil))
		require.NotNil(t, out.LastResult)
		require.Len(t, out.LastResult.Sides, 2)
		assert.Equal(t, b.GetLastResult().BankerTotal, out.LastResult.BankerTotal)
		assert.Equal(t, b.GetLastResult().BankerDelta, out.LastResult.BankerDelta)
		assert.Equal(t, "baccaratbanque.resultPhase", out.MessageCode)
	})

	t.Run("retiring is reported apart from being broken", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		require.NoError(t, b.BankerDraw(false))
		require.NoError(t, b.Retire())
		out := decode(t, p.Output(b, nil))
		assert.True(t, out.GameEndFlag)
		assert.True(t, out.Retired)
		assert.Equal(t, "baccaratbanque.result.retired", out.MessageCode)
	})

	t.Run("an error carries its message code", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		err := b.Retire()
		require.Error(t, err)
		out := decode(t, p.Output(b, err))
		assert.Equal(t, "baccaratbanque.errNotResultPhase", out.MessageCode)
		assert.NotEmpty(t, out.Message)
	})

	t.Run("hint and log", func(t *testing.T) {
		b := banqueAtBankerDecision(t)
		hint := decode(t, p.HintOutput(b))
		assert.Equal(t, "baccaratbanque.hintRequested", hint.MessageCode)
		assert.NotEmpty(t, hint.HintReason)

		// 棋譜は終局まで空。
		assert.NotContains(t, p.ActionLogOutput(b), "bankEnd")
		require.NoError(t, b.BankerDraw(false))
		require.NoError(t, b.Retire())
		assert.Contains(t, p.ActionLogOutput(b), "bankEnd")
	})
}
