//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupWhistCuiMock() *interfaces.MockWhistGame {
	m := new(interfaces.MockWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	m.On("IsHumanTurn").Return(false).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int(nil)).Maybe()

	m.On("GetPlayerCnt").Return(4)
	players := []*domain.WhistPlayer{
		domain.NewWhistPlayer(true, 0),
		domain.NewWhistPlayer(false, 1),
		domain.NewWhistPlayer(false, 0),
		domain.NewWhistPlayer(false, 1),
	}
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}

	return m
}

func TestWhistCuiPresenter_PointLimit(t *testing.T) {
	limit := 11

	m := setupWhistCuiMock()

	cfg := domain.DefaultWhistConfig()
	cfg.PointLimit = limit
	m.On("GetConfig").Return(cfg)

	p := new(presenter.WhistCuiPresenter)

	oldLoc := i18n.Lang()
	i18n.SetLang("ja")
	defer i18n.SetLang(oldLoc)

	out := p.Output(m, nil)

	assert.Contains(t, out, "チームスコア: チーム0=0  チーム1=0")
	assert.Contains(t, out, "目標点: 11")
}
