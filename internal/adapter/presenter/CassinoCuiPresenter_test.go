package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCassinoCuiPresenter_Output(t *testing.T) {
	p := new(presenter.CassinoCuiPresenter)

	t.Run("initial state includes header", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		out := p.Output(cg, nil)
		assert.Contains(t, out, "Cassino")
	})

	t.Run("error is displayed", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		out := p.Output(cg, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("table with cards is rendered", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		out := p.Output(cg, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("builds shown", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetBuilds([]*domain.CassinoBuild{
			domain.NewCassinoBuild(0, 9, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false), domain.NewCard(domain.CardDesignSpade, 5, false)}),
		})
		out := p.Output(cg, nil)
		assert.Contains(t, out, "ビルド")
	})

	t.Run("human action displayed", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetHumanAction(&domain.CassinoAction{
			PlayerIdx:  0,
			Type:       domain.CassinoActionTake,
			PlayedCard: domain.NewCard(domain.CardDesignHeart, 5, false),
			IsSweep:    true,
		})
		out := p.Output(cg, nil)
		assert.Contains(t, out, "あなたの行動")
	})

	t.Run("cpu actions displayed", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetCpuActions([]*domain.CassinoAction{
			{PlayerIdx: 1, Type: domain.CassinoActionBuild, BuildValue: 8, PlayedCard: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 2, Type: domain.CassinoActionTrail, PlayedCard: domain.NewCard(domain.CardDesignHeart, 7, false)},
		})
		out := p.Output(cg, nil)
		assert.Contains(t, out, "CPUの行動")
	})

	t.Run("game end result", func(t *testing.T) {
		players := makeCassinoPlayersForPresenter()
		cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
		cg.SetGameEndFlag(true)
		out := p.Output(cg, nil)
		assert.Contains(t, out, "ゲーム終了")
	})
}

func TestCassinoCuiPresenter_ActionLog(t *testing.T) {
	p := new(presenter.CassinoCuiPresenter)
	players := makeCassinoPlayersForPresenter()
	cg := domain.NewCassino(domain.NewTrumpCards(0), players, domain.DefaultCassinoConfig())
	assert.NotEmpty(t, p.ActionLogOutput(cg))
}
