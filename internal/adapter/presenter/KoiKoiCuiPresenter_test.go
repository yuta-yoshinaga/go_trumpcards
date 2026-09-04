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

func TestKoiKoiCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.KoiKoiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed field
}

func TestKoiKoiCuiPresenter_Output_DecisionInfo(t *testing.T) {
	p := new(presenter.KoiKoiCuiPresenter)

	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)

	// No koi-koi yet → the shobu multiplier is x1 (ja locale renders "倍率×1").
	out := p.Output(g, nil)
	assert.Contains(t, out, "×1")
	// Both players' current cumulative scores are shown.
	assert.Contains(t, out, "あなた")
	assert.Contains(t, out, "相手")

	// Once a koi-koi has been declared this round, the multiplier becomes x2.
	g.SetKoikoiCount(1)
	out = p.Output(g, nil)
	assert.Contains(t, out, "×2")
}

func TestKoiKoiCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.KoiKoiCuiPresenter)

	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空フィールド表示。
	g.SetPhase(domain.KoiKoiPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestKoiKoiCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KoiKoiCuiPresenter)

	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	// play-phase hint
	assert.NotEmpty(t, p.HintOutput(g))

	// decision-phase hint
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestKoiKoiCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	p := new(presenter.KoiKoiCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}

// koikoiSankoCards は三光 (松/桜/桐の光 = 5 点) を返す。
func koikoiSankoCards() []*domain.Card {
	return []*domain.Card{domain.NewCard(1, 1, false), domain.NewCard(3, 1, false), domain.NewCard(12, 1, false)}
}

func TestKoiKoiCuiPresenter_RoundResult_HumanWinner(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	g.GetPlayer(0).AddCaptured(koikoiSankoCards())
	require.NoError(t, g.PlayerDecide(false)) // shobu → human wins round
	require.Equal(t, domain.KoiKoiPhaseRoundEnd, g.GetPhase())

	p := new(presenter.KoiKoiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestKoiKoiCuiPresenter_RoundResult_CpuWinner(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	cfg := domain.DefaultKoiKoiConfig()
	cfg.CpuDifficulty = domain.KoiKoiCpuDifficultyEasy // Easy always stops (Shobu)
	g.SetConfig(cfg)
	g.SetCurrentTurn(1)
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	g.GetPlayer(1).AddCaptured(koikoiSankoCards())
	g.CpuDecide() // Easy → stop → CPU wins round
	require.Equal(t, domain.KoiKoiPhaseRoundEnd, g.GetPhase())

	// **CPU 名も i18n と太字を通す。**人間側だけ i18n を通し、CPU は英語
	// リテラルのままだった (#4855 と同型)。色を有効にして太字の有無まで見る。
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.KoiKoiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// **ラウンド勝利行そのものを見る。**プレイヤー一覧は元から太字の CPU 名を
	// 出しているので、出力全体で探すと修正前でも当たってしまう。
	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	var winLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "ラウンド勝利") {
			winLine = line
		}
	}
	require.NotEmpty(t, winLine, "round-win line missing from output: %q", out)
	assert.Contains(t, winLine, bold1)
}

// **同ランクの札が場に 2 枚あると、どちらを取るかは選択になる。**Web は一致する
// 場札をリングで示して pickField を案内するのに、CUI は場札と手札を目で突き合わせる
// しかなかった (#6506)。Basra / Tablanet と同じ共有ヘルパーを使う。
func TestKoiKoiCuiPresenter_ShowsWhichFieldCardsAHandCardCanTake(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	// 卓は手で組む ── 配りに任せると「二択」が出るかどうかが局ごとに変わる。
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetPhase(domain.KoiKoiPhasePlay)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(domain.NewCard(1, 1, false))  // 松: 場に 2 枚ある
	human.AddCard(domain.NewCard(5, 1, false))  // 菖蒲: 場に 1 枚
	human.AddCard(domain.NewCard(11, 1, false)) // 桐: 場に無い
	g.SetFieldCards([]*domain.Card{
		domain.NewCard(1, 2, false),
		domain.NewCard(5, 2, false),
		domain.NewCard(1, 3, false),
	})

	out := new(presenter.KoiKoiCuiPresenter).Output(g, nil)

	// 二択は両方の場札が言えて初めて選べる。
	assert.Contains(t, out, i18n.Tf("koikoi.captureHint",
		"hand", "[0]"+domain.KoiKoiCardLabel(human.GetCard(0)), "table", "[0][2]"))
	// 一択も出る。
	assert.Contains(t, out, i18n.Tf("koikoi.captureHint",
		"hand", "[1]"+domain.KoiKoiCardLabel(human.GetCard(1)), "table", "[1]"))
	// 負のコントロール: 取れない札には注記を付けない。
	assert.NotContains(t, out, "[2]"+domain.KoiKoiCardLabel(human.GetCard(2))+" →")
	assert.NotContains(t, out, "{{")
}
