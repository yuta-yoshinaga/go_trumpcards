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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestTrenteEtQuaranteCuiPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "チップ")
}

func TestTrenteEtQuaranteCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestTrenteEtQuaranteCuiPresenter_OutputResult(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Noir")
}

func TestTrenteEtQuaranteCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_HintOutputNone(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	// After the round resolves (result phase) GetHint returns nil.
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// setupTeqCuiMock returns a mock parked in the Result phase with a refait.
func setupTeqCuiMock(refait bool) *interfaces.MockTrenteEtQuaranteGame {
	m := new(interfaces.MockTrenteEtQuaranteGame)
	m.On("GetPhase").Return(domain.TrenteEtQuarantePhaseResult)
	m.On("GetRoundNumber").Return(1)
	m.On("GetChips").Return(1000)
	m.On("GetRemainingDeck").Return(200)
	m.On("GetCurrentBet").Return(domain.TrenteEtQuaranteBetNoir)
	m.On("GetStake").Return(100)
	m.On("GetNoirRow").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	m.On("GetRougeRow").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 1, false)})
	m.On("GetNoirTotal").Return(31)
	m.On("GetRougeTotal").Return(31)
	m.On("GetWinningRow").Return(-1)
	m.On("GetFirstCardRed").Return(false)
	m.On("GetRefait").Return(refait)
	m.On("GetResult").Return(domain.TrenteEtQuaranteResultDraw)
	m.On("GetPayout").Return(-50)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

// #5696: Web は refait.intro/half/edge の3文でルフェを説明するのに、CUI は
// 「31 での Refait — ステークの半額が胴元に渡ります。」の1行だけだった。
// **半額を取られること自体は既に出ている**ので、足りないのは「32〜40 の同点は
// 全額返るのに 31 だけ違う」という、唯一のハウスエッジである理由のほう。
func TestTrenteEtQuaranteCuiPresenter_ExplainsRefait(t *testing.T) {
	p := new(presenter.TrenteEtQuaranteCuiPresenter)

	t.Run("explains why a 31 tie costs half", func(t *testing.T) {
		out := p.Output(setupTeqCuiMock(true), nil)

		assert.Contains(t, out, i18n.T("trenteetquarante.result.refait"))
		assert.Contains(t, out, i18n.T("trenteetquarante.result.refaitWhy"))
	})

	t.Run("stays quiet on an ordinary round", func(t *testing.T) {
		out := p.Output(setupTeqCuiMock(false), nil)

		assert.NotContains(t, out, i18n.T("trenteetquarante.result.refaitWhy"))
	})
}

// **なぜその配当なのかは出目差で決まる。**Web は `result.margin` で勝ち列・差・
// 両列の出目を出しているのに、CUI はラベルだけだった (#6492)。
func TestTrenteEtQuaranteCuiPresenter_ShowsTheMargin(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TrenteEtQuaranteCuiPresenter)

	withRow := func(row, noir, rouge int, result domain.TrenteEtQuaranteResult) string {
		m := setupTeqCuiMock(false)
		for _, name := range []string{"GetWinningRow", "GetNoirTotal", "GetRougeTotal", "GetResult"} {
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, name)
		}
		m.On("GetWinningRow").Return(row)
		m.On("GetNoirTotal").Return(noir)
		m.On("GetRougeTotal").Return(rouge)
		m.On("GetResult").Return(result)
		return p.Output(m, nil)
	}

	// **低いほうが勝つ。**Noir 33 / Rouge 38 なら Noir が 5 点差で勝ち。
	t.Run("noir wins by the lower total", func(t *testing.T) {
		out := withRow(domain.TrenteEtQuaranteRowNoir, 33, 38, domain.TrenteEtQuaranteResultWin)
		assert.Contains(t, out, i18n.Tf("trenteetquarante.result.margin",
			"winner", i18n.T("trenteetquarante.winningNoir"),
			"diff", "5", "winnerTotal", "33", "loserTotal", "38"))
		assert.NotContains(t, out, "{{")
	})

	// 引く順を逆にすると符号が反転する ── 勝ち列が Rouge の場合も見る。
	t.Run("rouge wins by the lower total", func(t *testing.T) {
		out := withRow(domain.TrenteEtQuaranteRowRouge, 39, 32, domain.TrenteEtQuaranteResultWin)
		assert.Contains(t, out, i18n.Tf("trenteetquarante.result.margin",
			"winner", i18n.T("trenteetquarante.winningRouge"),
			"diff", "7", "winnerTotal", "32", "loserTotal", "39"))
	})

	// 両列が同じ出目のときは勝ち列が無い (プッシュ)。出すと存在しない差を報告する。
	//
	// 見るのは**書式の地の文**であって勝ち列の名前ではない。勝ち列が無いまま
	// 行を組むと名前だけが空になり、名前で探す表明は素通りする。
	t.Run("stays quiet on a push", func(t *testing.T) {
		lit := longestLiteralRun(i18n.T("trenteetquarante.result.margin"))
		require.NotEmpty(t, lit)
		out := withRow(domain.TrenteEtQuaranteRowNone, 35, 35, domain.TrenteEtQuaranteResultDraw)
		assert.NotContains(t, out, lit)
		// 負のコントロール: 勝ち列があれば同じ卓でも出る。
		assert.Contains(t, withRow(domain.TrenteEtQuaranteRowNoir, 35, 38, domain.TrenteEtQuaranteResultWin), lit)
	})
}

// longestLiteralRun は i18n の書式から、プレースホルダに挟まれた最長の地の文を返す。
func longestLiteralRun(tmpl string) string {
	best := ""
	for _, seg := range strings.Split(tmpl, "}}") {
		if i := strings.Index(seg, "{{"); i >= 0 {
			seg = seg[:i]
		}
		if len(seg) > len(best) {
			best = seg
		}
	}
	return strings.TrimSpace(best)
}
