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

func setupBauernschnapsenWebMock() *interfaces.MockBauernschnapsenGame {
	m := new(interfaces.MockBauernschnapsenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BauernschnapsenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetContract").Return(domain.BauernschnapsenContractRufer)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetValidPlayIndices", 0).Return([]int(nil))
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundMarriagePoints", 0).Return(0)
	m.On("GetRoundMarriagePoints", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	m.On("GetConfig").Return(domain.DefaultBauernschnapsenConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupBauernschnapsenWebMockWithPlayers() (*interfaces.MockBauernschnapsenGame, []*domain.BauernschnapsenPlayer) {
	m := setupBauernschnapsenWebMock()
	players := makeBauernschnapsenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBauernschnapsenWebPresenter_Output(t *testing.T) {
	p := new(presenter.BauernschnapsenWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBauernschnapsenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		var resObj controller.BauernschnapsenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.BauernschnapsenPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, int(domain.BauernschnapsenContractRufer), resObj.Contract)
		assert.Equal(t, 1, resObj.DeclarerIdx)
	})

	// **追従はトリック 1 から必須**なので、出せる札は画面に届かないといけない。
	// 届かないと、画面はどの札も押せる前提で描いてしまう。
	t.Run("valid play indices reach the page", func(t *testing.T) {
		m, players := setupBauernschnapsenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.On("GetValidPlayIndices", 0).Return([]int{1, 3})

		result := p.Output(m, nil)
		var resObj controller.BauernschnapsenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, []int{1, 3}, resObj.ValidPlayIndices)
	})

	// nil は JSON の null になり、画面側の .length で落ちる。空配列で出す。
	t.Run("valid play indices are an empty array, never null", func(t *testing.T) {
		m, players := setupBauernschnapsenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, `"validPlayIndices":[]`)
		var resObj controller.BauernschnapsenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.ValidPlayIndices)
		assert.Empty(t, resObj.ValidPlayIndices)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupBauernschnapsenWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.BauernschnapsenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
	})

	t.Run("round end phase", func(t *testing.T) {
		m, _ := setupBauernschnapsenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BauernschnapsenPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "bauernschnapsen.roundEnd")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupBauernschnapsenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "bauernschnapsen.result.team0Win")
	})
}

func TestBauernschnapsenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BauernschnapsenWebPresenter)
	m, players := setupBauernschnapsenWebMockWithPlayers()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	idx := 0
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return(&domain.BauernschnapsenHint{CardIndex: &idx, Reason: "follow_win"})
	result := p.HintOutput(m)
	var resObj controller.BauernschnapsenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.NotNil(t, resObj.Hint)
	assert.Equal(t, "follow_win", resObj.Hint.Reason)
}

func TestBauernschnapsenWebPresenter_HintOutput_Nil(t *testing.T) {
	p := new(presenter.BauernschnapsenWebPresenter)
	m, _ := setupBauernschnapsenWebMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
	m.On("GetHint").Return((*domain.BauernschnapsenHint)(nil))
	result := p.HintOutput(m)
	var resObj controller.BauernschnapsenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.Nil(t, resObj.Hint)
}

func TestBauernschnapsenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BauernschnapsenWebPresenter)
	m := setupBauernschnapsenWebMock()
	assert.NotNil(t, p.ActionLogOutput(m))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。Bauernschnapsen.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestBauernschnapsenWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ggg, _ := setupBauernschnapsenWebMockWithPlayers()
	ggg.ExpectedCalls = removeMockCall(ggg.ExpectedCalls, "GetHint")
	ggg.On("GetHint").Return(&domain.BauernschnapsenHint{CardIndex: &idx, Reason: "lead_low"})

	result := new(presenter.BauernschnapsenWebPresenter).Output(ggg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
