package presenter_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func buildPlayedCuarenta(t *testing.T) *domain.Cuarenta {
	t.Helper()
	g := domain.NewDefaultCuarenta()
	g.Reset()
	return g
}

func TestCuarentaCuiPresenter_Output(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	out := p.Output(g, nil)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	// i18n (ja) is loaded in this build; assert on rendered Japanese text.
	if !strings.Contains(out, "チーム") {
		t.Errorf("expected localized team text in output, got: %s", out)
	}
}

func TestCuarentaCuiPresenter_Error(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	out := p.Output(g, cuarentaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type cuarentaAssertErr struct{}

func (cuarentaAssertErr) Error() string { return "boom" }

func TestCuarentaCuiPresenter_GameEnd(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	cfg := domain.DefaultCuarentaConfig()
	cfg.TargetScore = 6
	players := make([]*domain.CuarentaPlayer, domain.CuarentaPlayerCnt)
	for i := 0; i < domain.CuarentaPlayerCnt; i++ {
		players[i] = domain.NewCuarentaPlayer(false)
	}
	g := domain.NewCuarenta(domain.NewTrumpCardsScopa(), players, cfg)
	g.Reset()
	// run to completion to exercise the game-end branch.
	for i := 0; i < 100000 && !g.GetGameEndFlag(); i++ {
		g.CpuPlay()
		if !g.GetGameEndFlag() && g.GetPhase() != int(domain.CuarentaPhasePlay) {
			g.NextRound()
		}
	}
	out := p.Output(g, nil)
	if !strings.Contains(out, "ゲーム終了") {
		t.Errorf("expected game-end text, got: %s", out)
	}
}

func TestCuarentaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	if out := p.ActionLogOutput(g); out == "" {
		t.Error("expected non-empty action log output")
	}
}

// **ボーナスまでの残りをチーム単位で出す。**プレイヤー単位の捕獲数しか出して
// おらず、2 人分を毎回自分で合計させていた (#4893)。
func TestCuarentaCuiPresenter_ShowsTheTeamCaptureTotals(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // 強調の有無まで見る
	defer color.SetNoColor(orig)

	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)

	card := func() *domain.Card { return domain.NewCard(domain.CardDesignSpade, 2, false) }
	give := func(seat, n int) {
		cards := make([]*domain.Card, n)
		for i := range cards {
			cards[i] = card()
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	// **チーム単位で合算する。**席 0 と 2 が同じチーム。
	give(0, 7)
	give(2, 5)
	out := p.Output(g, nil)
	if !strings.Contains(out, "チーム0 捕獲: 12枚") {
		t.Fatalf("team total should be the sum of both seats: %q", out)
	}
	// **必要枚数は「閾値を超える」枚数。**20 ちょうどでは付かない。
	if !strings.Contains(out, strconv.Itoa(domain.CuarentaMostCardsThreshold+1)+"枚以上で") {
		t.Fatalf("the stated requirement must be one above the threshold: %q", out)
	}
	// まだ遠いので強調しない。
	// 色は行全体を包むので、行ごと組み立てて照合する。
	line := func(team, count int) string {
		return i18n.Tf("cuarenta.teamCaptured",
			"team", strconv.Itoa(team), "count", strconv.Itoa(count),
			"need", strconv.Itoa(domain.CuarentaMostCardsThreshold+1),
			"bonus", strconv.Itoa(domain.CuarentaScoreMostCards))
	}
	if strings.Contains(out, color.Yellow(line(0, 12))) {
		t.Fatal("12 captured is not close to the bonus")
	}

	// Web と同じく閾値の 1 つ手前から強調する。
	give(0, domain.CuarentaMostCardsThreshold-1-12)
	near := p.Output(g, nil)
	if !strings.Contains(near, color.Yellow(line(0, domain.CuarentaMostCardsThreshold-1))) {
		t.Fatalf("one card short of the bonus should be highlighted: %q", near)
	}
}
