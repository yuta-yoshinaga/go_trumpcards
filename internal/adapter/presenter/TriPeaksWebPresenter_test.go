//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTriPeaksWebMockDefaults(tg *interfaces.MockTriPeaksGame) {
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(0).Maybe()
	tg.On("GetStockCount").Return(23).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()
	tg.On("AllRemoved").Return(false).Maybe()

	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	for r := range domain.TriPeaksRowCnt {
		for c := range domain.TriPeaksColCnt {
			if r == 3 && c < 10 {
				layout[r][c] = &domain.TriPeaksCard{
					Card:    domain.NewCard(domain.CardDesignSpade, c%13+1, false),
					Removed: false,
				}
			}
		}
	}
	tg.On("GetLayout").Return(layout).Maybe()
	for r := range domain.TriPeaksRowCnt {
		for c := range domain.TriPeaksColCnt {
			if r == 3 {
				tg.On("IsExposed", r, c).Return(true).Maybe()
			} else {
				tg.On("IsExposed", r, c).Return(false).Maybe()
			}
		}
	}
}

func parseTriPeaksOutput(t *testing.T, jsonStr string) *controller.TriPeaksWebOutput {
	t.Helper()
	var out controller.TriPeaksWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupTriPeaksOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupTriPeaksOutputMock(g *interfaces.MockTriPeaksGame) {
	setupTriPeaksWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestTriPeaksWebPresenterOutput_Playing(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksOutputMock(tg)
	p := &TriPeaksWebPresenter{}

	result := p.Output(tg, nil)
	out := parseTriPeaksOutput(t, result)

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 23, out.StockCount)
	assert.Equal(t, "tripeaks.playing", out.MessageCode)
	assert.Len(t, out.Layout, domain.TriPeaksRowCnt)
}

func TestTriPeaksWebPresenterOutput_Error(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksOutputMock(tg)
	p := &TriPeaksWebPresenter{}

	result := p.Output(tg, errors.New("test error"))
	out := parseTriPeaksOutput(t, result)

	assert.Equal(t, "test error", out.Message)
}

func TestTriPeaksWebPresenterOutput_Stalemate(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksOutputMock(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("IsStalemate").Return(true).Maybe()
	tg.On("UndoToEscape").Return(-1).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()
	for r := range domain.TriPeaksRowCnt {
		for c := range domain.TriPeaksColCnt {
			tg.On("IsExposed", r, c).Return(false).Maybe()
		}
	}

	p := &TriPeaksWebPresenter{}
	result := p.Output(tg, nil)
	out := parseTriPeaksOutput(t, result)

	assert.Equal(t, "tripeaks.stalemate", out.MessageCode)
}

func TestTriPeaksWebPresenterOutput_GameClear(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksOutputMock(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameClear).Maybe()
	tg.On("GetMoveCount").Return(10).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()
	for r := range domain.TriPeaksRowCnt {
		for c := range domain.TriPeaksColCnt {
			tg.On("IsExposed", r, c).Return(false).Maybe()
		}
	}

	p := &TriPeaksWebPresenter{}
	result := p.Output(tg, nil)
	out := parseTriPeaksOutput(t, result)

	assert.Equal(t, "tripeaks.gameClear", out.MessageCode)
	assert.Contains(t, out.Message, "10")
}

func TestTriPeaksWebPresenterOutput_GameOver(t *testing.T) {
	tg := new(interfaces.MockTriPeaksGame)
	setupTriPeaksOutputMock(tg)
	tg.ExpectedCalls = nil
	tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver).Maybe()
	tg.On("GetMoveCount").Return(5).Maybe()
	tg.On("GetStockCount").Return(0).Maybe()
	tg.On("GetScore").Return(0).Maybe()
	tg.On("GetCombo").Return(0).Maybe()
	tg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	tg.On("CanUndo").Return(false).Maybe()
	tg.On("IsStalemate").Return(false).Maybe()
	tg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	tg.On("GetLayout").Return(layout).Maybe()
	for r := range domain.TriPeaksRowCnt {
		for c := range domain.TriPeaksColCnt {
			tg.On("IsExposed", r, c).Return(false).Maybe()
		}
	}

	p := &TriPeaksWebPresenter{}
	result := p.Output(tg, nil)
	out := parseTriPeaksOutput(t, result)

	assert.Equal(t, "tripeaks.gameOver", out.MessageCode)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestTriPeaksWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		tpg := new(interfaces.MockTriPeaksGame)
		setupTriPeaksWebMockDefaults(tpg)
		tpg.On("GetHint").Return(&domain.TriPeaksHint{Type: "remove", Row: 2, Col: 3}).Maybe()

		result := new(TriPeaksWebPresenter).Output(tpg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		tpg := new(interfaces.MockTriPeaksGame)
		setupTriPeaksWebMockDefaults(tpg)
		tpg.ExpectedCalls = filterCalls(tpg.ExpectedCalls, "IsStalemate")
		tpg.On("IsStalemate").Return(true)
		tpg.On("GetHint").Return(&domain.TriPeaksHint{Type: "remove", Row: 2, Col: 3}).Maybe()

		result := new(TriPeaksWebPresenter).Output(tpg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestTriPeaksWebPresenterHintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return(&domain.TriPeaksHint{Type: "remove", Row: 3, Col: 0})
		tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying)
		tg.On("GetMoveCount").Return(0)
		tg.On("GetStockCount").Return(23)
		tg.On("GetScore").Return(0).Maybe()
		tg.On("GetCombo").Return(0).Maybe()
		tg.On("CanUndo").Return(false)
		tg.On("IsStalemate").Return(false)
		tg.On("UndoToEscape").Return(0)

		p := &TriPeaksWebPresenter{}
		result := p.HintOutput(tg)
		out := parseTriPeaksOutput(t, result)

		assert.NotNil(t, out.Hint)
		assert.Equal(t, "remove", out.Hint.Type)
		assert.Equal(t, "tripeaks.hintAvailable", out.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetHint").Return((*domain.TriPeaksHint)(nil))
		tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying)
		tg.On("GetMoveCount").Return(0)
		tg.On("GetStockCount").Return(0)
		tg.On("GetScore").Return(0).Maybe()
		tg.On("GetCombo").Return(0).Maybe()
		tg.On("CanUndo").Return(false)
		tg.On("IsStalemate").Return(false)
		tg.On("UndoToEscape").Return(0)

		p := &TriPeaksWebPresenter{}
		result := p.HintOutput(tg)
		out := parseTriPeaksOutput(t, result)

		assert.Nil(t, out.Hint)
		assert.Equal(t, "tripeaks.noHint", out.MessageCode)
	})
}

func TestTriPeaksWebPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhasePlaying)

		tg.On("GetGameEndFlag").Return(false)
		p := &TriPeaksWebPresenter{}
		result := p.ActionLogOutput(tg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		tg := new(interfaces.MockTriPeaksGame)
		tg.On("GetPhase").Return(domain.TriPeaksPhaseGameOver)
		tg.On("GetGameEndFlag").Return(true)
		tg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &TriPeaksWebPresenter{}
		result := p.ActionLogOutput(tg)
		assert.Contains(t, result, "entries")
	})
}
