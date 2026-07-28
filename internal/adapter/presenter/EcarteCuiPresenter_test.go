//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupEcarteCuiMock(trumpCard *domain.Card) *interfaces.MockEcarteGame {
	m := new(interfaces.MockEcarteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.EcartePhasePlay)
	m.On("GetNegStep").Return(domain.EcarteNegElderDecide)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetStockRemaining").Return(21)
	m.On("GetWinnerIdx").Return(-1)
	return m
}

func setupEcarteCuiMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockEcarteGame, []*domain.EcartePlayer) {
	m := setupEcarteCuiMock(trumpCard)
	players := []*domain.EcartePlayer{
		domain.NewEcartePlayer(true),
		domain.NewEcartePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetDealPoints", 0).Return(2)
	m.On("GetDealPoints", 1).Return(1)
	m.On("GetMatchScore", 0).Return(4)
	m.On("GetMatchScore", 1).Return(3)
	return m, players
}

func TestEcarteCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.EcarteCuiPresenter)

	t.Run("initial play-phase state shows header, trump, scores", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupEcarteCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, "carté")
		assert.Contains(t, out, "ディール: 1")
		assert.Contains(t, out, "トリック: 1")
		assert.Contains(t, out, "山札: 21枚")
		assert.Contains(t, out, "累積得点: あなた=4  CPU=3")
		assert.Contains(t, out, "play <idx>")
	})

	t.Run("exchange phase elder-decide prompt", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, _ := setupEcarteCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EcartePhaseExchange)

		out := p.Output(m, nil)
		assert.Contains(t, out, "交換を提案")
		assert.Contains(t, out, "勝負する")
	})

	t.Run("exchange phase dealer-respond prompt", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, _ := setupEcarteCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EcartePhaseExchange)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetNegStep")
		m.On("GetNegStep").Return(domain.EcarteNegDealerRespond)

		out := p.Output(m, nil)
		assert.Contains(t, out, "承諾")
		assert.Contains(t, out, "拒否")
	})

	t.Run("exchange phase discard prompt shows the limit (hand < stock)", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupEcarteCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EcartePhaseExchange)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetNegStep")
		m.On("GetNegStep").Return(domain.EcarteNegElderDiscard)
		// Elder (seat 0) holds 5 cards; stock is 21 → the hand caps the discard at 5.
		for range 5 {
			players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		}

		out := p.Output(m, nil)
		assert.Contains(t, out, "discard <idx")
		assert.Contains(t, out, "最大5枚")
		assert.Contains(t, out, "山札残21")
	})

	t.Run("exchange phase discard limit is capped by the stock (stock < hand)", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupEcarteCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EcartePhaseExchange)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetNegStep")
		m.On("GetNegStep").Return(domain.EcarteNegDealerDiscard)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetStockRemaining")
		m.On("GetStockRemaining").Return(2)
		for range 5 {
			players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		}

		out := p.Output(m, nil)
		assert.Contains(t, out, "最大2枚") // stock (2) < hand (5)
		assert.Contains(t, out, "山札残2")
	})

	t.Run("trump-none line when stock exhausted", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(nil)
		out := p.Output(m, nil)
		assert.Contains(t, out, "スペード")
	})

	t.Run("round-end prompt", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EcartePhaseRoundEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, "ディール終了")
	})

	t.Run("error is rendered", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, errors.New("kaboom"))
		assert.Contains(t, out, "kaboom")
	})

	t.Run("game end p0 banner", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.On("GetMatchScore", 0).Return(5)
		m.On("GetMatchScore", 1).Return(3)

		out := p.Output(m, nil)
		assert.Contains(t, out, "あなたの勝利")
		assert.Contains(t, out, "(5-3)")
	})

	t.Run("game end p1 banner", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.On("GetMatchScore", 0).Return(3)
		m.On("GetMatchScore", 1).Return(5)

		out := p.Output(m, nil)
		assert.Contains(t, out, "CPUの勝利")
	})
}

func TestEcarteCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.EcarteCuiPresenter)

	t.Run("card hint shows card and reason", func(t *testing.T) {
		m, players := setupEcarteCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.EcarteHint{CardIndex: &idx, Reason: "follow_win"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "このトリックを取る")
	})

	t.Run("action hint", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.EcarteHint{Action: "propose", Reason: "weak_hand"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "交換を提案")
	})

	t.Run("hint nil falls back to hintNone", func(t *testing.T) {
		m, _ := setupEcarteCuiMockWithPlayers(nil)
		m.On("GetHint").Return((*domain.EcarteHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})
}

func TestEcarteCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.EcarteCuiPresenter)
	m, _ := setupEcarteCuiMockWithPlayers(nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotNil(t, out)
}
