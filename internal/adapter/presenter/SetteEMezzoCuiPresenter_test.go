//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSemCuiMockDefaults(g *interfaces.MockSetteEMezzoGame) {
	g.On("GetPhase").Return(domain.SetteEMezzoPhasePlayerTurn).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetBankerIdx").Return(1).Maybe()
	g.On("IsHumanBanker").Return(false).Maybe()
	g.On("GetActiveSeat").Return(0).Maybe()
	g.On("GetNextBanker").Return(-1).Maybe()
	g.On("GetLastResult").Return("親は 6.5").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("CanHit").Return(true).Maybe()
	g.On("CanStand").Return(true).Maybe()
	g.On("CanSetMatta").Return(false).Maybe()
	g.On("GetHandHalves", mock.Anything).Return(9).Maybe()
	g.On("FormatHalves", mock.Anything).Return("4.5").Maybe()
	g.On("GetSeats").Return([]*domain.SetteEMezzoSeat{
		semSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":[{"d":1,"v":4,"f":true}],"bt":100}}`),
		semSeatFromJSON(`{"nm":"CPU1","cp":true}`),
		semSeatFromJSON(`{"nm":"CPU2","cp":true,"hd":{"cd":[{"d":3,"v":5,"f":true}],"bt":20}}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		semHandFromJSON(`{"cd":[{"d":2,"v":5,"f":true}]}`)).Maybe()
}

func TestSetteEMezzoCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("chips and the banker lead the view", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)

		out := new(SetteEMezzoCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "900")
		assert.Contains(t, out, "CPU1")
	})

	t.Run("the banker's card stays hidden until the round settles", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)

		out := new(SetteEMezzoCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("settemezzo.faceDown"))
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.On("GetPhase").Return(domain.SetteEMezzoPhaseEnd)
		g.On("GetGameEndFlag").Return(true)

		out := new(SetteEMezzoCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, i18n.T("settemezzo.faceDown"))
		assert.Contains(t, out, "親は 6.5")
	})

	// The matta's current value has to be visible: it is adjustable until the
	// hand stands, and a player cannot choose what they cannot see.
	t.Run("a hand holding the matta shows its current value", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetSeats")
		g.On("GetSeats").Return([]*domain.SetteEMezzoSeat{
			semSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":[{"d":1,"v":4,"f":true},{"d":4,"v":13,"f":true}],"bt":100,"mh":6}}`),
			semSeatFromJSON(`{"nm":"CPU1","cp":true}`),
			semSeatFromJSON(`{"nm":"CPU2","cp":true}`),
		})

		out := new(SetteEMezzoCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "マッタ")
	})

	t.Run("only the legal actions are listed", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)

		out := new(SetteEMezzoCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("settemezzo.optHit"))
		assert.Contains(t, out, i18n.T("settemezzo.optStand"))
		assert.NotContains(t, out, i18n.T("settemezzo.optMatta"))
	})

	t.Run("the matta option appears when the hand holds one", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanSetMatta")
		g.On("CanSetMatta").Return(true)

		assert.Contains(t, new(SetteEMezzoCuiPresenter).Output(g, nil), i18n.T("settemezzo.optMatta"))
	})

	t.Run("no actions leaves the line out entirely", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		for _, name := range []string{"CanHit", "CanStand", "CanSetMatta"} {
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, name)
			g.On(name).Return(false)
		}
		assert.NotContains(t, new(SetteEMezzoCuiPresenter).Output(g, nil), i18n.T("settemezzo.actionsLine"))
	})

	for _, tc := range []struct {
		name        string
		phase       int
		humanBanker bool
		want        string
	}{
		{"betting", domain.SetteEMezzoPhaseBet, false, i18n.T("settemezzo.placeBet")},
		{"betting while banking", domain.SetteEMezzoPhaseBet, true, i18n.T("settemezzo.dealAsBanker")},
		{"banker turn", domain.SetteEMezzoPhaseBankerTurn, true, i18n.T("settemezzo.bankerTurn")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSetteEMezzoGame)
			setupSemCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanBanker").Return(tc.humanBanker)

			assert.Contains(t, new(SetteEMezzoCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("the bank passing is announced", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetNextBanker")
		g.On("GetPhase").Return(domain.SetteEMezzoPhaseEnd)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetNextBanker").Return(0)

		assert.Contains(t, new(SetteEMezzoCuiPresenter).Output(g, nil), "あなた")
	})

	t.Run("the human banker is named as such", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerIdx")
		g.On("IsHumanBanker").Return(true)
		g.On("GetBankerIdx").Return(0)

		assert.Contains(t, new(SetteEMezzoCuiPresenter).Output(g, nil), i18n.T("settemezzo.bankerIsYou"))
	})

	t.Run("error block", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemCuiMockDefaults(g)
		assert.Contains(t, new(SetteEMezzoCuiPresenter).Output(g, assertError{}), "boom")
	})
}

func TestSetteEMezzoCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round hides the log", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		g.On("GetGameEndFlag").Return(false)
		assert.NotContains(t, new(SetteEMezzoCuiPresenter).ActionLogOutput(g), "deal")
	})

	t.Run("a settled round shows the log", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test detail"},
		})
		assert.Contains(t, new(SetteEMezzoCuiPresenter).ActionLogOutput(g), "test detail")
	})
}
