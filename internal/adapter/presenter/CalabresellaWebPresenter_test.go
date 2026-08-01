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

func setupCalabresellaWebMock() *interfaces.MockCalabresellaGame {
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetSoloistIdx").Return(0)
	m.On("GetWinningBid").Return(domain.CalabresellaBidChiamo)
	m.On("GetPlayerScores").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetRoundThirds").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultCalabresellaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupCalabresellaWebMockWithPlayers() (*interfaces.MockCalabresellaGame, []*domain.CalabresellaPlayer) {
	m := setupCalabresellaWebMock()
	players := makeCalabresellaPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestCalabresellaWebPresenter_Output(t *testing.T) {
	p := new(presenter.CalabresellaWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupCalabresellaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.CalabresellaPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, -1, resObj.LastTrickWinner)
		assert.Equal(t, "calabresella.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsSoloist)
		assert.False(t, resObj.Players[1].IsSoloist)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, 0, resObj.SoloistIdx)
		assert.Equal(t, 1, resObj.ForehandIdx)
		assert.Equal(t, int(domain.CalabresellaBidChiamo), resObj.WinningBid)
		assert.True(t, resObj.IsHumanTurn)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.CalabresellaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.CalabresellaWinTarget, resObj.Config.TargetPoints)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseBid)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "calabresella.bidPhase", resObj.MessageCode)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "calabresella.discardPhase", resObj.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.CalabresellaPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "calabresella.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "calabresella.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "calabresella.roundEnd", resObj.MessageCode)
	})

	t.Run("monte revealed from action log after discard", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{PlayerIdx: 0, ActionType: "monte_take", Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 3, false),
				domain.NewCard(domain.CardDesignSpade, 11, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
			}},
		})
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Monte, 4)
		assert.Equal(t, 3, resObj.Monte[0].Value)
		assert.Equal(t, 2, resObj.Monte[3].Value)
	})

	t.Run("monte hidden during bid phase even if log has entry", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetPhase").Return(domain.CalabresellaPhaseBid)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{PlayerIdx: 0, ActionType: "monte_take", Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 3, false),
			}},
		})
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.Monte)
	})

	t.Run("monte empty when no take logged", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.Monte)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "calabresella.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "calabresella.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.CalabresellaPlayerCnt]int{7, 3, 0})
		result := p.Output(m, nil)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 7, resObj.Players[0].Score)
		assert.Equal(t, 3, resObj.Players[1].Score)
		assert.Equal(t, 0, resObj.Players[2].Score)
	})
}

func TestCalabresellaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CalabresellaWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCalabresellaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.CalabresellaHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.CalabresellaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestCalabresellaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CalabresellaWebPresenter)
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCalabresellaWebPresenterOutputCarriesTheHint(t *testing.T) {
	cbg, _ := setupCalabresellaWebMockWithPlayers()
	cbg.ExpectedCalls = removeMockCall(cbg.ExpectedCalls, "GetHint")
	cbg.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.CalabresellaWebPresenter).Output(cbg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "calabresella.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestCalabresellaWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	cbg, _ := setupCalabresellaWebMockWithPlayers()
	cbg.ExpectedCalls = removeMockCall(cbg.ExpectedCalls, "GetHint")
	cbg.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.CalabresellaWebPresenter).HintOutput(cbg), "calabresella.hintRequested")

	none, _ := setupCalabresellaWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.CalabresellaHint)(nil))
	assert.Contains(t, new(presenter.CalabresellaWebPresenter).HintOutput(none), "calabresella.noHint")
}
