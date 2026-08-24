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

// piedmonteseGame は配り終えた卓を返す (既定は 4 人)。
func piedmonteseGame(seats int) *domain.PiedmonteseTarot {
	cfg := domain.DefaultPiedmonteseTarotConfig()
	cfg.Seats = seats
	g := domain.NewPiedmonteseTarot(domain.NewPiedmonteseTarotPlayersForTest(seats), cfg)
	g.Reset()
	return g
}

func TestPiedmonteseTarotCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.PiedmonteseTarotCuiPresenter)

	t.Run("scarto phase lists what may be buried", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := p.Output(g, nil)
		// **訳が引けていることまで見る。** キーの一致だけを見ると、ロケールが
		// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
		assert.Contains(t, out, "ピエモンテ・タロッコ")
		assert.NotContains(t, out, "piedmontesetarot.", "生キーが出ている")
		assert.Contains(t, out, i18n.T("piedmontesetarot.discardableLegend"))
		assert.Contains(t, out, strings.SplitN(i18n.T("piedmontesetarot.discardableList"), "{{", 2)[0])
		// 捨てる枚数は卓が決める。4 人卓は 2 枚。
		assert.Contains(t, out, i18n.Tf("piedmontesetarot.promptScarto",
			"name", "あなた", "n", "2"))
	})

	// **3 人卓は 3 枚。** 枚数を固定で書いていると、どちらかの卓で必ず嘘になる。
	t.Run("a three-handed table buries three", func(t *testing.T) {
		g := piedmonteseGame(3)
		out := p.Output(g, nil)
		assert.Contains(t, out, i18n.Tf("piedmontesetarot.promptScarto", "name", "あなた", "n", "3"))
	})

	t.Run("play phase lists the playable cards", func(t *testing.T) {
		g := piedmonteseGame(4)
		require.NoError(t, g.PlayerScarto(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))
		for g.GetPhase() == domain.PiedmonteseTarotPhasePlay && !g.IsHumanTurn() {
			g.CpuPlay()
		}
		if g.GetPhase() != domain.PiedmonteseTarotPhasePlay {
			t.Skip("配りによっては人間の手番の前にトリックが揃う")
		}
		out := p.Output(g, nil)
		// **出せる札を出す。** Web は playableIndices でボタンを絞るのに、CUI に
		// それが無いと総当たりで探すことになる。
		assert.Contains(t, out, strings.SplitN(i18n.T("piedmontesetarot.playableList"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("piedmontesetarot.promptPlayHelp"))
	})

	t.Run("errors are shown", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := p.Output(g, assert.AnError)
		assert.Contains(t, out, assert.AnError.Error())
	})

	// **点は 1/3 単位まで出す。** 切り捨てると卓の合計が 78 点にならない。
	t.Run("card points keep the thirds", func(t *testing.T) {
		assert.Equal(t, "26", domain.PiedmonteseTarotFormatThirds(78))
		assert.Equal(t, "26 1/3", domain.PiedmonteseTarotFormatThirds(79))
		assert.Equal(t, "26 2/3", domain.PiedmonteseTarotFormatThirds(80))
		assert.Equal(t, "0", domain.PiedmonteseTarotFormatThirds(0))
		assert.Equal(t, "-1 2/3", domain.PiedmonteseTarotFormatThirds(-5))
	})

	t.Run("hint names the recommended cards", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := p.HintOutput(g)
		assert.Contains(t, out, i18n.T("piedmontesetarot.hintReasonScarto"))
		assert.NotContains(t, out, "piedmontesetarot.hint", "理由キーが生で出ている")
	})

	t.Run("english renders too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(piedmonteseGame(4), nil)
		assert.Contains(t, out, "Tarocco Piemontese")
		assert.NotContains(t, out, "ピエモンテ", "日本語が漏れている")
	})
}

