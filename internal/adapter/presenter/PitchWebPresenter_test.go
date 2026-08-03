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

func setupPitchWebMock() *interfaces.MockPitchGame {
	m := new(interfaces.MockPitchGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetCurrentBid").Return(0)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetBidWinnerIdx").Return(-1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PitchPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultPitchConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupPitchWebMockWithPlayers() (*interfaces.MockPitchGame, []*domain.PitchPlayer) {
	m := setupPitchWebMock()
	players := makePitchPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPitchWebPresenter_Output(t *testing.T) {
	p := new(presenter.PitchWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupPitchWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		var resObj controller.PitchWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.Equal(t, 1, resObj.Phase) // PitchPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 3, resObj.DealerIdx)
		assert.Equal(t, -1, resObj.BidWinnerIdx)
		assert.Equal(t, 0, resObj.TrumpSuit)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupPitchWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.False(t, resObj.Players[1].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("error message included", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid bid"))
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid bid", resObj.Message)
	})

	t.Run("game end shows winner", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.NotEmpty(t, resObj.MessageCode)
	})
}

func TestPitchWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PitchWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.PitchHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		bid := 3
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.PitchHint{Bid: &bid, Reason: "bid_strong"})
		result := p.HintOutput(m)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "bid_strong", resObj.Hint.Reason)
		assert.Equal(t, 3, *resObj.Hint.Bid)
	})
}

// setupPitchWebMockCustom builds a fully-wired mock with a caller-controlled
// phase, trick number and action log so the last-trick reconstruction can be
// exercised deterministically.
func setupPitchWebMockCustom(phase domain.PitchPhase, trickNumber int, log []*domain.ActionLogEntry) *interfaces.MockPitchGame {
	m := new(interfaces.MockPitchGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(trickNumber)
	m.On("GetDealerIdx").Return(3)
	m.On("GetCurrentBid").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetBidWinnerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultPitchConfig())
	m.On("GetActionLog").Return(log)
	m.On("GetValidPlayIndices", 0).Return([]int{0})
	players := makePitchPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func TestPitchWebPresenter_LastTrick(t *testing.T) {
	p := new(presenter.PitchWebPresenter)

	t.Run("reconstructs the just-completed trick with its winner", func(t *testing.T) {
		log := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}},
			{TurnNumber: 1, PlayerIdx: 1, ActionType: "play", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)}},
			{TurnNumber: 1, PlayerIdx: 2, ActionType: "play", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}},
			{TurnNumber: 1, PlayerIdx: 3, ActionType: "play", Cards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)}},
			{TurnNumber: 1, PlayerIdx: 2, ActionType: "trick_win", Cards: nil},
		}
		// After pressing "next": play phase, trick 2, current trick cleared.
		m := setupPitchWebMockCustom(domain.PitchPhasePlay, 2, log)
		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.LastTrick, 4)
		assert.Equal(t, 0, resObj.LastTrick[0].PlayerIdx)
		assert.Equal(t, 3, resObj.LastTrick[3].PlayerIdx)
		assert.Equal(t, 2, resObj.LastTrickWinner)
	})

	t.Run("hidden on the round's first trick", func(t *testing.T) {
		m := setupPitchWebMockCustom(domain.PitchPhasePlay, 1, nil)
		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.LastTrick)
		assert.Equal(t, -1, resObj.LastTrickWinner)
	})

	t.Run("empty when no trick_win logged yet", func(t *testing.T) {
		// TrickEnd phase but the log has no trick_win (e.g. partial play): empty.
		log := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}},
		}
		m := setupPitchWebMockCustom(domain.PitchPhaseTrickEnd, 2, log)
		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.LastTrick)
		assert.Equal(t, -1, resObj.LastTrickWinner)
	})
}

func TestPitchWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PitchWebPresenter)
	m := new(interfaces.MockPitchGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bid 3"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You bid 3")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。Pitch.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestPitchWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ptg, _ := setupPitchWebMockWithPlayers()
	ptg.ExpectedCalls = removeMockCall(ptg.ExpectedCalls, "GetHint")
	ptg.On("GetHint").Return(&domain.PitchHint{CardIndex: &idx, Reason: "lead_trump"})

	result := new(presenter.PitchWebPresenter).Output(ptg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
