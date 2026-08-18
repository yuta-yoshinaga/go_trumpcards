//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBristolWebMockDefaults(bg *interfaces.MockBristolGame) {
	bg.On("GetPhase").Return(domain.BristolPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("GetStockCount").Return(28).Maybe()
	bg.On("CanUndo").Return(false).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()
	bg.On("LegalTargets", mock.Anything, mock.Anything).Return(([]int)(nil), ([]int)(nil)).Maybe()

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

// setupBristolOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと
// 先に登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupBristolOutputMock(g *interfaces.MockBristolGame) {
	setupBristolWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestBristolWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolOutputMock(bg)
		p := new(BristolWebPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, `"stockCount":28`)
		assert.Contains(t, result, `"messageCode":"bristol.playing"`)
		assert.Contains(t, result, `"tableau"`)
		assert.Contains(t, result, `"fan"`)
		assert.Contains(t, result, `"foundation"`)
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolOutputMock(bg)
		p := new(BristolWebPresenter)
		result := p.Output(bg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BristolPhaseGameClear)
		p := new(BristolWebPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "bristol.gameClear")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolOutputMock(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BristolPhaseGameOver)
		p := new(BristolWebPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "bristol.gameOver")
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBristolWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.BristolHint{FromZone: "tableau", FromCol: 2, ToZone: "foundation", ToCol: 0}

	bg := new(interfaces.MockBristolGame)
	setupBristolWebMockDefaults(bg)
	bg.On("GetHint").Return(hint).Maybe()

	result := new(BristolWebPresenter).Output(bg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **押すまで合法か分からなかった (#4813)。**選択中は全ての移動先が同じ見た目で
// 強調されていた。移動元ごとの合法な移動先を出力に載せる。
func TestBristolWebPresenter_LegalTargets(t *testing.T) {
	bg := new(interfaces.MockBristolGame)
	setupBristolWebMockDefaults(bg)
	bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "LegalTargets")
	bg.On("LegalTargets", "tableau", 0).Return([]int{3}, []int{1})
	bg.On("LegalTargets", mock.Anything, mock.Anything).Return(([]int)(nil), ([]int)(nil))
	bg.On("GetHint").Return((*domain.BristolHint)(nil)).Maybe()

	var out controller.BristolWebOutput
	assert.NoError(t, json.Unmarshal([]byte(new(BristolWebPresenter).Output(bg, nil)), &out))

	// 合法手のある移動元だけがキーになる。
	assert.Equal(t, []int{3}, out.LegalTargets["tableau-0"].Tableau)
	assert.Equal(t, []int{1}, out.LegalTargets["tableau-0"].Foundation)
	assert.NotContains(t, out.LegalTargets, "tableau-1", "動かせない移動元はキーごと出さない")
	assert.NotContains(t, out.LegalTargets, "fan-0")
}

func TestBristolWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolWebMockDefaults(bg)
		bg.On("GetHint").Return(&domain.BristolHint{FromZone: "tableau", FromCol: 0, ToZone: "foundation", ToCol: 1})
		p := new(BristolWebPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, `"bristol.hintAvailable"`)
		assert.Contains(t, result, `"toZone":"foundation"`)
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		setupBristolWebMockDefaults(bg)
		bg.On("GetHint").Return((*domain.BristolHint)(nil))
		p := new(BristolWebPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, `"bristol.noHint"`)
	})
}

func TestBristolWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetPhase").Return(domain.BristolPhasePlaying)
		bg.On("GetGameEndFlag").Return(false)
		p := new(BristolWebPresenter)
		_ = p.ActionLogOutput(bg)
	})

	t.Run("cleared", func(t *testing.T) {
		bg := new(interfaces.MockBristolGame)
		bg.On("GetPhase").Return(domain.BristolPhaseGameClear)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw"}})
		p := new(BristolWebPresenter)
		_ = p.ActionLogOutput(bg)
	})
}

// #5631: 手詰まりと脱出手数をレスポンスに載せる (画面が脱出ボタンを出せるように)。
func TestBristolWebPresenterCarriesTheStalemate(t *testing.T) {
	bg := new(interfaces.MockBristolGame)
	// Output は受動ヒントも埋めるので、GetHint を含む Output 用の既定を使う。
	setupBristolOutputMock(bg)
	bg.ExpectedCalls = bristolMockWithout(bg.ExpectedCalls, "IsStalemate", "UndoToEscape")
	bg.On("IsStalemate").Return(true)
	bg.On("UndoToEscape").Return(2)

	var out controller.BristolWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(BristolWebPresenter).Output(bg, nil)), &out))
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 2, out.UndoToEscape)
}

// bristolMockWithout drops the listed expectations so a test can override them.
func bristolMockWithout(calls []*mock.Call, methods ...string) []*mock.Call {
	drop := make(map[string]bool, len(methods))
	for _, m := range methods {
		drop[m] = true
	}
	// **新しいスライスに詰め直す。**`calls[:0]` で書き戻すと、呼び出し側が
	// 元のスライスを保持している場合に中身が壊れる。
	out := make([]*mock.Call, 0, len(calls))
	for _, c := range calls {
		if !drop[c.Method] {
			out = append(out, c)
		}
	}
	return out
}
