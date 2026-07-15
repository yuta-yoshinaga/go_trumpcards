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
)

func makeDeuceToSevenForPresenter() (*domain.DeuceToSeven, []*domain.DeuceToSevenPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.DeuceToSevenPlayer{
		domain.NewDeuceToSevenPlayer(true, domain.DeuceToSevenStyleBalanced),
		domain.NewDeuceToSevenPlayer(false, domain.DeuceToSevenStyleConservative),
		domain.NewDeuceToSevenPlayer(false, domain.DeuceToSevenStyleAggressive),
		domain.NewDeuceToSevenPlayer(false, domain.DeuceToSevenStyleBluffer),
	}
	return domain.NewDeuceToSeven(tc, players, domain.DefaultDeuceToSevenConfig()), players
}

func TestDeuceToSevenWebPresenter_Output_Initial(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	dt.SetPhase(domain.DeuceToSevenPhaseDeal)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

	result := pres.Output(dt, nil)
	var out controller.DeuceToSevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, domain.DeuceToSevenPhaseDeal, out.Phase)
	assert.Equal(t, 4, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, "", out.Message)
	assert.Len(t, out.SidePots, 0)
	assert.Len(t, out.RoundResults, 0)
}

func TestDeuceToSevenWebPresenter_Output_HumanCardsVisibleInDraw(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	dt.SetPhase(domain.DeuceToSevenPhaseDraw)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	result := pres.Output(dt, nil)
	var out controller.DeuceToSevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[0].Cards, 2)
	assert.Equal(t, "SPADE", out.Players[0].Cards[0].Design)
	assert.Equal(t, 1, out.Players[0].Cards[0].Value)
}

func TestDeuceToSevenWebPresenter_Output_CpuCardsHiddenUntilEnd(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	dt.SetPhase(domain.DeuceToSevenPhaseBet)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	result := pres.Output(dt, nil)
	var out controller.DeuceToSevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[1].Cards, 0, "CPU hand is hidden mid-hand")
}

func TestDeuceToSevenWebPresenter_Output_ShowdownRevealsHand(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, players := makeDeuceToSevenForPresenter()
	// Nut low 7-5-4-3-2 for player 1.
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	_ = players[1].EvalHand()
	dt.SetPhase(domain.DeuceToSevenPhaseEnd)

	result := pres.Output(dt, nil)
	var out controller.DeuceToSevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[1].Cards, 5, "CPU hand is revealed at showdown")
	assert.Equal(t, domain.PokerHandHighCard, out.Players[1].HandRank)
	assert.Equal(t, "High Card", out.Players[1].HandName)
}

func TestDeuceToSevenWebPresenter_Output_ErrorMessageSurfaces(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	result := pres.Output(dt, errors.New("boom"))
	var out controller.DeuceToSevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "boom", out.Message)
}

func TestDeuceToSevenWebPresenter_Output_EndMessages(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)

	t.Run("human wins", func(t *testing.T) {
		dt, _ := makeDeuceToSevenForPresenter()
		dt.SetPhase(domain.DeuceToSevenPhaseEnd)
		dt.SetGameEndFlag(true)
		dt.SetRoundResults([]domain.DeuceToSevenResult{{PlayerIdx: 0, WonAmount: 100}})

		var out controller.DeuceToSevenWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(dt, nil)), &out))
		assert.Equal(t, "deucetoseven.result.win", out.MessageCode)
	})

	t.Run("human lost", func(t *testing.T) {
		dt, _ := makeDeuceToSevenForPresenter()
		dt.SetPhase(domain.DeuceToSevenPhaseEnd)
		dt.SetGameEndFlag(true)
		dt.SetRoundResults([]domain.DeuceToSevenResult{{PlayerIdx: 1, WonAmount: 100}})

		var out controller.DeuceToSevenWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(dt, nil)), &out))
		assert.Equal(t, "deucetoseven.result.lose", out.MessageCode)
	})

	t.Run("human folded", func(t *testing.T) {
		dt, players := makeDeuceToSevenForPresenter()
		dt.SetPhase(domain.DeuceToSevenPhaseEnd)
		dt.SetGameEndFlag(true)
		players[0].SetFolded(true)
		dt.SetRoundResults([]domain.DeuceToSevenResult{{PlayerIdx: 1, WonAmount: 100}})

		var out controller.DeuceToSevenWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(dt, nil)), &out))
		assert.Equal(t, "deucetoseven.result.folded", out.MessageCode)
	})

	t.Run("no results", func(t *testing.T) {
		dt, _ := makeDeuceToSevenForPresenter()
		dt.SetPhase(domain.DeuceToSevenPhaseEnd)
		dt.SetGameEndFlag(true)

		var out controller.DeuceToSevenWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(dt, nil)), &out))
		assert.Equal(t, "deucetoseven.result.gameOver", out.MessageCode)
	})
}

func TestDeuceToSevenWebPresenter_ActionLogOutput(t *testing.T) {
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	got := pres.ActionLogOutput(dt)
	assert.NotEmpty(t, got)
}

func TestDeuceToSevenWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	pres := new(presenter.DeuceToSevenWebPresenter)
	dt, _ := makeDeuceToSevenForPresenter()
	assert.Equal(t, pres.Output(dt, nil), pres.HintOutput(dt))
}
