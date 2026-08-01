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

func setupSpoilFiveWebMock() *interfaces.MockSpoilFiveGame {
	m := new(interfaces.MockSpoilFiveGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetPot").Return(5)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultSpoilFiveConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupSpoilFiveWebMockWithPlayers() (*interfaces.MockSpoilFiveGame, []*domain.SpoilFivePlayer) {
	m := setupSpoilFiveWebMock()
	players := makeSpoilFivePlayers()
	m.On("GetPlayerCnt").Return(5)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestSpoilFiveWebPresenter_Output(t *testing.T) {
	p := new(presenter.SpoilFiveWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupSpoilFiveWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 5)
		assert.Equal(t, int(domain.SpoilFivePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, 5, resObj.Pot)
		assert.Equal(t, "spoilfive.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SpoilFiveCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 30, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SpoilFivePhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "spoilfive.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpoilFivePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "spoilfive.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetPhase").Return(domain.SpoilFivePhaseRoundEnd)
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "spoilfive.roundEnd", resObj.MessageCode)
	})

	t.Run("spoil message code", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetPhase").Return(domain.SpoilFivePhaseRoundEnd)
		m.On("GetRoundWinnerIdx").Return(-1)
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "spoilfive.spoil", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "spoilfive.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "spoilfive.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("score and round tricks propagated", func(t *testing.T) {
		m, players := setupSpoilFiveWebMockWithPlayers()
		players[2].SetScore(12)
		players[2].IncRoundTricks()
		result := p.Output(m, nil)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 12, resObj.Players[2].Score)
		assert.Equal(t, 1, resObj.Players[2].RoundTricks)
	})
}

func TestSpoilFiveWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SpoilFiveWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.SpoilFiveHint{CardIndices: []int{2}, Reason: "take_trick"})
		result := p.HintOutput(m)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "take_trick", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSpoilFiveWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.SpoilFiveHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SpoilFiveWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSpoilFiveWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoilFiveWebPresenter)
	m := new(interfaces.MockSpoilFiveGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。SpoilFive.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestSpoilFiveWebPresenterOutputCarriesTheHint(t *testing.T) {
	sfg, _ := setupSpoilFiveWebMockWithPlayers()
	sfg.ExpectedCalls = removeMockCall(sfg.ExpectedCalls, "GetHint")
	sfg.On("GetHint").Return(&domain.SpoilFiveHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.SpoilFiveWebPresenter).Output(sfg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
