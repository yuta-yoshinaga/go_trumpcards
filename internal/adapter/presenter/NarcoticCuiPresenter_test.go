//go:build test

package presenter

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupNarcoticCuiMockDefaults(g *interfaces.MockNarcoticGame) {
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetDiscardCount").Return(4).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetColumns").Return(sampleNarcoticColumns()).Maybe()
	g.On("CanRemoveSet").Return(false).Maybe()
	g.On("CanMove", mock.AnythingOfType("int")).Return(false).Maybe()
}

func TestNarcoticCuiPresenterOutput_Playing(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticCuiMockDefaults(g)
	p := &NarcoticCuiPresenter{}

	result := p.Output(g, nil)
	assert.Contains(t, result, i18n.T("narcotic.helpTitle"))
	assert.Contains(t, result, "Stock: 44枚")
	assert.Contains(t, result, "Discard: 4枚")
	assert.Contains(t, result, "手数: 0")
	assert.Contains(t, result, "[空]") // col2 is empty
}

func TestNarcoticCuiPresenterOutput_TopCardMarkers(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetDiscardCount").Return(4).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetColumns").Return(sampleNarcoticColumns()).Maybe()
	// **除去の印は盤面全体の性質。**4枚揃ったときだけ付き、そのときは全列に付く
	// (クローン元は列ごとに違った)。移動の印は行き先がある列だけ。
	g.On("CanRemoveSet").Return(true)
	g.On("CanMove", 0).Return(false)
	g.On("CanMove", 1).Return(true)
	g.On("CanMove", mock.AnythingOfType("int")).Return(false).Maybe()

	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := &NarcoticCuiPresenter{}
	result := p.Output(g, nil)
	// The markers attach to the top-card bracket ("]*" / "]>"), distinguishing
	// them from the "* =" / "> =" glyphs in the always-present legend line.
	// **揃っているときは全列に * が付く。**列ごとに違うのはクローン元の性質。
	assert.Contains(t, result, "]*")
	// 列1 だけが重ねられるので、そこだけ > が続く。
	assert.Contains(t, result, "]*>")
	assert.Contains(t, result, i18n.T("narcotic.markerLegend"))
}

// 負のコントロール: 揃っていなければ * はどこにも付かない。
func TestNarcoticCuiPresenterOutput_NoDiscardMarkerWhenRanksDiffer(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticCuiMockDefaults(g)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRemoveSet")
	g.On("CanRemoveSet").Return(false)

	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	result := new(NarcoticCuiPresenter).Output(g, nil)
	assert.NotContains(t, result, "]*")
}

func TestNarcoticCuiPresenterOutput_Error(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticCuiMockDefaults(g)
	p := &NarcoticCuiPresenter{}
	assert.Contains(t, p.Output(g, errors.New("test error")), "test error")
}

func TestNarcoticCuiPresenterOutput_Stalemate(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(10).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "手詰まり")
}

func TestNarcoticCuiPresenterOutput_GameClear(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(20).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(48).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "ゲームクリア")
}

func TestNarcoticCuiPresenterOutput_GameOver(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(2).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticCuiPresenter{}
	assert.Contains(t, p.Output(g, nil), "ゲームオーバー")
}

func TestNarcoticCuiPresenterHintOutput(t *testing.T) {
	// **除去は列を名指ししない。**4枚まとめてなので Col は -1。
	t.Run("remove hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "remove", Col: -1})
		assert.Contains(t, new(NarcoticCuiPresenter).HintOutput(g), i18n.T("narcotic.hintRemove"))
	})

	// **移動先は「空き列」ではなく「左の同ランク」。**クローン元の文面を残すと、
	// 存在しない操作を案内することになる。
	t.Run("move hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "move", Col: 2})
		out := new(NarcoticCuiPresenter).HintOutput(g)
		assert.Contains(t, out, i18n.Tf("narcotic.hintMove", "col", "2"))
		assert.NotContains(t, out, "空き列", "Aces Up の文面が残っていないこと")
	})

	t.Run("redeal hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "redeal", Col: -1})
		assert.Contains(t, new(NarcoticCuiPresenter).HintOutput(g), i18n.T("narcotic.hintRedeal"))
	})

	t.Run("draw hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "draw", Col: -1})
		p := &NarcoticCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "配って")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return((*domain.NarcoticHint)(nil))
		p := &NarcoticCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "ヒントはありません")
	})

	t.Run("unknown hint type", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "unknown"})
		p := &NarcoticCuiPresenter{}
		assert.Contains(t, p.HintOutput(g), "不明")
	})
}

func TestNarcoticCuiPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetPhase").Return(domain.NarcoticPhasePlaying)
		p := &NarcoticCuiPresenter{}
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetPhase").Return(domain.NarcoticPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
		p := &NarcoticCuiPresenter{}
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}

// #5620: CUI のヒントは「何をするか」だけで、**なぜ**が無かった。Web は同じ
// 場面で `hintReason.*` の一文を出している (「同スートでより小さい札」など)。
func TestNarcoticCuiPresenterHintGivesTheReason(t *testing.T) {
	cases := []struct {
		name string
		hint *domain.NarcoticHint
		key  string
	}{
		{"remove", &domain.NarcoticHint{Type: "remove", Col: 0}, "narcotic.hintReasonRemove"},
		{"move", &domain.NarcoticHint{Type: "move", Col: 2}, "narcotic.hintReasonMove"},
		{"draw", &domain.NarcoticHint{Type: "draw", Col: -1}, "narcotic.hintReasonDraw"},
	}

	seen := make(map[string]bool, len(cases))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockNarcoticGame)
			g.On("GetHint").Return(tc.hint)

			out := (&NarcoticCuiPresenter{}).HintOutput(g)
			reason := i18n.Tf(tc.key, "col", strconv.Itoa(tc.hint.Col))
			assert.Contains(t, out, reason)
			// 生のキーが漏れていない (ロケールに無いと Tf はキーをそのまま返す)。
			assert.NotContains(t, out, tc.key)
			seen[reason] = true
		})
	}
	// **3 パターンが別の文になっていること。**同じ文を 3 回出すなら、理由が
	// 付いていないのと変わらない。
	assert.Len(t, seen, len(cases))
}
