//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// **押せない操作をそもそも見せない、が Web 側の設計。** CanUndo() はどのソリティアの
// インタフェースにも在るのに、CUI presenter は 1 つも読んでいなかった (#5830)。
//
// 各ゲームで両方向を見る:
//   - 配った直後 → 「戻せる手はありません」だけ
//   - 1 手打ったあと → 「u で 1 手戻せます」だけ
//
// **どちらも先に require でドメインの CanUndo() を確かめる。** そうしないと
// 「たまたま片方だけ出ている」状態を検査するだけになる。
func TestCuiSolitaireUndoHint_IsShownByEveryGame(t *testing.T) {
	type game struct {
		name string
		// render は現在の盤面を CUI 出力にする。
		render func() string
		// canUndo はドメインの答え。
		canUndo func() bool
		// advance は「必ず打てる 1 手」。無い game は nil。
		advance func() error
	}

	games := []game{}

	add := func(name string, render func() string, canUndo func() bool, advance func() error) {
		games = append(games, game{name, render, canUndo, advance})
	}

	{
		g := domain.NewDefaultAccordion()
		g.Reset()
		p := new(presenter.AccordionCuiPresenter)
		add("accordion", func() string { return p.Output(g, nil) }, g.CanUndo, nil)
	}
	{
		g := domain.NewDefaultAcesUp()
		g.Reset()
		p := new(presenter.AcesUpCuiPresenter)
		add("acesup", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultAgnes()
		g.Reset()
		p := new(presenter.AgnesCuiPresenter)
		add("agnes", func() string { return p.Output(g, nil) }, g.CanUndo, g.DealStock)
	}
	{
		g := domain.NewDefaultBlackHole()
		g.Reset()
		p := new(presenter.BlackHoleCuiPresenter)
		add("blackhole", func() string { return p.Output(g, nil) }, g.CanUndo, nil)
	}
	{
		g := domain.NewDefaultBristol()
		g.Reset()
		p := new(presenter.BristolCuiPresenter)
		add("bristol", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultCanfield()
		g.Reset()
		p := new(presenter.CanfieldCuiPresenter)
		add("canfield", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultFourSeasons()
		g.Reset()
		p := new(presenter.FourSeasonsCuiPresenter)
		add("fourseasons", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultGaps()
		g.Reset()
		p := new(presenter.GapsCuiPresenter)
		add("gaps", func() string { return p.Output(g, nil) }, g.CanUndo, g.Redeal)
	}
	{
		g := domain.NewDefaultGolf()
		g.Reset()
		p := new(presenter.GolfCuiPresenter)
		add("golf", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		p := new(presenter.LaBelleLucieCuiPresenter)
		add("labellelucie", func() string { return p.Output(g, nil) }, g.CanUndo, g.Redeal)
	}
	{
		g := domain.NewDefaultOsmosis()
		g.Reset()
		p := new(presenter.OsmosisCuiPresenter)
		add("osmosis", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultPyramid()
		g.Reset()
		p := new(presenter.PyramidCuiPresenter)
		add("pyramid", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}
	{
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		p := new(presenter.SimpleSimonCuiPresenter)
		add("simplesimon", func() string { return p.Output(g, nil) }, g.CanUndo, nil)
	}
	{
		g := domain.NewDefaultTriPeaks()
		g.Reset()
		p := new(presenter.TriPeaksCuiPresenter)
		add("tripeaks", func() string { return p.Output(g, nil) }, g.CanUndo, g.Draw)
	}

	// 対象を数えておく。ここが減ったら、どこかの game が表から漏れている。
	require.Len(t, games, 14, "#5830 の対象は 14 game")

	available := i18n.T("cuiSolitaireUndoAvailable")
	unavailable := i18n.T("cuiSolitaireUndoUnavailable")
	require.NotEqual(t, available, unavailable, "2 つの文言が同じでは何も区別できない")

	for _, g := range games {
		t.Run(g.name+"/fresh deal has nothing to undo", func(t *testing.T) {
			require.False(t, g.canUndo(), "配った直後は戻せないはず")
			out := g.render()
			assert.Contains(t, out, unavailable)
			assert.NotContains(t, out, available)
		})
	}

	for _, g := range games {
		if g.advance == nil {
			continue
		}
		t.Run(g.name+"/after a move it can be undone", func(t *testing.T) {
			require.NoError(t, g.advance())
			require.True(t, g.canUndo(), "1 手打てば戻せるはず")
			out := g.render()
			assert.Contains(t, out, available)
			assert.NotContains(t, out, unavailable)
		})
	}
}
