//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPontoonCuiMockDefaults(g *interfaces.MockPontoonGame) {
	g.On("GetPhase").Return(domain.PontoonPhasePlayerTurn).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetBankerIdx").Return(1).Maybe()
	g.On("IsHumanBanker").Return(false).Maybe()
	g.On("GetActiveSeat").Return(0).Maybe()
	g.On("GetActiveHand").Return(0).Maybe()
	g.On("GetNextBanker").Return(-1).Maybe()
	g.On("GetLastResult").Return("親は 19").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("CanStick").Return(true).Maybe()
	g.On("CanTwist").Return(true).Maybe()
	g.On("CanBuy").Return(false).Maybe()
	g.On("CanSplit").Return(false).Maybe()
	g.On("GetHandTotal", mock.Anything).Return(18).Maybe()
	g.On("GetHandRank", mock.Anything).Return(domain.PontoonRankPoints).Maybe()
	g.On("GetSeats").Return([]*domain.PontoonSeat{
		pontoonSeatFromJSON(`{"nm":"あなた","cp":false,"hd":[{"cd":[{"d":1,"v":10,"f":true},{"d":2,"v":8,"f":true}],"bt":100}]}`),
		pontoonSeatFromJSON(`{"nm":"CPU1","cp":true}`),
		pontoonSeatFromJSON(`{"nm":"CPU2","cp":true,"hd":[{"cd":[{"d":3,"v":9,"f":true},{"d":4,"v":7,"f":true}],"bt":20}]}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		pontoonHandFromJSON(`{"cd":[{"d":1,"v":10,"f":true},{"d":2,"v":9,"f":true}]}`)).Maybe()
}

func TestPontoonCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")

	// The banker's hand is the one you cannot see; that is the whole difference
	// from blackjack, so the view must not leak it mid-round.
	t.Run("the banker's cards stay hidden until the round settles", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)

		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("pontoon.faceDown"))
		assert.Contains(t, out, i18n.T("pontoon.bankerHandHeader"))
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.On("GetPhase").Return(domain.PontoonPhaseEnd)
		g.On("GetGameEndFlag").Return(true)

		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, i18n.T("pontoon.faceDown"))
		assert.Contains(t, out, "親は 19")
	})

	t.Run("chips and the banker lead the view", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)

		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "900")
		assert.Contains(t, out, "CPU1")
	})

	// Only the legal declarations are offered -- sticking is not suggested below
	// fifteen, and buying is not suggested once the hand has twisted.
	t.Run("only the legal actions are listed", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)

		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("pontoon.optStick"))
		assert.Contains(t, out, i18n.T("pontoon.optTwist"))
		assert.NotContains(t, out, i18n.T("pontoon.optBuy"))
		assert.NotContains(t, out, i18n.T("pontoon.optSplit"))
	})

	t.Run("no actions leaves the line out entirely", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		for _, name := range []string{"CanStick", "CanTwist", "CanBuy", "CanSplit"} {
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, name)
			g.On(name).Return(false)
		}
		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, i18n.T("pontoon.actionsLine"))
	})

	t.Run("split is offered when it is legal", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanSplit")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanBuy")
		g.On("CanSplit").Return(true)
		g.On("CanBuy").Return(true)

		out := new(PontoonCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("pontoon.optSplit"))
		assert.Contains(t, out, i18n.T("pontoon.optBuy"))
	})

	for _, tc := range []struct {
		name        string
		phase       int
		humanBanker bool
		want        string
	}{
		{"betting", domain.PontoonPhaseBet, false, i18n.T("pontoon.placeBet")},
		{"betting while banking", domain.PontoonPhaseBet, true, i18n.T("pontoon.dealAsBanker")},
		{"banker turn", domain.PontoonPhaseBankerTurn, true, i18n.T("pontoon.bankerTurn")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockPontoonGame)
			setupPontoonCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanBanker").Return(tc.humanBanker)

			assert.Contains(t, new(PontoonCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("the bank passing is announced", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetNextBanker")
		g.On("GetPhase").Return(domain.PontoonPhaseEnd)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetNextBanker").Return(0)

		assert.Contains(t, new(PontoonCuiPresenter).Output(g, nil),
			strings.Split(i18n.Tf("pontoon.bankPasses", "name", "あなた"), "{{")[0])
	})

	t.Run("the human banker is named as such", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerIdx")
		g.On("IsHumanBanker").Return(true)
		g.On("GetBankerIdx").Return(0)

		assert.Contains(t, new(PontoonCuiPresenter).Output(g, nil), i18n.T("pontoon.bankerIsYou"))
	})

	t.Run("error block", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonCuiMockDefaults(g)
		assert.Contains(t, new(PontoonCuiPresenter).Output(g, assertError{}), "boom")
	})
}

func TestPontoonCuiPresenter_RankLabels(t *testing.T) {
	i18n.SetLang("ja")
	tests := []struct {
		rank domain.PontoonRank
		want string
	}{
		{domain.PontoonRankPontoon, i18n.T("pontoon.rankPontoon")},
		{domain.PontoonRankFiveCard, i18n.T("pontoon.rankFiveCard")},
		{domain.PontoonRankBust, i18n.T("pontoon.rankBust")},
		{domain.PontoonRankPoints, ""},
	}
	for _, tt := range tests {
		if got := pontoonRankLabel(tt.rank); got != tt.want {
			t.Errorf("pontoonRankLabel(%v) = %q, want %q", tt.rank, got, tt.want)
		}
	}
}

func TestPontoonCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round hides the log", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		g.On("GetGameEndFlag").Return(false)
		assert.NotContains(t, new(PontoonCuiPresenter).ActionLogOutput(g), "deal")
	})

	t.Run("a settled round shows the log", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test detail"},
		})
		assert.Contains(t, new(PontoonCuiPresenter).ActionLogOutput(g), "test detail")
	})
}
