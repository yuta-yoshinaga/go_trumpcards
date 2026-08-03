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

func setupSoloWhistWebMock() *interfaces.MockSoloWhistGame {
	m := new(interfaces.MockSoloWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.SoloWhistBidSolo)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetBids").Return([domain.SoloWhistPlayerCnt]domain.SoloWhistBid{domain.SoloWhistBidSolo, domain.SoloWhistBidPass, domain.SoloWhistBidPass, domain.SoloWhistBidPass})
	m.On("GetPlayerScores").Return([domain.SoloWhistPlayerCnt]int{0, 0, 0, 0})
	m.On("GetRoundTricks").Return([domain.SoloWhistPlayerCnt]int{0, 0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultSoloWhistConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupSoloWhistWebMockWithPlayers() (*interfaces.MockSoloWhistGame, []*domain.SoloWhistPlayer) {
	m := setupSoloWhistWebMock()
	players := makeSoloWhistPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestSoloWhistWebPresenter_Output(t *testing.T) {
	p := new(presenter.SoloWhistWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupSoloWhistWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.SoloWhistPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, "solowhist.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, int(domain.SoloWhistBidSolo), resObj.Contract)
		assert.Equal(t, int(domain.SoloWhistBidSolo), resObj.Bids[0])
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SoloWhistCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 21, resObj.Config.TargetPoints)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanBidTurn")
		m.On("GetPhase").Return(domain.SoloWhistPhaseBid)
		m.On("IsHumanBidTurn").Return(true)
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "solowhist.bidPhase", resObj.MessageCode)
		assert.True(t, resObj.IsHumanBidTurn)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "solowhist.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SoloWhistPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "solowhist.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SoloWhistPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "solowhist.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "solowhist.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "solowhist.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.SoloWhistPlayerCnt]int{4, 2, 0, 0})
		result := p.Output(m, nil)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, resObj.Players[0].Score)
		assert.Equal(t, 2, resObj.Players[1].Score)
	})
}

func TestSoloWhistWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SoloWhistWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.SoloWhistHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSoloWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.SoloWhistHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SoloWhistWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSoloWhistWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SoloWhistWebPresenter)
	m := new(interfaces.MockSoloWhistGame)
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
// Output 側にゲートは置きません。SoloWhist.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestSoloWhistWebPresenterOutputCarriesTheHint(t *testing.T) {
	swg, _ := setupSoloWhistWebMockWithPlayers()
	swg.ExpectedCalls = removeMockCall(swg.ExpectedCalls, "GetHint")
	swg.On("GetHint").Return(&domain.SoloWhistHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.SoloWhistWebPresenter).Output(swg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
// ページは `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
// 付いていないとヒントを押しても画面に何も出ない。
func TestSoloWhistWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultSoloWhist()
	g.Reset()
	// **Reset 直後は人間の手番とは限らず、フェーズも配り直しのまま。**GetHint は
	// プレイフェーズかつ人間の手番でなければ nil を返すので、両方そろえないと
	// このテストは前提で落ちる。
	g.SetPhase(domain.SoloWhistPhasePlay)
	g.SetCurrentPlayerIdx(0)
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")
	assert.Contains(t, new(presenter.SoloWhistWebPresenter).HintOutput(g), "solowhist.hintRequested")
}
