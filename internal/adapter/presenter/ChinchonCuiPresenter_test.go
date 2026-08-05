//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeChinchonPlayers() []*domain.ChinchonPlayer {
	return []*domain.ChinchonPlayer{
		domain.NewChinchonPlayer(true),
		domain.NewChinchonPlayer(false),
	}
}

func setupChinchonCuiMock(phase domain.ChinchonPhase, ended bool, winner int) (*interfaces.MockChinchonGame, []*domain.ChinchonPlayer) {
	m := new(interfaces.MockChinchonGame)
	players := makeChinchonPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerDeadwoodValue", mock.Anything).Return(15)
	m.On("GetPlayerMeldSplit", mock.Anything).Return(([][]*domain.Card)(nil), ([]*domain.Card)(nil)).Maybe()
	m.On("GetKnockThreshold").Return(5)
	return m, players
}

// **どの札が成立しているかは捨て札選びの前提。**Web は緑/破線で色分けし
// 「5 + 3 + 2 = 10」の内訳まで出しているのに、CUI は合計点だけだった (#4838)。
func TestChinchonCuiPresenter_MeldSplit(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ChinchonCuiPresenter)

	t.Run("lists melds and the deadwood breakdown", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDiscard, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		melds := [][]*domain.Card{{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 6, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		}}
		dead := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 2, false),
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerMeldSplit")
		m.On("GetPlayerMeldSplit", 0).Return(melds, dead)
		m.On("GetPlayerMeldSplit", mock.Anything).Return(([][]*domain.Card)(nil), ([]*domain.Card)(nil)).Maybe()

		out := p.Output(m, nil)
		assert.Contains(t, out, "メルド済み: SPADE 5, SPADE 6, SPADE 7")
		assert.Contains(t, out, "デッドウッド: HEART 9, CLOVER 2 (9 + 2 = 11)")
	})

	t.Run("no lines when the split is empty", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDiscard, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		out := p.Output(m, nil)
		assert.NotContains(t, out, "メルド済み:")
		// 既存の deadwoodLine (「デッドウッド: 15点」) とは別物。内訳の括弧が出ないことを見る。
		assert.NotContains(t, out, " = ")
	})
}

func TestChinchonCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ChinchonCuiPresenter)

	t.Run("draw phase shows header and hand", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDraw, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Chinchón (チンチョン)")
		assert.Contains(t, result, "ラウンド: 1")
		// Human deadwood (default 15) is shown but above the threshold → no knock note.
		assert.Contains(t, result, strings.Split(i18n.T("chinchon.deadwoodLine"), "{{")[0])
		assert.NotContains(t, result, i18n.T("chinchon.knockReady"))
	})

	t.Run("human deadwood at threshold shows knock-ready, cpu hidden", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDiscard, false, -1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerDeadwoodValue")
		m.On("GetPlayerDeadwoodValue", 0).Return(3) // human, at/under threshold 5
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false)) // CPU also holds cards
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("chinchon.knockReady"))
		// Only the human's deadwood line appears; the CPU's stays hidden.
		prefix := strings.Split(i18n.T("chinchon.deadwoodLine"), "{{")[0]
		assert.Equal(t, 1, strings.Count(result, prefix))
	})

	t.Run("discard phase prompt", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDiscard, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("layoff phase shows knocker melds", func(t *testing.T) {
		m := new(interfaces.MockChinchonGame)
		players := makeChinchonPlayers()
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.ChinchonPhaseLayoff)
		m.On("GetCurrentPlayerIdx").Return(1)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetKnockerIdx").Return(0)
		m.On("GetKnockerMelds").Return([][]*domain.Card{{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
		}})
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("discard top is displayed", func(t *testing.T) {
		m, _ := setupChinchonCuiMock(domain.ChinchonPhaseDraw, false, -1)
		m.ExpectedCalls = nil
		players := makeChinchonPlayers()
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 7, false))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.ChinchonPhaseDraw)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetKnockerIdx").Return(-1)
		m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札")
	})

	t.Run("eliminated player tagged", func(t *testing.T) {
		m, players := setupChinchonCuiMock(domain.ChinchonPhaseDraw, false, -1)
		players[1].SetEliminated(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "脱落")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupChinchonCuiMock(domain.ChinchonPhaseRoundEnd, false, -1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end with winner", func(t *testing.T) {
		m, _ := setupChinchonCuiMock(domain.ChinchonPhaseGameEnd, true, 0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game end no winner", func(t *testing.T) {
		m, _ := setupChinchonCuiMock(domain.ChinchonPhaseGameEnd, true, -1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "勝者なし")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupChinchonCuiMock(domain.ChinchonPhaseDraw, false, -1)
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestChinchonCuiPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupChinchonCuiMock(domain.ChinchonPhaseDraw, false, -1)
	p := new(presenter.ChinchonCuiPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}
