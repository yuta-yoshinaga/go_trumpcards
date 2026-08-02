//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKnockoutWhistWebMock() *interfaces.MockKnockoutWhistGame {
	m := new(interfaces.MockKnockoutWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetHandSize").Return(7)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetActiveCount").Return(4)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultKnockoutWhistConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupKnockoutWhistWebMockWithPlayers() (*interfaces.MockKnockoutWhistGame, []*domain.KnockoutWhistPlayer) {
	m := setupKnockoutWhistWebMock()
	players := makeKnockoutWhistPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKnockoutWhistWebPresenter_Output(t *testing.T) {
	p := new(presenter.KnockoutWhistWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupKnockoutWhistWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.KnockoutWhistPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, 7, resObj.HandSize)
		assert.Equal(t, 4, resObj.ActiveCount)
		assert.Equal(t, "knockoutwhist.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		// default player state
		assert.Equal(t, 1, resObj.Players[0].Dogbones)
		assert.False(t, resObj.Players[0].Eliminated)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.KnockoutWhistCpuDifficultyNormal), resObj.Config.CpuDifficulty)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "knockoutwhist.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KnockoutWhistPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "knockoutwhist.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KnockoutWhistPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "knockoutwhist.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human; winner = 0
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "knockoutwhist.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human; winner = 1 (CPU)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "knockoutwhist.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("eliminated and dogbones propagated", func(t *testing.T) {
		m, players := setupKnockoutWhistWebMockWithPlayers()
		players[1].SetEliminated(true)
		players[2].SetDogbones(0)
		players[2].IncRoundTricks()
		result := p.Output(m, nil)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.Players[1].Eliminated)
		assert.Equal(t, 0, resObj.Players[2].Dogbones)
		assert.Equal(t, 1, resObj.Players[2].RoundTricks)
	})
}

func TestKnockoutWhistWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KnockoutWhistWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.KnockoutWhistHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupKnockoutWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.KnockoutWhistHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.KnockoutWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestKnockoutWhistWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KnockoutWhistWebPresenter)
	m := new(interfaces.MockKnockoutWhistGame)
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
// Output 側にゲートは置きません。KnockoutWhist.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestKnockoutWhistWebPresenterOutputCarriesTheHint(t *testing.T) {
	kwg, _ := setupKnockoutWhistWebMockWithPlayers()
	kwg.ExpectedCalls = removeMockCall(kwg.ExpectedCalls, "GetHint")
	kwg.On("GetHint").Return(&domain.KnockoutWhistHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.KnockoutWhistWebPresenter).Output(kwg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
// ページは `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
// 付いていないとヒントを押しても画面に何も出ない。
func TestKnockoutWhistWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultKnockoutWhist()
	g.Reset()
	// **Reset 直後は人間の手番とは限らない。**GetHint は手番でなければ nil を
	// 返すので、席を人間に固定しないとこのテストは前提で落ちる。
	g.SetCurrentPlayerIdx(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")
	assert.Contains(t, new(presenter.KnockoutWhistWebPresenter).HintOutput(g), "knockoutwhist.hintRequested")
}
