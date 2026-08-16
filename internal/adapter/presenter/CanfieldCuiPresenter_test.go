//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCanfieldCuiMockDefaults(cg *interfaces.MockCanfieldGame) {
	cg.On("GetPhase").Return(domain.CanfieldPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("CanUndo").Return(false).Maybe()
	cg.On("GetStockCount").Return(34).Maybe()
	cg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	cg.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)}).Maybe()
	cg.On("GetBaseRank").Return(7).Maybe()

	var tableau [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	for i := 0; i < domain.CanfieldTableauCnt; i++ {
		tableau[i] = []*domain.CanfieldTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, i+1, false)}}
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CanfieldFoundationCnt][]*domain.Card
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func TestCanfieldCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "Canfield")
		assert.Contains(t, result, "Base rank: 7")
		assert.Contains(t, result, "Reserve")
		assert.Contains(t, result, "Stock: 34枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("waste card", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetWaste")
		cg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "Waste: HEART 5")
	})

	t.Run("error", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameClear)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameOver)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		cg.On("GetTableau").Return(emptyTab)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty reserve", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetReserve")
		cg.On("GetReserve").Return([]*domain.Card{})
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "Reserve: [空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetFoundation")
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		cg.On("GetFoundation").Return(f)
		p := new(CanfieldCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "SPADE 7")
	})
}

func TestCanfieldCuiPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetHint").Return((*domain.CanfieldHint)(nil))
		p := new(CanfieldCuiPresenter)
		assert.Contains(t, p.HintOutput(cg), "ヒントはありません")
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetHint").Return(&domain.CanfieldHint{FromZone: "tableau", FromCol: 0, CardIndex: 2, ToZone: "foundation", ToCol: 0})
		p := new(CanfieldCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("reserve to tableau", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetHint").Return(&domain.CanfieldHint{FromZone: "reserve", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: 2})
		p := new(CanfieldCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "リザーブ")
		assert.Contains(t, result, "タブロー列2")
	})

	t.Run("waste to tableau", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetHint").Return(&domain.CanfieldHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: 1})
		p := new(CanfieldCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "ウェイスト")
	})
}

func TestCanfieldCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetPhase").Return(domain.CanfieldPhasePlaying)
		p := new(CanfieldCuiPresenter)
		assert.Contains(t, p.ActionLogOutput(cg), "棋譜はありません")
	})

	t.Run("after clear", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameClear)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "test"}})
		p := new(CanfieldCuiPresenter)
		result := p.ActionLogOutput(cg)
		assert.Contains(t, result, "draw")
	})
}
