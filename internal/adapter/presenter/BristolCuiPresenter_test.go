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

func setupBristolCuiMockDefaults(bg *interfaces.MockBristolGame) {
	bg.On("GetPhase").Return(domain.BristolPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	// 手詰まりは既定で無し (#5631)。検知そのものを見るテストは自分で上書きする。
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()
	bg.On("GetStockCount").Return(28).Maybe()

	var tableau [domain.BristolTableauCnt][]*domain.Card
	for i := 0; i < domain.BristolTableauCnt; i++ {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var fan [domain.BristolFanCnt][]*domain.Card
	fan[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)}
	bg.On("GetFan").Return(fan).Maybe()

	var foundation [domain.BristolFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestBristolCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "Bristol")
		assert.Contains(t, result, "組札0:")
		assert.Contains(t, result, "降順ビルド") // tableau build-down rule line
		assert.Contains(t, result, "タブロー0列:")
		assert.Contains(t, result, "ファン0:")
		assert.Contains(t, result, "ストック: 28枚")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var empty [domain.BristolTableauCnt][]*domain.Card
		bg.On("GetTableau").Return(empty)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty fan", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFan")
		var empty [domain.BristolFanCnt][]*domain.Card
		bg.On("GetFan").Return(empty)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("error", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BristolPhaseGameClear)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BristolPhaseGameOver)
		p := new(BristolCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestBristolCuiPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetHint").Return((*domain.BristolHint)(nil))
		p := new(BristolCuiPresenter)
		assert.Contains(t, p.HintOutput(bg), "ヒントはありません")
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetHint").Return(&domain.BristolHint{FromZone: "tableau", FromCol: 0, ToZone: "foundation", ToCol: 1})
		p := new(BristolCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー0列")
		assert.Contains(t, result, "組札1")
	})

	t.Run("fan to tableau", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetHint").Return(&domain.BristolHint{FromZone: "fan", FromCol: 2, ToZone: "tableau", ToCol: 3})
		p := new(BristolCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ファン2")
		assert.Contains(t, result, "タブロー3列")
	})
}

func TestBristolCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetPhase").Return(domain.BristolPhasePlaying)
		p := new(BristolCuiPresenter)
		assert.Contains(t, p.ActionLogOutput(bg), "棋譜はありません")
	})

	t.Run("after clear", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetPhase").Return(domain.BristolPhaseGameClear)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "test"}})
		p := new(BristolCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "draw")
	})
}

// #5631: 手詰まりを検知しても画面に出さなければ、プレイヤーは動かせる札を
// 探し続けることになる。
func TestBristolCuiPresenterWarnsOnStalemate(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	bg := new(interfaces.MockBristolGame)
	setupBristolCuiMockDefaults(bg)
	bg.ExpectedCalls = bristolWithout(bg.ExpectedCalls, "IsStalemate", "UndoToEscape")
	bg.On("IsStalemate").Return(true)
	bg.On("UndoToEscape").Return(3)

	out := new(BristolCuiPresenter).Output(bg, nil)
	assert.Contains(t, out, i18n.T("cuiSolitaireStalemate"))
	// 何回戻せば打てるようになるかまで言う (Web の脱出ボタンと同じ情報)。
	assert.Contains(t, out, i18n.Tf("bristol.undoToEscape", "count", "3"))
}

// 手詰まりでなければ何も出さない。
func TestBristolCuiPresenterSaysNothingWhenPlayable(t *testing.T) {
	bg := new(interfaces.MockBristolGame)
	setupBristolCuiMockDefaults(bg)

	out := new(BristolCuiPresenter).Output(bg, nil)
	assert.NotContains(t, out, i18n.T("cuiSolitaireStalemate"))
}

// bristolWithout drops the listed expectations so a test can override them.
// testify は最初に一致した期待値を返すので、既定 (.Maybe()) を消さずに足すと
// 上書きしたつもりのケースが何も確かめない。
func bristolWithout(calls []*mock.Call, methods ...string) []*mock.Call {
	drop := make(map[string]bool, len(methods))
	for _, m := range methods {
		drop[m] = true
	}
	out := calls[:0]
	for _, c := range calls {
		if !drop[c.Method] {
			out = append(out, c)
		}
	}
	return out
}
