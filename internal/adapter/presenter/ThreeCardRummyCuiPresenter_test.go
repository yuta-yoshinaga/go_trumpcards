//go:build test && (!js || !wasm || casino)

package presenter

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tcrCard は 1 枚を作る。**役 (0 点) を踏まないように値とスートを選ぶこと。**
// 同ランク 3 枚も同スート連番 3 枚も 0 点になるので、ただの合計を意図した手は
// スートを散らし、値を連番から外す。
func tcrCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, true)
}

// tcrPlayerHand は 2+6+9=17 点の手。スートばらばら・値も連番でないので役ではない。
func tcrPlayerHand() []*domain.Card {
	return []*domain.Card{
		tcrCard(domain.CardDesignSpade, 2),
		tcrCard(domain.CardDesignHeart, 6),
		tcrCard(domain.CardDesignClover, 9),
	}
}

// tcrDealerHand は 4+8+10(K)=22 点の手。プレイヤーの手と 1 枚も文字列が
// 重ならないので「まだ出ていない」ことを NotContains で言える。
func tcrDealerHand() []*domain.Card {
	return []*domain.Card{
		tcrCard(domain.CardDesignDiamond, 4),
		tcrCard(domain.CardDesignClover, 8),
		tcrCard(domain.CardDesignSpade, 13),
	}
}

// tcrMeldHand は同ランク 3 枚 = 0 点 (このゲームの最強手)。
func tcrMeldHand() []*domain.Card {
	return []*domain.Card{
		tcrCard(domain.CardDesignSpade, 5),
		tcrCard(domain.CardDesignHeart, 5),
		tcrCard(domain.CardDesignClover, 5),
	}
}

// tcrGame は指定フェーズの卓を返す。**配りを通さない**ので、シャッフル順に
// 一切依存しない。手札・点数は呼び出し側がセッターで置く。
func tcrGame(t *testing.T, phase int) *domain.ThreeCardRummy {
	t.Helper()
	g := domain.NewDefaultThreeCardRummy()
	g.SetPhase(phase)
	return g
}

func TestThreeCardRummyCuiPresenter_Output_BetPhaseExplainsTheInvertedRule(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)
	g := tcrGame(t, domain.ThreeCardRummyPhaseBet)

	out := p.Output(g, nil)

	assert.Contains(t, out, "チップ: 1000")
	assert.Contains(t, out, "フェーズ: ベット")
	// **賭ける前に読ませる 2 行。**「低いほど強い」はこのゲーム最大の意外性で、
	// 絵札=10 / A=1 / 役=0 点まで含めてベット前にしか出す場所が無い。
	assert.Contains(t, out, "3枚の合計が低いほど強く、0点が最強")
	assert.Contains(t, out, "絵札=10、A=1")
	assert.Contains(t, out, "ディーラーは合計20点以下でクオリファイします")
}

func TestThreeCardRummyCuiPresenter_Output_ScoringNotesOnlyInTheBetPhase(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	for _, tt := range []struct {
		name  string
		phase int
	}{
		{"action phase", domain.ThreeCardRummyPhaseAction},
		{"end phase", domain.ThreeCardRummyPhaseEnd},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := tcrGame(t, tt.phase)
			g.SetPlayerHand(tcrPlayerHand())
			g.SetPlayerScore(17)

			out := p.Output(g, nil)

			// ルール説明は賭ける前の 1 回だけ。毎画面出すと盤が埋まる。
			assert.NotContains(t, out, "3枚の合計が低いほど強く、0点が最強")
			assert.NotContains(t, out, "ディーラーは合計20点以下でクオリファイします")
		})
	}
}

func TestThreeCardRummyCuiPresenter_Output_PlayerScoreShownFromTheActionPhase(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)
	g := tcrGame(t, domain.ThreeCardRummyPhaseAction)
	g.SetPlayerHand(tcrPlayerHand())
	g.SetPlayerScore(17)

	out := p.Output(g, nil)

	assert.Contains(t, out, "あなた")
	// **点数がそのまま強さ。** 役名ではなく数字を出す。
	assert.Contains(t, out, "点数: 17（低いほど強い）")
	assert.Contains(t, out, "SPADE 2")
	assert.Contains(t, out, "HEART 6")
	assert.Contains(t, out, "CLOVER 9")
}

