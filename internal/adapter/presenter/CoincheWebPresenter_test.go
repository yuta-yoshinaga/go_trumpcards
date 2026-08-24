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

func setupCoincheWebMock() *interfaces.MockCoincheGame {
	m := new(interfaces.MockCoincheGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CoinchePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetContractPoints").Return(0)
	m.On("GetMultiplier").Return(1)
	m.On("GetDouble").Return(domain.CoincheDoubleNone)
	m.On("GetBiddablePoints").Return(domain.CoincheContractPoints)
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundBeloteBonus", 0).Return(0)
	m.On("GetRoundBeloteBonus", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultCoincheConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupCoincheWebMockWithPlayers() (*interfaces.MockCoincheGame, []*domain.CoinchePlayer) {
	m := setupCoincheWebMock()
	players := makeCoinchePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCoincheWebPresenter_Output(t *testing.T) {
	p := new(presenter.CoincheWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCoincheWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.CoincheWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.CoinchePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
	})

	t.Run("CPU cards hidden", func(t *testing.T) {
		m, players := setupCoincheWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.CoincheWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 1, len(resObj.Players[0].Cards), "human cards visible")
		assert.Equal(t, 0, len(resObj.Players[1].Cards), "CPU cards hidden")
	})

	t.Run("last error message", func(t *testing.T) {
		m, _ := setupCoincheWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid play"))
		var resObj controller.CoincheWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid play", resObj.Message)
	})

	t.Run("game-end message", func(t *testing.T) {
		m, _ := setupCoincheWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.CoincheWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.Equal(t, 1, resObj.WinnerTeam)
	})

	t.Run("each phase produces message code", func(t *testing.T) {
		cases := []struct {
			phase domain.CoinchePhase
			code  string
		}{
			{domain.CoinchePhaseBid, "coinche.bidPhase"},
			{domain.CoinchePhaseDouble, "coinche.doublePhase"},
			{domain.CoinchePhasePlay, "coinche.playPhase.lead"},
			{domain.CoinchePhaseTrickEnd, "coinche.trickEnd"},
			{domain.CoinchePhaseRoundEnd, "coinche.roundEnd"},
		}
		for _, tc := range cases {
			m, _ := setupCoincheWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(tc.phase)
			result := p.Output(m, nil)
			var resObj controller.CoincheWebOutput
			_ = json.Unmarshal([]byte(result), &resObj)
			assert.Equal(t, tc.code, resObj.MessageCode, "phase %d", tc.phase)
		}
	})
}

func TestCoincheWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CoincheWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupCoincheWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.CoincheHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.CoincheWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupCoincheWebMockWithPlayers()
		points, suit := 110, domain.CardDesignSpade
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CoincheHint{Bid: &points, Suit: &suit, Reason: "strategic_bid"})

		result := p.HintOutput(m)
		var resObj controller.CoincheWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.NotNil(t, resObj.Hint.Bid)
		assert.NotNil(t, resObj.Hint.Suit)
		assert.Equal(t, "strategic_bid", resObj.Hint.Reason)
	})
}

func TestCoincheWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupCoincheWebMockWithPlayers()
	p := new(presenter.CoincheWebPresenter)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。Coinche.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestCoincheWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	blg, _ := setupCoincheWebMockWithPlayers()
	blg.ExpectedCalls = removeMockCall(blg.ExpectedCalls, "GetHint")
	blg.On("GetHint").Return(&domain.CoincheHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.CoincheWebPresenter).Output(blg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