func TestPiedmonteseTarotWebPresenter_Output(t *testing.T) {
	p := new(presenter.PiedmonteseTarotWebPresenter)

	decode := func(t *testing.T, raw string) map[string]any {
		t.Helper()
		var out map[string]any
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		return out
	}

	t.Run("reports the table and the talon", func(t *testing.T) {
		out := decode(t, p.Output(piedmonteseGame(4), nil))
		assert.Equal(t, float64(2), out["talonSize"], "4 人卓のタロンは 2 枚")
		assert.Equal(t, float64(19), out["trickCount"], "4 人卓は 19 トリック")
		assert.Len(t, out["players"], 4)

		out3 := decode(t, p.Output(piedmonteseGame(3), nil))
		assert.Equal(t, float64(3), out3["talonSize"], "3 人卓のタロンは 3 枚")
		assert.Equal(t, float64(25), out3["trickCount"])
		assert.Len(t, out3["players"], 3)
	})

	// **点は生の 1/3 単位と読める形の両方を出す。** 234 という数字だけを出すと、
	// 78 点のゲームの画面に意味の分からない数が並ぶ。
	t.Run("sends the card points in both forms", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := decode(t, p.Output(g, nil))
		players := out["players"].([]any)
		dealer := players[g.GetDealerIdx()].(map[string]any)
		thirds := int(dealer["cardThirds"].(float64))
		assert.Equal(t, domain.PiedmonteseTarotFormatThirds(thirds), dealer["cardPoints"])
	})

	t.Run("only the human's hand is sent", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := decode(t, p.Output(g, nil))
		for _, raw := range out["players"].([]any) {
			pl := raw.(map[string]any)
			cards := pl["cards"].([]any)
			if pl["isHuman"].(bool) {
				assert.NotEmpty(t, cards, "自分の手札が見えない")
				continue
			}
			assert.Empty(t, cards, "CPU の手札が見えている")
		}
	})

	t.Run("playable indices only on the human play turn", func(t *testing.T) {
		g := piedmonteseGame(4)
		out := decode(t, p.Output(g, nil))
		assert.Empty(t, out["playableIndices"], "スカルトの最中に出せる札は無い")
	})

	t.Run("errors carry their message code", func(t *testing.T) {
		g := piedmonteseGame(4)
		err := g.PlayerScarto([]int{0})
		require.Error(t, err)
		out := decode(t, p.Output(g, err))
		assert.Equal(t, "piedmontesetarot.errScartoCount", out["messageCode"])
	})

	t.Run("hint output marks that it was requested", func(t *testing.T) {
		out := decode(t, p.HintOutput(piedmonteseGame(4)))
		assert.Equal(t, "piedmontesetarot.hintRequested", out["messageCode"])
		assert.NotNil(t, out["hint"])
	})

	t.Run("action log is JSON", func(t *testing.T) {
		out := decode(t, p.ActionLogOutput(piedmonteseGame(4)))
		assert.NotNil(t, out)
	})
}

// piedmonteseFinishedDeal は 1 ディールを打ち切った卓を返す。
func piedmonteseFinishedDeal(t *testing.T, seats int) *domain.PiedmonteseTarot {
	t.Helper()
	cfg := domain.DefaultPiedmonteseTarotConfig()
	cfg.Seats = seats
	cfg.TargetDeals = 2
	g := domain.NewPiedmonteseTarot(domain.NewPiedmonteseTarotPlayersForTest(seats), cfg)
	g.Reset()
	for step := 0; step < 6000; step++ {
		switch g.GetPhase() {
		case domain.PiedmonteseTarotPhaseScarto:
			if g.IsHumanScartoTurn() {
				require.NoError(t, g.PlayerScarto(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))
				continue
			}
			g.CpuScarto()
		case domain.PiedmonteseTarotPhasePlay:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerPlay(g.GetPlayableIndices(g.GetCurrentPlayerIdx())[0]))
				continue
			}
			g.CpuPlay()
		case domain.PiedmonteseTarotPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.PiedmonteseTarotPhaseTrickEnd {
				g.NextTrick()
			}
		default:
			return g
		}
	}
	require.FailNow(t, "ディールが終わらない")
	return nil
}

// **精算の内訳を出す。** 取り分だけを並べると、変動の数字がどこから来たのか
// 検算できない (席数倍ちがう)。
func TestPiedmonteseTarotCuiPresenter_RoundEndBreakdown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.PiedmonteseTarotCuiPresenter)

	g := piedmonteseFinishedDeal(t, 4)
	out := p.Output(g, nil)

	assert.Contains(t, out, strings.SplitN(i18n.T("piedmontesetarot.promptRoundEnd"), "{{", 2)[0])
	assert.Contains(t, out, i18n.T("piedmontesetarot.roundEndFormulaLine"))
	assert.Contains(t, out, strings.SplitN(i18n.T("piedmontesetarot.roundEndTotal"), "{{", 2)[0])
	// 卓の合計は必ず 78 点。
	assert.Contains(t, out, domain.PiedmonteseTarotFormatThirds(domain.PiedmonteseTarotTotalThirds))
	// 席ごとの行が全部出る。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Contains(t, out, domain.PiedmonteseTarotFormatThirds(g.GetCardThirds(i)))
	}
	assert.NotContains(t, out, "piedmontesetarot.", "生キーが出ている")

	// 結果ラベルは 3 通りとも訳が引ける。
	for _, o := range []domain.PiedmonteseTarotOutcome{
		domain.PiedmonteseTarotOutcomeNone,
		domain.PiedmonteseTarotOutcomeWin,
		domain.PiedmonteseTarotOutcomeLoss,
	} {
		g.SetOutcomeForTest(o)
		assert.NotContains(t, p.Output(g, nil), "outcome", "結果ラベルが生キー")
	}
}

