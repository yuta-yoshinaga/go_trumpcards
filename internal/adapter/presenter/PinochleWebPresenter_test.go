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

func setupPinochleWebMock() *interfaces.MockPinochleGame {
	m := new(interfaces.MockPinochleGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PinochlePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetHighestBid").Return(20)
	m.On("GetHighestBidder").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultPinochleConfig())
	m.On("GetPlayerMelds").Return([domain.PinochlePlayerCnt][]*domain.PinochleMeld{})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupPinochleWebMockWithPlayers() (*interfaces.MockPinochleGame, []*domain.PinochlePlayer) {
	m := setupPinochleWebMock()
	players := []*domain.PinochlePlayer{
		domain.NewPinochlePlayer(true, 0),
		domain.NewPinochlePlayer(false, 1),
		domain.NewPinochlePlayer(false, 0),
		domain.NewPinochlePlayer(false, 1),
	}
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPinochleWebPresenter_Output(t *testing.T) {
	p := new(presenter.PinochleWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupPinochleWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.PinochleWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.PinochlePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupPinochleWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		result := p.Output(m, nil)
		var resObj controller.PinochleWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 1, len(resObj.Players[0].Cards))
		assert.Equal(t, 0, len(resObj.Players[1].Cards))
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupPinochleWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.PinochleWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end message", func(t *testing.T) {
		m, _ := setupPinochleWebMockWithPlayers()
		m.ExpectedCalls = nil
		m.On("GetRoundNumber").Return(1)
		m.On("GetTrickNumber").Return(12)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPhase").Return(domain.PinochlePhaseGameEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetBidPlayerIdx").Return(0)
		m.On("GetDealerIdx").Return(0)
		m.On("GetTrumpSuit").Return(1)
		m.On("GetHighestBid").Return(30)
		m.On("GetHighestBidder").Return(0)
		m.On("GetTeamScore", 0).Return(1500)
		m.On("GetTeamScore", 1).Return(800)
		m.On("GetWinnerTeam").Return(0)
		m.On("GetLeadPlayerIdx").Return(0)
		m.On("GetConfig").Return(domain.DefaultPinochleConfig())
		m.On("GetPlayerMelds").Return([domain.PinochlePlayerCnt][]*domain.PinochleMeld{})
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("IsHumanTurn").Return(false)
		m.On("GetPlayerCnt").Return(4)
		m.On("GetPlayer", 0).Return(domain.NewPinochlePlayer(true, 0))
		m.On("GetPlayer", 1).Return(domain.NewPinochlePlayer(false, 1))
		m.On("GetPlayer", 2).Return(domain.NewPinochlePlayer(false, 0))
		m.On("GetPlayer", 3).Return(domain.NewPinochlePlayer(false, 1))
		m.On("GetHint").Return(nil).Maybe()

		result := p.Output(m, nil)
		var resObj controller.PinochleWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.WinnerTeam)
		assert.Contains(t, resObj.MessageCode, "pinochle.result.team0Win")
	})

	t.Run("valid play indices included in play phase", func(t *testing.T) {
		m, _ := setupPinochleWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.PinochleWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
	})
}

func TestPinochleWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PinochleWebPresenter)

	t.Run("returns hint", func(t *testing.T) {
		m, _ := setupPinochleWebMockWithPlayers()
		bid := 25
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.PinochleHint{BidAmount: &bid, Reason: "hint_bid"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "hint_bid")
	})

	t.Run("returns nil hint", func(t *testing.T) {
		m, _ := setupPinochleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.PinochleHint)(nil))

		result := p.HintOutput(m)
		assert.Contains(t, result, "pinochle.noHint")
	})
}

func TestPinochleWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PinochleWebPresenter)
	m := setupPinochleWebMock()

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestPinochleWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	png, _ := setupPinochleWebMockWithPlayers()
	png.ExpectedCalls = removeMockCall(png.ExpectedCalls, "GetHint")
	png.On("GetHint").Return(&domain.PinochleHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.PinochleWebPresenter).Output(png, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "pinochle.hintRequested")
}

// **HintOutput は状態レスポンスを返し、「頼んだヒント」の印を付ける。**
// 以前はヒント構造体を裸で返していて `hint` キーが無かった (#4483)。
func TestPinochleWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	png, _ := setupPinochleWebMockWithPlayers()
	png.ExpectedCalls = removeMockCall(png.ExpectedCalls, "GetHint")
	png.On("GetHint").Return(&domain.PinochleHint{CardIndex: &idx, Reason: "follow_suit"})
	got := new(presenter.PinochleWebPresenter).HintOutput(png)
	assert.Contains(t, got, "pinochle.hintRequested")
	assert.Contains(t, got, `"hint"`)
	assert.Contains(t, got, `"players"`, "HintOutput must return a state response, not a bare hint")

	none, _ := setupPinochleWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.PinochleHint)(nil))
	assert.Contains(t, new(presenter.PinochleWebPresenter).HintOutput(none), "pinochle.noHint")
}
