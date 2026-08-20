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

func setupMachiavelliCuiMock(phase domain.MachiavelliPhase, gameEnd bool, table [][]*domain.Card) (*interfaces.MockMachiavelliGame, []*domain.MachiavelliPlayer) {
	m := new(interfaces.MockMachiavelliGame)
	players := []*domain.MachiavelliPlayer{
		domain.NewMachiavelliPlayer(true),
		domain.NewMachiavelliPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(52)
	m.On("GetTable").Return(table)
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultMachiavelliConfig())
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestMachiavelliCuiPresenter_Output(t *testing.T) {
	p := new(presenter.MachiavelliCuiPresenter)
	table := [][]*domain.Card{{
		domain.NewCard(domain.CardDesignSpade, 4, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
	}}

	t.Run("turn phase with table", func(t *testing.T) {
		m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseTurn, false, table)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "マキャヴェッリ")
		assert.Contains(t, out, "ラウンド")
		assert.Contains(t, out, "場のメルド")
		// With melds on the table, the layoff-help format line is shown.
		assert.Contains(t, out, strings.Split(i18n.T("machiavelli.promptLayoffHelp"), "{{")[0])
		assert.NotContains(t, out, i18n.T("machiavelli.promptLayoffNone"))
	})

	t.Run("turn phase empty table", func(t *testing.T) {
		m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseTurn, false, nil)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		// With no melds, layoff is announced as unavailable.
		assert.Contains(t, out, i18n.T("machiavelli.promptLayoffNone"))
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseRoundEnd, false, table)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseGameEnd, true, table)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseTurn, false, table)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})
}

func TestMachiavelliCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MachiavelliCuiPresenter)
	m, _ := setupMachiavelliCuiMock(domain.MachiavelliPhaseTurn, false, nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

// #5704: 場の組み替えはこのゲームの核心なのに、CUI の案内は draw / newmeld /
// layoff だけで、rearrange という手があること自体が伝わっていなかった。
func TestMachiavelliCuiPresenter_MentionsRearrange(t *testing.T) {
	p := new(presenter.MachiavelliCuiPresenter)

	t.Run("offers the rearrange command while a meld is on the table", func(t *testing.T) {
		g := domain.NewDefaultMachiavelli()
		g.Reset()
		g.SetPhase(domain.MachiavelliPhaseTurn)
		g.SetTable([][]*domain.Card{{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		}})

		out := p.Output(g, nil)

		assert.Contains(t, out, i18n.T("machiavelli.promptRearrangeHelp"))
	})

	// 場が空なら組み替える対象が無いので出さない (layoff と同じ扱い)。
	t.Run("stays quiet while the table is empty", func(t *testing.T) {
		g := domain.NewDefaultMachiavelli()
		g.Reset()
		g.SetPhase(domain.MachiavelliPhaseTurn)
		g.SetTable(nil)

		out := p.Output(g, nil)

		assert.NotContains(t, out, i18n.T("machiavelli.promptRearrangeHelp"))
	})
}