// 終局の画面は勝者を出す。
func TestPiedmonteseTarotCuiPresenter_GameEnd(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.PiedmonteseTarotCuiPresenter)

	g := piedmonteseGame(4)
	g.SetGameEndForTest(0)
	out := p.Output(g, nil)
	assert.Contains(t, out, strings.SplitN(i18n.T("piedmontesetarot.gameEnd"), "{{", 2)[0])

	// 引き分けでも落ちない (勝者 -1)。
	g.SetGameEndForTest(-1)
	assert.NotEmpty(t, p.Output(g, nil))
}

// 棋譜はテキストで出る。
func TestPiedmonteseTarotCuiPresenter_ActionLog(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.PiedmonteseTarotCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(piedmonteseGame(4)))
}

// ヒントが札を伴わない場面 (次のトリックへ、など) も出せる。
func TestPiedmonteseTarotCuiPresenter_HintWithoutCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.PiedmonteseTarotCuiPresenter)

	g := piedmonteseFinishedDeal(t, 4)
	out := p.HintOutput(g)
	assert.Contains(t, out, i18n.T("piedmontesetarot.hintReasonNextRound"))
	assert.NotContains(t, out, "piedmontesetarot.hint", "理由キーが生で出ている")
}

// **プレイ中は出せる札だけを返す。** ここが空のままだと Web のボタンが
// 1 つも押せない。
func TestPiedmonteseTarotWebPresenter_PlayableIndicesOnTheHumanTurn(t *testing.T) {
	p := new(presenter.PiedmonteseTarotWebPresenter)
	g := piedmonteseGame(4)
	require.NoError(t, g.PlayerScarto(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))
	for g.GetPhase() == domain.PiedmonteseTarotPhasePlay && !g.IsHumanTurn() {
		g.CpuPlay()
	}
	if !g.IsHumanTurn() {
		t.Skip("配りによっては人間の手番の前にトリックが揃う")
	}
	var out struct {
		PlayableIndices []int `json:"playableIndices"`
		IsHumanTurn     bool  `json:"isHumanTurn"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
	assert.True(t, out.IsHumanTurn)
	assert.NotEmpty(t, out.PlayableIndices)
}

// **フェーズごとにメッセージコードが変わる。** 画面の案内が 1 種類しか出ないと、
// いま何を待たれているのか分からない。
func TestPiedmonteseTarotWebPresenter_MessageCodesPerPhase(t *testing.T) {
	p := new(presenter.PiedmonteseTarotWebPresenter)
	code := func(g *domain.PiedmonteseTarot) string {
		var out struct {
			MessageCode string `json:"messageCode"`
		}
		require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
		return out.MessageCode
	}

	g := piedmonteseGame(4)
	assert.Equal(t, "piedmontesetarot.scartoPhase", code(g))

	require.NoError(t, g.PlayerScarto(domain.PiedmonteseTarotCpuScartoForTest(g, g.GetDealerIdx())))
	if g.GetPhase() == domain.PiedmonteseTarotPhasePlay {
		want := "piedmontesetarot.playPhase.lead"
		if len(g.GetCurrentTrick()) > 0 {
			want = "piedmontesetarot.playPhase.follow"
		}
		assert.Equal(t, want, code(g))
	}

	done := piedmonteseFinishedDeal(t, 4)
	assert.Contains(t, []string{
		"piedmontesetarot.roundEnd", "piedmontesetarot.roundEnd.win", "piedmontesetarot.roundEnd.loss",
	}, code(done))

	done.SetGameEndForTest(0)
	assert.Contains(t, []string{
		"piedmontesetarot.result.humanWin", "piedmontesetarot.result.cpuWin", "piedmontesetarot.result.draw",
	}, code(done))
	done.SetGameEndForTest(-1)
	assert.Equal(t, "piedmontesetarot.result.draw", code(done))
}
