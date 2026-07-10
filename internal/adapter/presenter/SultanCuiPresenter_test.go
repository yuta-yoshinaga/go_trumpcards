//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSultanCuiMockDefaults(sg *interfaces.MockSultanGame) {
	sg.On("GetPhase").Return(domain.SultanPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(88).Maybe()
	sg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("GetRedealCount").Return(0).Maybe()

	divan := make([]*domain.Card, domain.SultanDivanCnt)
	for i := range divan {
		divan[i] = domain.NewCard(domain.CardDesignSpade, i+1, false)
	}
	sg.On("GetDivan").Return(divan).Maybe()

	var foundation [domain.SultanFoundationCnt][]*domain.Card
	for i := range domain.SultanFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, false)}
	}
	sg.On("GetFoundation").Return(foundation).Maybe()
}

func TestSultanCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		p := new(SultanCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Sultan of Turkey")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "Divan")
		assert.Contains(t, result, "Stock: 88枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "リディール残: 2回")
		assert.Contains(t, result, "手数: 0")
		assert.Contains(t, result, "操作: m で移動")
	})

	t.Run("with waste card", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetWaste")
		sg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(SultanCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "Waste:")
		assert.NotContains(t, result, "Waste: [空]")
	})

	t.Run("nil divan slot", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetDivan")
		divan := make([]*domain.Card, domain.SultanDivanCnt) // all nil
		sg.On("GetDivan").Return(divan)

		p := new(SultanCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		p := new(SultanCuiPresenter)

		result := p.Output(sg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SultanPhaseGameClear)

		p := new(SultanCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SultanPhaseGameOver)

		p := new(SultanCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)

		p := new(SultanCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まり")
	})
}

func TestSultanCuiPresenter_HintOutput(t *testing.T) {
	t.Run("divan hint", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetHint").Return(&domain.SultanHint{FromZone: "divan", FromIdx: 2, ToFoundation: 0})

		p := new(SultanCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ディヴァン")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("waste hint", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetHint").Return(&domain.SultanHint{FromZone: "waste", FromIdx: -1, ToFoundation: 1})

		p := new(SultanCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ウェイスト")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetHint").Return((*domain.SultanHint)(nil))

		p := new(SultanCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestSultanCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetPhase").Return(domain.SultanPhasePlaying)

		p := new(SultanCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetPhase").Return(domain.SultanPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(SultanCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "draw")
	})
}