func TestThreeCardRummyCuiPresenter_Output_ScoreIsSetAtDealTime(t *testing.T) {
	// **配った時点で自分の点は決まっている** (domain の deal が playerScore を
	// 置く)。resolve まで 0 のままだと、アクションフェーズの画面が
	// 「0 点 = 最強」を出してしまう。
	//
	// 配りは乱数なので、役 (0 点 = 約 0.45%) を引いたら引き直す。点数そのものは
	// 配られた手から計算するので、どの手が来ても期待値が決まる。
	p := new(ThreeCardRummyCuiPresenter)
	g := domain.NewDefaultThreeCardRummy()
	var score int
	for range 100 {
		g.Reset()
		require.NoError(t, g.Bet(10, 0))
		score = domain.ThreeCardRummyScore(g.GetPlayerHand())
		if score != domain.ThreeCardRummyPerfectScore {
			break
		}
	}
	require.NotEqual(t, domain.ThreeCardRummyPerfectScore, score, "100 回続けて役は引かない")
	require.Equal(t, domain.ThreeCardRummyPhaseAction, g.GetPhase())

	out := p.Output(g, nil)

	assert.Contains(t, out, "点数: "+strconv.Itoa(score)+"（低いほど強い）")
}

func TestThreeCardRummyCuiPresenter_Output_MeldRendersAsAHandNotAsZero(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)
	g := tcrGame(t, domain.ThreeCardRummyPhaseAction)
	g.SetPlayerHand(tcrMeldHand())
	g.SetPlayerScore(domain.ThreeCardRummyPerfectScore)

	out := p.Output(g, nil)

	// 0 は「点が無い」ではなく最強の役。素の 0 と書くと弱そうに見える。
	assert.Contains(t, out, "0点（役 = 最強）")
	assert.NotContains(t, out, "点数: 0（低いほど強い）")
}

func TestThreeCardRummyCuiPresenter_Output_DealerIsHiddenUntilTheEnd(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	for _, tt := range []struct {
		name  string
		phase int
	}{
		{"bet phase", domain.ThreeCardRummyPhaseBet},
		{"action phase", domain.ThreeCardRummyPhaseAction},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := tcrGame(t, tt.phase)
			g.SetPlayerHand(tcrPlayerHand())
			g.SetPlayerScore(17)
			g.SetDealerHand(tcrDealerHand())
			g.SetDealerScore(22)

			out := p.Output(g, nil)

			// 伏せたままの 3 枚も、その点数も出してはいけない。
			assert.NotContains(t, out, "DIAMOND 4")
			assert.NotContains(t, out, "CLOVER 8")
			assert.NotContains(t, out, "SPADE 13")
			assert.NotContains(t, out, "点数: 22")
			assert.NotContains(t, out, "クオリファイせず")
		})
	}
}

func TestThreeCardRummyCuiPresenter_Output_DealerRevealedAtTheEnd(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	t.Run("qualified dealer", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseEnd)
		g.SetPlayerHand(tcrPlayerHand())
		g.SetPlayerScore(17)
		g.SetDealerHand([]*domain.Card{
			tcrCard(domain.CardDesignDiamond, 4),
			tcrCard(domain.CardDesignClover, 6),
			tcrCard(domain.CardDesignSpade, 9),
		})
		g.SetDealerScore(19)
		g.SetDealerQualified(true)

		out := p.Output(g, nil)

		assert.Contains(t, out, "ディーラー")
		assert.Contains(t, out, "点数: 19（低いほど強い）")
		assert.Contains(t, out, "DIAMOND 4")
		assert.Contains(t, out, "CLOVER 6")
		assert.Contains(t, out, "SPADE 9")
		assert.Contains(t, out, "ディーラー: クオリファイ")
		// 「クオリファイ」は「クオリファイせず」の接頭辞なので、否定側も見る。
		assert.NotContains(t, out, "クオリファイせず")
	})

	t.Run("dealer does not qualify", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseEnd)
		g.SetPlayerHand(tcrPlayerHand())
		g.SetPlayerScore(17)
		g.SetDealerHand(tcrDealerHand())
		g.SetDealerScore(22)
		g.SetDealerQualified(false)
		g.SetPlayBet(100) // 勝負した手。配当の帰結を書いてよい局面。

		out := p.Output(g, nil)

		assert.Contains(t, out, "点数: 22（低いほど強い）")
		assert.Contains(t, out, "クオリファイせず")
		// 帰結の一文は勝負した手にだけ付く。降りた手での扱いは
		// TestThreeCardRummyCuiPresenter_QualifyConsequenceOnlyAfterAContest。
		assert.Contains(t, out, "プレイはプッシュ")
	})
}

