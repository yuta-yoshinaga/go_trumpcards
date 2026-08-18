//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupEasthavenCuiMockDefaults(eg *interfaces.MockEasthavenGame) {
	eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
	eg.On("GetMoveCount").Return(0).Maybe()
	eg.On("GetStockCount").Return(31).Maybe()
	eg.On("IsStalemate").Return(false).Maybe()
	eg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.EasthavenTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	eg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.EasthavenFoundationCnt][]*domain.Card
	eg.On("GetFoundation").Return(foundation).Maybe()
}

func TestEasthavenCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(eg)
		p := new(EasthavenCuiPresenter)

		result := p.Output(eg, nil)
		assert.Contains(t, result, "Easthaven")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(eg)
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, assert.AnError), assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(5).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(true).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameClear).Maybe()
		eg.On("GetMoveCount").Return(42).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver).Maybe()
		eg.On("GetMoveCount").Return(10).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "ゲームオーバー")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(0).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "SPADE 1")
	})
}

func TestEasthavenCuiPresenter_HintOutput(t *testing.T) {
	t.Run("to foundation", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return(&domain.EasthavenHint{FromCol: 0, CardIndex: 1, ToZone: "foundation", ToCol: 0})
		p := new(EasthavenCuiPresenter)
		result := p.HintOutput(eg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("to tableau", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return(&domain.EasthavenHint{FromCol: 0, CardIndex: 1, ToZone: "tableau", ToCol: 3})
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.HintOutput(eg), "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return((*domain.EasthavenHint)(nil))
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.HintOutput(eg), "ヒントはありません")
	})
}

func TestEasthavenCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying)
		p := new(EasthavenCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(eg))
	})

	t.Run("game over", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver)
		eg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move", Detail: "test"}})
		p := new(EasthavenCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(eg))
	})
}

// #5634: `Deal()` は空き列が 1 つでもあると拒否する。Web はボタンを揺らして
// 理由を出しているのに、CUI は `deal` を打って初めてその規則を知る形だった。
func TestEasthavenCuiPresenterWarnsWhileAColumnIsEmpty(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(EasthavenCuiPresenter)

	build := func(t *testing.T, stock int, empty bool, faceDownLeft bool) *interfaces.MockEasthavenGame {
		t.Helper()
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(eg)
		eg.ExpectedCalls = easthavenMockWithout(eg.ExpectedCalls, "GetStockCount", "GetTableau")
		eg.On("GetStockCount").Return(stock)

		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		for i := range domain.EasthavenTableauCnt {
			tableau[i] = []*domain.KlondikeTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
			}
		}
		if empty {
			tableau[0] = nil
		}
		if faceDownLeft {
			tableau[1] = []*domain.KlondikeTableauCard{
				{Card: domain.NewCard(domain.CardDesignHeart, 9, false), FaceUp: false},
				{Card: domain.NewCard(domain.CardDesignHeart, 8, false), FaceUp: true},
			}
		}
		eg.On("GetTableau").Return(tableau)
		return eg
	}

	t.Run("says why deal is refused while a column is empty", func(t *testing.T) {
		out := p.Output(build(t, 20, true, false), nil)
		assert.Contains(t, out, i18n.T("easthaven.cannotDealEmptyCol"))
	})

	// 山札が尽きていれば、そもそもめくれない。空き列の話をしても混乱させるだけ。
	t.Run("says nothing about dealing once the stock is gone", func(t *testing.T) {
		out := p.Output(build(t, 0, true, false), nil)
		assert.NotContains(t, out, i18n.T("easthaven.cannotDealEmptyCol"))
	})

	t.Run("warns about face-down cards left with no stock", func(t *testing.T) {
		out := p.Output(build(t, 0, false, true), nil)
		assert.Contains(t, out, i18n.T("easthaven.faceDownLeft"))
	})

	// ギブアップ後は言わない。もう打てないゲームの遊び方を説明することになる
	// (レビュー #5997)。
	t.Run("says nothing once the game is over", func(t *testing.T) {
		eg := build(t, 0, false, true)
		eg.ExpectedCalls = easthavenMockWithout(eg.ExpectedCalls, "GetPhase")
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver)

		out := p.Output(eg, nil)
		assert.NotContains(t, out, i18n.T("easthaven.faceDownLeft"))
		assert.NotContains(t, out, i18n.T("easthaven.cannotDealEmptyCol"))
	})

	// 通常時は余計な行を増やさない。
	t.Run("stays quiet when neither applies", func(t *testing.T) {
		out := p.Output(build(t, 20, false, false), nil)
		assert.NotContains(t, out, i18n.T("easthaven.cannotDealEmptyCol"))
		assert.NotContains(t, out, i18n.T("easthaven.faceDownLeft"))
	})
}

// easthavenMockWithout drops the listed expectations so a test can override them.
func easthavenMockWithout(calls []*mock.Call, methods ...string) []*mock.Call {
	drop := make(map[string]bool, len(methods))
	for _, m := range methods {
		drop[m] = true
	}
	out := make([]*mock.Call, 0, len(calls))
	for _, c := range calls {
		if !drop[c.Method] {
			out = append(out, c)
		}
	}
	return out
}
