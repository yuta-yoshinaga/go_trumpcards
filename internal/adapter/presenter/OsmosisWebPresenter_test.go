//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupOsmosisWebMockDefaults(og *interfaces.MockOsmosisGame) {
	og.On("IsStalemate").Return(false).Maybe()
	og.On("GetPhase").Return(domain.OsmosisPhasePlaying).Maybe()
	og.On("GetMoveCount").Return(0).Maybe()
	og.On("GetStockCount").Return(34).Maybe()
	og.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	og.On("GetBaseRank").Return(7).Maybe()
	og.On("CanUndo").Return(false).Maybe()

	var reserve [domain.OsmosisReserveCnt][]*domain.Card
	for i := 0; i < domain.OsmosisReserveCnt; i++ {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	og.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.OsmosisFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	og.On("GetFoundation").Return(foundation).Maybe()
}

// setupOsmosisOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupOsmosisOutputMock(g *interfaces.MockOsmosisGame) {
	setupOsmosisWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestOsmosisWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisOutputMock(og)
		p := new(OsmosisWebPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, `"baseRank":7`)
		assert.Contains(t, result, `"stockCount":34`)
		assert.Contains(t, result, `"messageCode":"osmosis.playing"`)
		assert.Contains(t, result, `"reserve"`)
		assert.Contains(t, result, `"foundation"`)
	})

	t.Run("with error", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisOutputMock(og)
		p := new(OsmosisWebPresenter)
		result := p.Output(og, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisOutputMock(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		p := new(OsmosisWebPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "osmosis.gameClear")
	})

	t.Run("game over", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisOutputMock(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameOver)
		p := new(OsmosisWebPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "osmosis.gameOver")
	})

	t.Run("with waste", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisOutputMock(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetWaste")
		og.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		p := new(OsmosisWebPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, `"waste"`)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
// **手詰まりでもフェーズは Playing のまま (#4808)。**メッセージコードと
// isStalemate の両方で通知する。両側を踏まないと常時 true の実装でも通る。
func TestOsmosisWebPresenter_Stalemate(t *testing.T) {
	t.Run("reports the dead end", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("IsStalemate").Return(true)
		setupOsmosisWebMockDefaults(og)
		og.On("GetHint").Return((*domain.OsmosisHint)(nil)).Maybe()

		result := new(OsmosisWebPresenter).Output(og, nil)
		assert.Contains(t, result, `"isStalemate":true`)
		assert.Contains(t, result, `"messageCode":"osmosis.stalemate"`)
	})

	t.Run("stays in the plain playing state while a move remains", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisWebMockDefaults(og)
		og.On("GetHint").Return((*domain.OsmosisHint)(nil)).Maybe()

		result := new(OsmosisWebPresenter).Output(og, nil)
		assert.Contains(t, result, `"isStalemate":false`)
		assert.Contains(t, result, `"messageCode":"osmosis.playing"`)
	})
}

func TestOsmosisWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisWebMockDefaults(og)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "reserve", FromCol: 1, ToCol: 2}).Maybe()

		result := new(OsmosisWebPresenter).Output(og, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not once cleared", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisWebMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "reserve", FromCol: 1, ToCol: 2}).Maybe()

		result := new(OsmosisWebPresenter).Output(og, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestOsmosisWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisWebMockDefaults(og)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "waste", FromCol: -1, ToCol: 0})
		p := new(OsmosisWebPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, `"osmosis.hintAvailable"`)
	})

	t.Run("no hint", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisWebMockDefaults(og)
		og.On("GetHint").Return((*domain.OsmosisHint)(nil))
		p := new(OsmosisWebPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, `"osmosis.noHint"`)
	})
}

func TestOsmosisWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhasePlaying)
		og.On("GetGameEndFlag").Return(false)
		p := new(OsmosisWebPresenter)
		_ = p.ActionLogOutput(og)
	})

	t.Run("cleared", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		og.On("GetGameEndFlag").Return(true)
		og.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw"}})
		p := new(OsmosisWebPresenter)
		_ = p.ActionLogOutput(og)
	})
}