func TestThreeCardRummyCuiPresenter_Output_EndPhaseResults(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	tests := []struct {
		name    string
		result  domain.GameResult
		playBet int
		want    string
		notWant []string
	}{
		{
			name: "player wins", result: domain.GameResultWin, playBet: 10,
			want: "あなたの勝ちです", notWant: []string{"ディーラーの勝ちです", "フォールドしました", "引き分け"},
		},
		{
			name: "dealer wins", result: domain.GameResultLose, playBet: 10,
			want: "ディーラーの勝ちです", notWant: []string{"あなたの勝ちです", "フォールドしました"},
		},
		{
			// **降りた負けは負けと書き分ける。** プレイベットが 0 なのが唯一の手掛かり。
			name: "folded", result: domain.GameResultLose, playBet: 0,
			want: "フォールドしました", notWant: []string{"ディーラーの勝ちです", "あなたの勝ちです"},
		},
		{
			name: "push", result: domain.GameResultDraw, playBet: 10,
			want: "引き分け（プッシュ）", notWant: []string{"あなたの勝ちです", "ディーラーの勝ちです"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tcrGame(t, domain.ThreeCardRummyPhaseEnd)
			g.SetPlayerHand(tcrPlayerHand())
			g.SetPlayerScore(17)
			g.SetDealerHand(tcrDealerHand())
			g.SetDealerScore(22)
			g.SetAnteBet(10)
			g.SetPlayBet(tt.playBet)
			g.SetResult(tt.result)
			g.SetGameEndFlag(true)

			out := p.Output(g, nil)

			assert.Contains(t, out, tt.want)
			for _, nw := range tt.notWant {
				assert.NotContains(t, out, nw)
			}
		})
	}
}

func TestThreeCardRummyCuiPresenter_Output_PayoutBreakdown(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	// NOTE: 金額そのものは検査していない。プレゼンタは Tf に "ante"/"play"/
	// "payout" を渡すのに、ロケールの差し込み口はいずれも {{amount}} なので、
	// 現状どの行も金額が置換されずに出る (別途報告済みの本番バグ)。
	// ここで固めるのは「どの行がいつ出るか」。
	t.Run("bonus lines appear when non-zero", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseEnd)
		g.SetPlayerHand(tcrMeldHand())
		g.SetPlayerScore(domain.ThreeCardRummyPerfectScore)
		g.SetDealerHand(tcrDealerHand())
		g.SetDealerScore(22)
		g.SetAnteBet(10)
		g.SetPlayBet(10)
		g.SetResult(domain.GameResultWin)
		g.SetAntePayout(20)
		g.SetPlayPayout(20)
		g.SetAnteBonusPayout(90)
		g.SetLowBonusPayout(1010)
		g.SetGameEndFlag(true)

		out := p.Output(g, nil)

		assert.Contains(t, out, "アンテ:")
		assert.Contains(t, out, "プレイ:")
		assert.Contains(t, out, "アンテボーナス:")
		assert.Contains(t, out, "ローボーナス:")
		assert.Contains(t, out, "配当合計:")
	})

	t.Run("zero bonus lines are omitted", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseEnd)
		g.SetPlayerHand(tcrPlayerHand())
		g.SetPlayerScore(17)
		g.SetDealerHand(tcrDealerHand())
		g.SetDealerScore(22)
		g.SetAnteBet(10)
		g.SetPlayBet(0) // フォールドしたのでプレイベットは置いていない
		g.SetResult(domain.GameResultLose)
		g.SetAnteBonusPayout(0)
		g.SetLowBonusPayout(0)
		g.SetGameEndFlag(true)

		out := p.Output(g, nil)

		assert.Contains(t, out, "アンテ:")
		assert.Contains(t, out, "配当合計:")
		assert.NotContains(t, out, "アンテボーナス")
		assert.NotContains(t, out, "ローボーナス")
		assert.NotContains(t, out, "プレイ:")
	})

	t.Run("nothing is settled before the game ends", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseAction)
		g.SetPlayerHand(tcrPlayerHand())
		g.SetPlayerScore(17)
		g.SetAnteBet(10)
		g.SetAnteBonusPayout(90)

		out := p.Output(g, nil)

		assert.NotContains(t, out, "配当合計")
		assert.NotContains(t, out, "アンテボーナス")
	})
}

