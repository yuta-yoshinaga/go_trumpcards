//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeGanjifaPlayers() []*domain.GanjifaPlayer {
	return []*domain.GanjifaPlayer{
		domain.NewGanjifaPlayer(true),
		domain.NewGanjifaPlayer(false),
		domain.NewGanjifaPlayer(false),
	}
}

func setupGanjifaWebMock() *interfaces.MockGanjifaGame {
	m := new(interfaces.MockGanjifaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GanjifaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetPlayerScores").Return([domain.GanjifaPlayerCnt]int{0, 0, 0})
	m.On("GetRoundTricks").Return([domain.GanjifaPlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultGanjifaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupGanjifaWebMockWithPlayers() (*interfaces.MockGanjifaGame, []*domain.GanjifaPlayer) {
	m := setupGanjifaWebMock()
	players := makeGanjifaPlayers()
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestGanjifaWebPresenter_Output(t *testing.T) {
	p := new(presenter.GanjifaWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupGanjifaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 12, false))
		players[1].AddCard(domain.NewCard(5, 1, false))

		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.GanjifaPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, "ganjifa.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, 1, resObj.TrumpSuit)
	})

	// The 8-suit deck has no PNG art, so every card must arrive with a procedural
	// descriptor. Without one the frontend falls back to the standard 52-card
	// path and renders designs 5-8 as jokers.
	t.Run("cards carry the procedural face descriptor", func(t *testing.T) {
		m, players := setupGanjifaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(2, 7, false))  // strong group
		players[0].AddCard(domain.NewCard(6, 11, false)) // weak group

		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		strong, weak := resObj.Players[0].Cards[0], resObj.Players[0].Cards[1]
		assert.Equal(t, "ganjifa", strong.Deck)
		assert.Equal(t, "ganjifa", weak.Deck)
		assert.Equal(t, domain.GanjifaSuitGlyph(2), strong.Glyph)
		assert.Equal(t, domain.GanjifaSuitGlyph(6), weak.Glyph)
		assert.NotEqual(t, strong.Glyph, weak.Glyph)
		assert.Equal(t, "7", strong.Label)
		assert.Equal(t, "11", weak.Label)
		// Ink colour encodes the group, which is what tells the player which way
		// the ranks read; both suits sharing a colour would erase that.
		assert.Equal(t, "black", strong.Color)
		assert.Equal(t, "blue", weak.Color)
	})

	t.Run("trick cards carry the descriptor too", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(8, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "ganjifa", resObj.CurrentTrick[0].Card.Deck)
		assert.Equal(t, domain.GanjifaSuitGlyph(8), resObj.CurrentTrick[0].Card.Glyph)
		assert.Equal(t, "ganjifa.playPhase.follow", resObj.MessageCode)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.GanjifaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 3, resObj.Config.TargetRounds)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GanjifaPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ganjifa.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GanjifaPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ganjifa.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "ganjifa.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ganjifa.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.GanjifaPlayerCnt]int{4, 2, 0})
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, resObj.Players[0].Score)
		assert.Equal(t, 2, resObj.Players[1].Score)
	})

	t.Run("no playable indices outside the human play turn", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		result := p.Output(m, nil)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.PlayableIndices)
	})
}

func TestGanjifaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GanjifaWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GanjifaHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGanjifaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.GanjifaHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.GanjifaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestGanjifaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GanjifaWebPresenter)
	m := new(interfaces.MockGanjifaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays Taj 12"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGanjifaWebPresenterOutputCarriesTheHint(t *testing.T) {
	g, _ := setupGanjifaWebMockWithPlayers()
	g.ExpectedCalls = removeMockCall(g.ExpectedCalls, "GetHint")
	g.On("GetHint").Return(&domain.GanjifaHint{CardIndices: []int{0}, Reason: "follow_win"})

	result := new(presenter.GanjifaWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "ganjifa.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestGanjifaWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g, _ := setupGanjifaWebMockWithPlayers()
	g.ExpectedCalls = removeMockCall(g.ExpectedCalls, "GetHint")
	g.On("GetHint").Return(&domain.GanjifaHint{CardIndices: []int{0}, Reason: "follow_win"})
	assert.Contains(t, new(presenter.GanjifaWebPresenter).HintOutput(g), "ganjifa.hintRequested")

	none, _ := setupGanjifaWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.GanjifaHint)(nil))
	assert.Contains(t, new(presenter.GanjifaWebPresenter).HintOutput(none), "ganjifa.noHint")
}
