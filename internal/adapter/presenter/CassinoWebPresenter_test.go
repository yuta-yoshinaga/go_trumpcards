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

func makeCassinoPlayersForPresenter() []*domain.CassinoPlayer {
	return []*domain.CassinoPlayer{
		domain.NewCassinoPlayer(true),
		domain.NewCassinoPlayer(false),
		domain.NewCassinoPlayer(false),
		domain.NewCassinoPlayer(false),
	}
}

func TestCassinoWebPresenter_Output(t *testing.T) {
	p := new(presenter.CassinoWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		raw := p.Output(cg, nil)
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 0, out.CurrentTurn)
		assert.Equal(t, -1, out.LastCaptureIdx)
		assert.Equal(t, 21, out.Config.TargetScore)
	})

	t.Run("error message is propagated", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		raw := p.Output(cg, errors.New("test-err"))
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Equal(t, "test-err", out.Message)
	})

	t.Run("game end produces result message", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetGameEndFlag(true)
		raw := p.Output(cg, nil)
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Contains(t, out.Message, "Game over")
		assert.Equal(t, "cassino.result.scores", out.MessageCode)
		assert.Contains(t, out.MessageParams, "scores")
	})

	t.Run("builds serialise", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetBuilds([]*domain.CassinoBuild{
			domain.NewCassinoBuild(0, 8, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false), domain.NewCard(domain.CardDesignSpade, 5, false)}),
		})
		raw := p.Output(cg, nil)
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		require.Len(t, out.Builds, 1)
		assert.Equal(t, 8, out.Builds[0].Value)
	})

	t.Run("human action is included", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetHumanAction(&domain.CassinoAction{
			PlayerIdx:  0,
			Type:       domain.CassinoActionTrail,
			PlayedCard: domain.NewCard(domain.CardDesignHeart, 5, false),
		})
		raw := p.Output(cg, nil)
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		require.NotNil(t, out.HumanAction)
		assert.Equal(t, "trail", out.HumanAction.Type)
	})

	t.Run("score detail serialises", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 10, false)})
		cg.SetLastCaptureIdx(0)
		cg.FinishRoundForTest()
		raw := p.Output(cg, nil)
		var out controller.CassinoWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.NotNil(t, out.LastRoundDetail)
	})
}

func TestCassinoWebPresenter_ActionLog(t *testing.T) {
	p := new(presenter.CassinoWebPresenter)
	players := makeCassinoPlayersForPresenter()
	cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
	out := p.ActionLogOutput(cg)
	assert.NotEmpty(t, out)
}

func TestCassinoWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	p := new(presenter.CassinoWebPresenter)
	players := makeCassinoPlayersForPresenter()
	cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
	assert.Equal(t, p.Output(cg, nil), p.HintOutput(cg))
}
