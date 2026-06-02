//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupOsmosisCuiMockDefaults(og *interfaces.MockOsmosisGame) {
	og.On("GetPhase").Return(domain.OsmosisPhasePlaying).Maybe()
	og.On("GetMoveCount").Return(0).Maybe()
	og.On("GetStockCount").Return(34).Maybe()
	og.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	og.On("GetBaseRank").Return(7).Maybe()

	var reserve [domain.OsmosisReserveCnt][]*domain.Card
	for i := 0; i < domain.OsmosisReserveCnt; i++ {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	og.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.OsmosisFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	og.On("GetFoundation").Return(foundation).Maybe()
}

func TestOsmosisCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "オズモシス")
		assert.Contains(t, result, "ベースランク: 7")
		assert.Contains(t, result, "組札0段:")
		assert.Contains(t, result, "リザーブ0列:")
		assert.Contains(t, result, "ストック: 34枚")
		assert.Contains(t, result, "ウェイスト: [空]")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("waste card", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetWaste")
		og.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ウェイスト: HEART 5")
	})

	t.Run("empty foundation row", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetFoundation")
		var empty [domain.OsmosisFoundationCnt][]*domain.Card
		og.On("GetFoundation").Return(empty)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty reserve column", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetReserve")
		var empty [domain.OsmosisReserveCnt][]*domain.Card
		og.On("GetReserve").Return(empty)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("error", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameOver)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestOsmosisCuiPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return((*domain.OsmosisHint)(nil))
		p := new(OsmosisCuiPresenter)
		assert.Contains(t, p.HintOutput(og), "ヒントはありません")
	})

	t.Run("reserve to foundation", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "reserve", FromCol: 0, ToCol: 1})
		p := new(OsmosisCuiPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, "リザーブ0列")
		assert.Contains(t, result, "組札1段")
	})

	t.Run("waste to foundation", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "waste", FromCol: -1, ToCol: 0})
		p := new(OsmosisCuiPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, "ウェイスト")
	})
}

func TestOsmosisCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhasePlaying)
		p := new(OsmosisCuiPresenter)
		assert.Contains(t, p.ActionLogOutput(og), "棋譜はありません")
	})

	t.Run("after clear", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		og.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "test"}})
		p := new(OsmosisCuiPresenter)
		result := p.ActionLogOutput(og)
		assert.Contains(t, result, "draw")
	})
}
