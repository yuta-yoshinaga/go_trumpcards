//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupCariocaCuiMock(phase domain.CariocaPhase, gameEnd bool) (*interfaces.MockCariocaGame, []*domain.CariocaPlayer) {
	m := new(interfaces.MockCariocaGame)
	players := []*domain.CariocaPlayer{
		domain.NewCariocaPlayer(true),
		domain.NewCariocaPlayer(false),
		domain.NewCariocaPlayer(false),
		domain.NewCariocaPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultCariocaConfig())
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentContract").Return(domain.CariocaContractForRound(1))
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCariocaCuiPresenter_Output(t *testing.T) {
	p := new(presenter.CariocaCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupCariocaCuiMock(domain.CariocaPhaseDraw, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "カリオカ")
		assert.Contains(t, out, "ラウンド")
		// Each contract slot is expanded with its index and the meld syntax hint.
		assert.Contains(t, out, strings.Split(i18n.T("carioca.slotLine"), "{{")[0])
		assert.Contains(t, out, i18n.T("carioca.meldSyntaxHint"))
	})

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupCariocaCuiMock(domain.CariocaPhasePlay, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "コントラクト達成が必須")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupCariocaCuiMock(domain.CariocaPhaseRoundEnd, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupCariocaCuiMock(domain.CariocaPhaseGameEnd, true)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupCariocaCuiMock(domain.CariocaPhaseDraw, false)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})

	t.Run("contract met player", func(t *testing.T) {
		m, players := setupCariocaCuiMock(domain.CariocaPhasePlay, false)
		players[0].SetContractMet(true)
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "コントラクト達成済み")
		assert.NotContains(t, out, "コントラクト達成が必須")
	})
}

func TestCariocaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CariocaCuiPresenter)
	m, _ := setupCariocaCuiMock(domain.CariocaPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
