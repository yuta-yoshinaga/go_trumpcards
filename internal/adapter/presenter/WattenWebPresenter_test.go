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

func setupWattenWebMock() *interfaces.MockWattenGame {
	m := new(interfaces.MockWattenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WattenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetSchlagRank").Return(10)
	m.On("GetCriticalSuit").Return(1)
	m.On("GetStake").Return(2)
	m.On("GetPendingStake").Return(0)
	m.On("GetRaiseCount").Return(0)
	m.On("GetRaiserTeam").Return(-1)
	m.On("GetResponderIdx").Return(-1)
	m.On("CanHumanRaise").Return(false)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetTeamTricks", 0).Return(0)
	m.On("GetTeamTricks", 1).Return(0)
	m.On("GetDealWinnerTeam").Return(-1)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetResult").Return(domain.WattenResultNone)
	m.On("GetConfig").Return(domain.DefaultWattenConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupWattenWebMockWithPlayers() (*interfaces.MockWattenGame, []*domain.WattenPlayer) {
	m := setupWattenWebMock()
	players := makeWattenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWattenWebPresenter_Output(t *testing.T) {
	p := new(presenter.WattenWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupWattenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		result := p.Output(m, nil)
		var resObj controller.WattenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.WattenPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 2, resObj.Stake)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupWattenWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.WattenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
	})

	t.Run("declare phase", func(t *testing.T) {
		m, _ := setupWattenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WattenPhaseDeclare)
		result := p.Output(m, nil)
		assert.Contains(t, result, "watten.declarePhase")
	})

	t.Run("respond phase", func(t *testing.T) {
		m, _ := setupWattenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WattenPhaseRespond)
		result := p.Output(m, nil)
		assert.Contains(t, result, "watten.respondPhase")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupWattenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "watten.result.team0Win")
	})
}

func TestWattenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.WattenWebPresenter)
	m, _ := setupWattenWebMockWithPlayers()
	rank, suit := 10, 2
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.WattenHint{Action: "declare", Rank: &rank, Suit: &suit, Reason: "declare_strong"})
	result := p.HintOutput(m)
	var resObj controller.WattenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.NotNil(t, resObj.Hint)
	assert.Equal(t, "declare", resObj.Hint.Action)
}

func TestWattenWebPresenter_HintOutput_Nil(t *testing.T) {
	p := new(presenter.WattenWebPresenter)
	m, _ := setupWattenWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return((*domain.WattenHint)(nil))
	result := p.HintOutput(m)
	var resObj controller.WattenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.Nil(t, resObj.Hint)
}

func TestWattenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WattenWebPresenter)
	m := setupWattenWebMock()
	assert.NotNil(t, p.ActionLogOutput(m))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestWattenWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	wtg, _ := setupWattenWebMockWithPlayers()
	wtg.ExpectedCalls = removeMockCall(wtg.ExpectedCalls, "GetHint")
	wtg.On("GetHint").Return(&domain.WattenHint{Action: "play", CardIndex: &idx, Reason: "lead_low"})

	result := new(presenter.WattenWebPresenter).Output(wtg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "watten.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestWattenWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	wtg, _ := setupWattenWebMockWithPlayers()
	wtg.ExpectedCalls = removeMockCall(wtg.ExpectedCalls, "GetHint")
	wtg.On("GetHint").Return(&domain.WattenHint{Action: "play", CardIndex: &idx, Reason: "lead_low"})
	assert.Contains(t, new(presenter.WattenWebPresenter).HintOutput(wtg), "watten.hintRequested")

	none, _ := setupWattenWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.WattenHint)(nil))
	assert.Contains(t, new(presenter.WattenWebPresenter).HintOutput(none), "watten.noHint")
}
