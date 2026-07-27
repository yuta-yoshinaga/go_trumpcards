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

func setupBeziqueCuiMock(trumpCard *domain.Card) *interfaces.MockBeziqueGame {
	m := new(interfaces.MockBeziqueGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BeziquePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetStockRemaining").Return(40)
	m.On("IsEndgame").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	return m
}

func setupBeziqueCuiMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBeziqueGame, []*domain.BeziquePlayer) {
	m := setupBeziqueCuiMock(trumpCard)
	players := []*domain.BeziquePlayer{
		domain.NewBeziquePlayer(true),
		domain.NewBeziquePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetDealPoints", 0).Return(18)
	m.On("GetDealPoints", 1).Return(5)
	m.On("GetDealMeldPoints", 0).Return(8)
	m.On("GetDealMeldPoints", 1).Return(0)
	m.On("GetMatchScore", 0).Return(118)
	m.On("GetMatchScore", 1).Return(45)
	return m, players
}

func TestBeziqueCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.BeziqueCuiPresenter)

	t.Run("initial state shows header, trump, scores", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupBeziqueCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, "Bezique")
		assert.Contains(t, out, "ディール: 1")
		assert.Contains(t, out, "トリック: 1")
		assert.Contains(t, out, "山札: 40枚")
		assert.Contains(t, out, "累積得点: あなた=118  CPU=45")
		assert.Contains(t, out, "play <idx>")
	})

	t.Run("meld phase lists available melds", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, _ := setupBeziqueCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BeziquePhaseMeld)
		m.On("GetAvailableMelds", 0).Return([]domain.BeziqueMeld{
			{Type: domain.BeziqueMeldBezique, Suit: -1, Points: 40},
		})

		out := p.Output(m, nil)
		assert.Contains(t, out, "宣言できる役")
		assert.Contains(t, out, "ベジーク")
		assert.Contains(t, out, "meld <idx>")
	})

	t.Run("meld phase with no melds", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, _ := setupBeziqueCuiMockWithPlayers(trump)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BeziquePhaseMeld)
		m.On("GetAvailableMelds", 0).Return([]domain.BeziqueMeld(nil))

		out := p.Output(m, nil)
		assert.Contains(t, out, "宣言できる役はありません")
	})

	t.Run("round-end prompt", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BeziquePhaseRoundEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, "ディール終了")
	})

	t.Run("endgame phase label", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsEndgame")
		m.On("IsEndgame").Return(true)
		out := p.Output(m, nil)
		assert.Contains(t, out, "第2フェーズ")
	})

	t.Run("error is rendered", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, errors.New("kaboom"))
		assert.Contains(t, out, "kaboom")
	})

	t.Run("game end p0 banner", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.On("GetMatchScore", 0).Return(1010)
		m.On("GetMatchScore", 1).Return(800)

		out := p.Output(m, nil)
		assert.Contains(t, out, "あなたの勝利")
		assert.Contains(t, out, "(1010-800)")
	})

	t.Run("game end p1 banner", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchScore")
		m.On("GetMatchScore", 0).Return(800)
		m.On("GetMatchScore", 1).Return(1010)

		out := p.Output(m, nil)
		assert.Contains(t, out, "CPUの勝利")
	})
}

func TestBeziqueCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BeziqueCuiPresenter)

	t.Run("card hint shows card and reason", func(t *testing.T) {
		m, players := setupBeziqueCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.BeziqueHint{CardIndex: &idx, Reason: "follow_cut"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "切り札でカット")
	})

	t.Run("meld hint", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		idx := 1
		m.On("GetHint").Return(&domain.BeziqueHint{MeldIndex: &idx, Reason: "meld_declare"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "役を宣言")
	})

	t.Run("meld skip hint", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		skip := -1
		m.On("GetHint").Return(&domain.BeziqueHint{MeldIndex: &skip, Reason: "meld_skip"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
	})

	t.Run("hint nil falls back to hintNone", func(t *testing.T) {
		m, _ := setupBeziqueCuiMockWithPlayers(nil)
		m.On("GetHint").Return((*domain.BeziqueHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})
}

func TestBeziqueCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BeziqueCuiPresenter)
	m, _ := setupBeziqueCuiMockWithPlayers(nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotNil(t, out)
}
