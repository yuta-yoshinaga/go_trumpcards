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

func setupGermanSoloWebMock() *interfaces.MockGermanSoloGame {
	m := new(interfaces.MockGermanSoloGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetWinningBid").Return(domain.GermanSoloBidFrage)
	m.On("GetHighestBid").Return(domain.GermanSoloBidNone)
	m.On("IsPlayingAlone").Return(false)
	m.On("GetCalledAceSuit").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCallableAceSuits").Return([]int(nil))
	m.On("IsHumanAceCallTurn").Return(false)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetOutcome").Return(domain.GermanSoloOutcomeNone)
	m.On("GetResult").Return(domain.GermanSoloResultNone)
	m.On("GetPlayerScores").Return([domain.GermanSoloPlayerCnt]int{})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultGermanSoloConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetBiddableBids").Return([]int{int(domain.GermanSoloBidFrage), int(domain.GermanSoloBidSolo), int(domain.GermanSoloBidTout)})
	m.On("RequiredTricks").Return(domain.GermanSoloMakeTricks)
	m.On("GetSideTrickCounts").Return(2, 3)
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupGermanSoloWebMockWithPlayers() (*interfaces.MockGermanSoloGame, []*domain.GermanSoloPlayer) {
	m := setupGermanSoloWebMock()
	players := makeGermanSoloPlayers()
	m.On("GetPlayerCnt").Return(domain.GermanSoloPlayerCnt)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGermanSoloWebPresenter_Output(t *testing.T) {
	p := new(presenter.GermanSoloWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupGermanSoloWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, domain.GermanSoloPlayerCnt)
		assert.Equal(t, int(domain.GermanSoloPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, -1, resObj.LastTrickWinner)
		assert.Equal(t, "germansolo.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, 0, resObj.DeclarerIdx)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.Equal(t, 1, resObj.ForehandIdx)
		assert.Equal(t, int(domain.GermanSoloBidFrage), resObj.WinningBid)
		assert.True(t, resObj.IsHumanTurn)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.GermanSoloCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.GermanSoloWinRounds, resObj.Config.TargetRounds)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GermanSoloPhaseBid)
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "germansolo.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.GermanSoloPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "germansolo.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GermanSoloPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "germansolo.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code by outcome", func(t *testing.T) {
		cases := []struct {
			outcome domain.GermanSoloOutcome
			code    string
		}{
			{domain.GermanSoloOutcomeMade, "germansolo.roundEnd.made"},
			{domain.GermanSoloOutcomeFailed, "germansolo.roundEnd.failed"},
			{domain.GermanSoloOutcomeNone, "germansolo.roundEnd"},
		}
		for _, c := range cases {
			m, _ := setupGermanSoloWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetOutcome")
			m.On("GetPhase").Return(domain.GermanSoloPhaseRoundEnd)
			m.On("GetOutcome").Return(c.outcome)
			result := p.Output(m, nil)
			var resObj controller.GermanSoloWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, c.code, resObj.MessageCode)
			assert.Equal(t, int(c.outcome), resObj.Outcome)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "germansolo.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "germansolo.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.GermanSoloPlayerCnt]int{7, 3, 0})
		result := p.Output(m, nil)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 7, resObj.Players[0].Score)
		assert.Equal(t, 3, resObj.Players[1].Score)
		assert.Equal(t, 0, resObj.Players[2].Score)
	})
}

func TestGermanSoloWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GermanSoloWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGermanSoloWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.GermanSoloHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.GermanSoloWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestGermanSoloWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GermanSoloWebPresenter)
	m := new(interfaces.MockGermanSoloGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGermanSoloWebPresenterOutputCarriesTheHint(t *testing.T) {
	obg, _ := setupGermanSoloWebMockWithPlayers()
	obg.ExpectedCalls = removeMockCall(obg.ExpectedCalls, "GetHint")
	obg.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.GermanSoloWebPresenter).Output(obg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "germansolo.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestGermanSoloWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	obg, _ := setupGermanSoloWebMockWithPlayers()
	obg.ExpectedCalls = removeMockCall(obg.ExpectedCalls, "GetHint")
	obg.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.GermanSoloWebPresenter).HintOutput(obg), "germansolo.hintRequested")

	none, _ := setupGermanSoloWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.GermanSoloHint)(nil))
	assert.Contains(t, new(presenter.GermanSoloWebPresenter).HintOutput(none), "germansolo.noHint")
}
