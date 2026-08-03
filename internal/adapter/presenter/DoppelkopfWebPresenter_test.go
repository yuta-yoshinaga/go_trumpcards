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

func setupDoppelkopfWebMock() *interfaces.MockDoppelkopfGame {
	m := new(interfaces.MockDoppelkopfGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("IsSoloRe").Return(false)
	m.On("AreTeamsRevealed").Return(false)
	m.On("IsReAnnounced").Return(false)
	m.On("IsKontraAnnounced").Return(false)
	m.On("CanHumanAnnounce").Return(false)
	m.On("GetRoundRePoints").Return(0)
	m.On("GetRoundReWon").Return(false)
	m.On("GetRoundGamePoints").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsRe", 0).Return(false)
	m.On("IsRe", 1).Return(false)
	m.On("IsRe", 2).Return(false)
	m.On("IsRe", 3).Return(false)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultDoppelkopfConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupDoppelkopfWebMockWithPlayers() (*interfaces.MockDoppelkopfGame, []*domain.DoppelkopfPlayer) {
	m := setupDoppelkopfWebMock()
	players := makeDoppelkopfPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestDoppelkopfWebPresenter_Output(t *testing.T) {
	p := new(presenter.DoppelkopfWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupDoppelkopfWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.DoppelkopfPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, "doppelkopf.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		// teams not revealed: reTeam all false
		assert.Equal(t, []bool{false, false, false, false}, resObj.ReTeam)
		assert.False(t, resObj.Players[0].IsRe)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.DoppelkopfCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 2, resObj.Config.BaseChips)
		assert.Equal(t, 20, resObj.Config.StartChips)
		assert.Equal(t, 40, resObj.Config.TargetChips)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "doppelkopf.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.DoppelkopfPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "doppelkopf.trickEnd", resObj.MessageCode)
	})

	t.Run("round end reveals teams", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "AreTeamsRevealed")
		// removeMockCall only removes the first match; remove all 4 IsRe stubs.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.On("GetPhase").Return(domain.DoppelkopfPhaseRoundEnd)
		m.On("AreTeamsRevealed").Return(true)
		m.On("IsRe", 0).Return(true)
		m.On("IsRe", 1).Return(false)
		m.On("IsRe", 2).Return(true)
		m.On("IsRe", 3).Return(false)
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "doppelkopf.roundEnd", resObj.MessageCode)
		assert.Equal(t, []bool{true, false, true, false}, resObj.ReTeam)
		assert.True(t, resObj.Players[0].IsRe)
		assert.False(t, resObj.Players[1].IsRe)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "doppelkopf.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "doppelkopf.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("youAreRe reflects human team", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		// removeMockCall only removes the first match; remove all 4 IsRe stubs.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsRe")
		m.On("IsRe", 0).Return(true)
		m.On("IsRe", 1).Return(false)
		m.On("IsRe", 2).Return(false)
		m.On("IsRe", 3).Return(false)
		result := p.Output(m, nil)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.YouAreRe)
	})
}

func TestDoppelkopfWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.DoppelkopfWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.DoppelkopfHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupDoppelkopfWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.DoppelkopfHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.DoppelkopfWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestDoppelkopfWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DoppelkopfWebPresenter)
	m := new(interfaces.MockDoppelkopfGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠Q"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestDoppelkopfWebPresenterOutputCarriesTheHint(t *testing.T) {
	dkg, _ := setupDoppelkopfWebMockWithPlayers()
	dkg.ExpectedCalls = removeMockCall(dkg.ExpectedCalls, "GetHint")
	dkg.On("GetHint").Return(&domain.DoppelkopfHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.DoppelkopfWebPresenter).Output(dkg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "doppelkopf.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestDoppelkopfWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	dkg, _ := setupDoppelkopfWebMockWithPlayers()
	dkg.ExpectedCalls = removeMockCall(dkg.ExpectedCalls, "GetHint")
	dkg.On("GetHint").Return(&domain.DoppelkopfHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.DoppelkopfWebPresenter).HintOutput(dkg), "doppelkopf.hintRequested")

	none, _ := setupDoppelkopfWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.DoppelkopfHint)(nil))
	assert.Contains(t, new(presenter.DoppelkopfWebPresenter).HintOutput(none), "doppelkopf.noHint")
}