func TestThreeCardRummyCuiPresenter_Output_RendersTheLastError(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)
	g := tcrGame(t, domain.ThreeCardRummyPhaseBet)

	out := p.Output(g, domain.NewDomainError(domain.ErrWrongPhase, "Bet is only allowed during the bet phase."))

	assert.Contains(t, out, "Bet is only allowed during the bet phase.")
	// 拒否は盤の中の 1 行なので、盤ごと stderr に飛ばす ErrorPrefix ではなく
	// 行単位の marker が付く。
	assert.Contains(t, out, i18n.ErrorLinePrefix)
	assert.NotContains(t, p.Output(g, nil), "Bet is only allowed")
}

func TestThreeCardRummyCuiPresenter_HintOutput(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	hintAt := func(phase, score int) string {
		g := tcrGame(t, phase)
		g.SetPlayerScore(score)
		return p.HintOutput(g)
	}

	t.Run("no advice outside the action phase", func(t *testing.T) {
		for _, phase := range []int{domain.ThreeCardRummyPhaseBet, domain.ThreeCardRummyPhaseEnd} {
			out := hintAt(phase, 5)
			assert.Contains(t, out, "いま助言できることはありません")
			assert.NotContains(t, out, "プレイ推奨")
			assert.NotContains(t, out, "フォールド推奨")
		}
	})

	t.Run("play below the dealer's qualifying limit", func(t *testing.T) {
		// 19 点はクオリファイ上限 20 の 1 点下 = 境界のすぐ内側。
		out := hintAt(domain.ThreeCardRummyPhaseAction, domain.ThreeCardRummyDealerQualifyMax-1)
		assert.Contains(t, out, "プレイ推奨")
		assert.NotContains(t, out, "フォールド推奨")
	})

	t.Run("fold at the qualifying limit", func(t *testing.T) {
		// 20 点ちょうどは引き分け含みなので降りる = 境界の外側。
		out := hintAt(domain.ThreeCardRummyPhaseAction, domain.ThreeCardRummyDealerQualifyMax)
		assert.Contains(t, out, "フォールド推奨")
		assert.NotContains(t, out, "プレイ推奨")
	})

	t.Run("a meld always plays", func(t *testing.T) {
		out := hintAt(domain.ThreeCardRummyPhaseAction, domain.ThreeCardRummyPerfectScore)
		assert.Contains(t, out, "プレイ推奨")
	})

	t.Run("a hopeless total folds", func(t *testing.T) {
		out := hintAt(domain.ThreeCardRummyPhaseAction, 28)
		assert.Contains(t, out, "フォールド推奨")
	})
}

func TestThreeCardRummyCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(ThreeCardRummyCuiPresenter)

	t.Run("empty before the game ends", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseAction)
		assert.Contains(t, p.ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("transcribes the round once it ends", func(t *testing.T) {
		g := tcrGame(t, domain.ThreeCardRummyPhaseAction)
		g.SetPlayerHand(tcrPlayerHand())
		g.SetDealerHand(tcrDealerHand())
		g.SetAnteBet(10)
		require.NoError(t, g.Fold())

		out := p.ActionLogOutput(g)

		assert.Contains(t, out, "棋譜")
		assert.Contains(t, out, "player folds")
		assert.Contains(t, out, "player folded")
		assert.NotContains(t, out, "棋譜はありません")
	})
}

