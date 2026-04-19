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

func makeBadugiForPresenter() (*domain.Badugi, []*domain.BadugiPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.BadugiPlayer{
		domain.NewBadugiPlayer(true, domain.BadugiStyleBalanced),
		domain.NewBadugiPlayer(false, domain.BadugiStyleConservative),
		domain.NewBadugiPlayer(false, domain.BadugiStyleAggressive),
		domain.NewBadugiPlayer(false, domain.BadugiStyleBluffer),
	}
	return domain.NewBadugi(tc, players, domain.DefaultBadugiConfig()), players
}

func TestBadugiWebPresenter_Output_Initial(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDeal)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

	result := pres.Output(bd, nil)
	var out controller.BadugiWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, domain.BadugiPhaseDeal, out.Phase)
	assert.Equal(t, 4, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, "", out.Message)
	assert.Len(t, out.SidePots, 0)
	assert.Len(t, out.RoundResults, 0)
}

func TestBadugiWebPresenter_Output_HumanCardsVisibleInDraw(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseDraw)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	result := pres.Output(bd, nil)
	var out controller.BadugiWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[0].Cards, 2)
	assert.Equal(t, "SPADE", out.Players[0].Cards[0].Design)
	assert.Equal(t, 1, out.Players[0].Cards[0].Value)
}

func TestBadugiWebPresenter_Output_CpuCardsHiddenUntilEnd(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, players := makeBadugiForPresenter()
	bd.SetPhase(domain.BadugiPhaseBet)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	result := pres.Output(bd, nil)
	var out controller.BadugiWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[1].Cards, 0, "CPU hand is hidden mid-hand")
}

func TestBadugiWebPresenter_Output_ShowdownIncludesBestCards(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, players := makeBadugiForPresenter()
	// Perfect Badugi for player 1.
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
	_ = players[1].EvalHand()
	bd.SetPhase(domain.BadugiPhaseEnd)

	result := pres.Output(bd, nil)
	var out controller.BadugiWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out.Players[1].Cards, 4, "CPU hand is revealed at showdown")
	assert.Equal(t, 4, out.Players[1].HandSize)
	assert.Equal(t, "Badugi", out.Players[1].HandName)
	assert.Len(t, out.Players[1].BestCards, 4)
}

func TestBadugiWebPresenter_Output_ErrorMessageSurfaces(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, _ := makeBadugiForPresenter()
	result := pres.Output(bd, errors.New("boom"))
	var out controller.BadugiWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "boom", out.Message)
}

func TestBadugiWebPresenter_Output_EndMessages(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)

	t.Run("human wins", func(t *testing.T) {
		bd, _ := makeBadugiForPresenter()
		bd.SetPhase(domain.BadugiPhaseEnd)
		bd.SetGameEndFlag(true)
		bd.SetRoundResults([]domain.BadugiResult{{PlayerIdx: 0, WonAmount: 100}})

		var out controller.BadugiWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(bd, nil)), &out))
		assert.Equal(t, "badugi.result.win", out.MessageCode)
	})

	t.Run("human lost", func(t *testing.T) {
		bd, _ := makeBadugiForPresenter()
		bd.SetPhase(domain.BadugiPhaseEnd)
		bd.SetGameEndFlag(true)
		bd.SetRoundResults([]domain.BadugiResult{{PlayerIdx: 1, WonAmount: 100}})

		var out controller.BadugiWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(bd, nil)), &out))
		assert.Equal(t, "badugi.result.lose", out.MessageCode)
	})

	t.Run("human folded", func(t *testing.T) {
		bd, players := makeBadugiForPresenter()
		bd.SetPhase(domain.BadugiPhaseEnd)
		bd.SetGameEndFlag(true)
		players[0].SetFolded(true)
		bd.SetRoundResults([]domain.BadugiResult{{PlayerIdx: 1, WonAmount: 100}})

		var out controller.BadugiWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(bd, nil)), &out))
		assert.Equal(t, "badugi.result.folded", out.MessageCode)
	})

	t.Run("no results", func(t *testing.T) {
		bd, _ := makeBadugiForPresenter()
		bd.SetPhase(domain.BadugiPhaseEnd)
		bd.SetGameEndFlag(true)

		var out controller.BadugiWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(bd, nil)), &out))
		assert.Equal(t, "badugi.result.gameOver", out.MessageCode)
	})
}

func TestBadugiWebPresenter_ActionLogOutput(t *testing.T) {
	pres := new(presenter.BadugiWebPresenter)
	bd, _ := makeBadugiForPresenter()
	got := pres.ActionLogOutput(bd)
	assert.NotEmpty(t, got)
}
