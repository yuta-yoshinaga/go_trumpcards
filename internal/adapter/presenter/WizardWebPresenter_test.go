//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeWizardPlayers() []*domain.WizardPlayer {
	return []*domain.WizardPlayer{
		domain.NewWizardPlayer(true),
		domain.NewWizardPlayer(false),
		domain.NewWizardPlayer(false),
		domain.NewWizardPlayer(false),
	}
}

func setupWizardWebMock() *interfaces.MockWizardGame {
	m := new(interfaces.MockWizardGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTotalRounds").Return(15)
	m.On("GetHandSize").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WizardPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetRestrictedBid").Return(-1)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultWizardConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupWizardWebMockWithPlayers() (*interfaces.MockWizardGame, []*domain.WizardPlayer) {
	m := setupWizardWebMock()
	players := makeWizardPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWizardWebPresenter_Output(t *testing.T) {
	p := new(presenter.WizardWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupWizardWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.WizardWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 1, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 15, resObj.TotalRounds)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.NotNil(t, resObj.TrumpCard)
		assert.Equal(t, -1, resObj.RestrictedBid)
		assert.Equal(t, -1, resObj.WinnerIdx)
	})

	t.Run("wizard and jester cards carry procedural face descriptors", func(t *testing.T) {
		m, players := setupWizardWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))    // standard
		players[0].AddCard(domain.NewCard(domain.WizardDesignWizard, 1, false)) // wizard
		players[0].AddCard(domain.NewCard(domain.WizardDesignJester, 1, false)) // jester

		result := p.Output(m, nil)

		var resObj controller.WizardWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		cards := resObj.Players[0].Cards
		assert.Equal(t, 3, len(cards))

		// Standard card: no procedural fields.
		std := cards[0]
		assert.Equal(t, "", std.Deck)
		assert.Equal(t, "", std.Label)
		assert.Equal(t, "", std.Glyph)

		// Wizard card: procedural face descriptor.
		wiz := cards[1]
		assert.Equal(t, "wizard", wiz.Deck)
		assert.Equal(t, "Wizard", wiz.Label)
		assert.Equal(t, "✦", wiz.Glyph)
		assert.Equal(t, "purple", wiz.Color)

		// Jester card: procedural face descriptor.
		jes := cards[2]
		assert.Equal(t, "wizard", jes.Deck)
		assert.Equal(t, "Jester", jes.Label)
		assert.Equal(t, "☺", jes.Glyph)
		assert.Equal(t, "green", jes.Color)

		// The raw JSON must omit procedural keys for the standard card.
		firstCardJSON := extractFirstPlayerFirstCardJSON(t, result)
		assert.NotContains(t, firstCardJSON, `"deck"`)
		assert.NotContains(t, firstCardJSON, `"glyph"`)
		assert.NotContains(t, firstCardJSON, `"label"`)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("bid phase message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseBid)

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "wizard.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "wizard.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
		})

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "wizard.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "wizard.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "wizard.roundEnd", resObj.MessageCode)
	})

	t.Run("game end message", func(t *testing.T) {
		m, players := setupWizardWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[0].SetBid(3)

		result := p.Output(m, nil)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.NotEmpty(t, resObj.Message)
	})
}

// extractFirstPlayerFirstCardJSON returns the raw JSON object substring for the
// first card of the first player, so tests can assert on omitempty behaviour.
func extractFirstPlayerFirstCardJSON(t *testing.T, result string) string {
	t.Helper()
	idx := strings.Index(result, `"cards":[`)
	assert.GreaterOrEqual(t, idx, 0)
	start := strings.Index(result[idx:], "{")
	assert.GreaterOrEqual(t, start, 0)
	rest := result[idx+start:]
	end := strings.Index(rest, "}")
	assert.GreaterOrEqual(t, end, 0)
	return rest[:end+1]
}

func TestWizardWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.WizardWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		bid := 3
		m.On("GetHint").Return(&domain.WizardHint{Bid: &bid, Reason: "strategic_bid"})

		result := p.HintOutput(m)
		var resObj controller.WizardWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, 3, *resObj.Hint.Bid)
	})

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupWizardWebMockWithPlayers()
		m.On("GetHint").Return((*domain.WizardHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.WizardWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})
}

func TestWizardWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WizardWebPresenter)
	m := setupWizardWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "test"},
	})

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