// TestThreeCardRummyCuiPresenter_MoneyLinesShowNumbers pins the interpolation
// parameter names against the locale placeholders.
//
// **A name mismatch does not error — it prints the placeholder.** The five money
// lines shipped as `アンテ: {{amount}}` because the presenter passes `"ante"`,
// `"play"` and `"payout"` while the locale spelled every one of them
// `{{amount}}`; `i18n.Tf` substitutes nothing and returns the template. Every
// assertion phrased as `Contains(out, i18n.Tf(key, ...))` agrees with that
// output, so only looking for the *digits* catches it.
func TestThreeCardRummyCuiPresenter_MoneyLinesShowNumbers(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)

			g := domain.NewDefaultThreeCardRummy()
			g.SetPhase(domain.ThreeCardRummyPhaseEnd)
			g.SetGameEndFlag(true)
			g.SetResult(domain.GameResultWin)
			g.SetChips(1234)
			g.SetAnteBet(11)
			g.SetPlayBet(22)
			g.SetAnteBonusPayout(33)
			g.SetLowBonusPayout(44)
			g.SetAntePayout(55)
			g.SetPlayPayout(66)

			out := new(ThreeCardRummyCuiPresenter).Output(g, nil)

			assert.NotContains(t, out, "{{", "a placeholder reached the player: %s", out)
			for _, want := range []string{"1234", "11", "22", "33", "44", "198"} {
				assert.Contains(t, out, want, "amount %s is missing from the board", want)
			}
		})
	}
}

// TestThreeCardRummyCuiPresenter_ShowsALosingSideBet keeps the Low Bonus stake
// on the result board even when it paid nothing.
//
// The payout lines are omitted at zero to stay concise, so a losing side bet
// left no trace at all: a player who staked 50 saw an ante line, a play line,
// and a total that silently excluded it. The stake is what makes the missing
// payout readable.
func TestThreeCardRummyCuiPresenter_ShowsALosingSideBet(t *testing.T) {
	g := domain.NewDefaultThreeCardRummy()
	g.SetPhase(domain.ThreeCardRummyPhaseEnd)
	g.SetGameEndFlag(true)
	g.SetResult(domain.GameResultWin)
	g.SetAnteBet(100)
	g.SetPlayBet(100)
	g.SetLowBonusBet(50)
	g.SetLowBonusPayout(0) // 21 点以上だったので側注は没収

	out := new(ThreeCardRummyCuiPresenter).Output(g, nil)
	assert.Contains(t, out, "50", "the forfeited Low Bonus stake must still be on the board")

	// 賭けていないラウンドで行が出ないことも押さえる。出しっぱなしにすると
	// 「50」が別の数字にたまたま当たるだけの空振りアサーションになる。
	g.SetLowBonusBet(0)
	assert.NotContains(t, new(ThreeCardRummyCuiPresenter).Output(g, nil),
		i18n.T("threecardrummy.lowBonusLine")[:6],
		"no Low Bonus line when no side bet was placed")
}

// TestThreeCardRummyCuiPresenter_QualifyConsequenceOnlyAfterAContest keeps the
// payout clause off a folded hand.
//
// The not-qualified line used to read "クオリファイせず（アンテのみ配当、プレイは
// プッシュ）" unconditionally. After a fold the ante is *forfeited*, so that
// sentence told the player the opposite of what happened to their money.
func TestThreeCardRummyCuiPresenter_QualifyConsequenceOnlyAfterAContest(t *testing.T) {
	newBoard := func(playBet int) string {
		g := domain.NewDefaultThreeCardRummy()
		g.SetPhase(domain.ThreeCardRummyPhaseEnd)
		g.SetGameEndFlag(true)
		g.SetResult(domain.GameResultLose)
		g.SetDealerHand([]*domain.Card{
			domain.NewCard(0, 13, true), domain.NewCard(1, 12, true), domain.NewCard(2, 1, true),
		}) // 21 点 -- クオリファイ上限のすぐ上
		g.SetDealerScore(21)
		g.SetDealerQualified(false)
		g.SetAnteBet(100)
		g.SetPlayBet(playBet)
		return new(ThreeCardRummyCuiPresenter).Output(g, nil)
	}

	consequence := i18n.T("threecardrummy.notQualifiedPayout")
	fact := i18n.T("threecardrummy.notQualified")
	require.NotEqual(t, "threecardrummy.notQualifiedPayout", consequence, "key must resolve")
	require.NotEqual(t, "threecardrummy.notQualified", fact, "key must resolve")

	played := newBoard(100)
	assert.Contains(t, played, fact)
	assert.Contains(t, played, consequence, "勝負した手には配当の帰結を書く")

	folded := newBoard(0)
	assert.Contains(t, folded, fact, "降りても資格の有無そのものは出す")
	assert.NotContains(t, folded, consequence,
		"降りてアンテを没収された手に『アンテのみ配当』とは書けない")
}
