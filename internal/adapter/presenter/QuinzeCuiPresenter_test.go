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

func setupQuinzeCuiMockDefaults(g *interfaces.MockQuinzeGame) {
	g.On("GetPhase").Return(domain.QuinzePhasePlayerTurn).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetBankerIdx").Return(1).Maybe()
	g.On("IsHumanBanker").Return(false).Maybe()
	g.On("GetActiveSeat").Return(0).Maybe()
	g.On("GetNextBanker").Return(-1).Maybe()
	g.On("GetLastResult").Return("親は 6.5").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("CanHit").Return(true).Maybe()
	g.On("CanStand").Return(true).Maybe()
	g.On("GetHandPoints", mock.Anything).Return(9).Maybe()
	g.On("FormatPoints", mock.Anything).Return("4.5").Maybe()
	g.On("GetSeats").Return([]*domain.QuinzeSeat{
		quinzeSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":[{"d":1,"v":4,"f":true}],"bt":100}}`),
		quinzeSeatFromJSON(`{"nm":"CPU1","cp":true}`),
		quinzeSeatFromJSON(`{"nm":"CPU2","cp":true,"hd":{"cd":[{"d":3,"v":5,"f":true}],"bt":20}}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		quinzeHandFromJSON(`{"cd":[{"d":2,"v":5,"f":true}]}`)).Maybe()
}

func TestQuinzeCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("chips and the banker lead the view", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)

		out := new(QuinzeCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "900")
		assert.Contains(t, out, "CPU1")
	})

	t.Run("the banker's card stays hidden until the round settles", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)

		out := new(QuinzeCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("quinze.faceDown"))
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.On("GetPhase").Return(domain.QuinzePhaseEnd)
		g.On("GetGameEndFlag").Return(true)

		out := new(QuinzeCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, i18n.T("quinze.faceDown"))
		assert.Contains(t, out, "親は 6.5")
	})

	t.Run("only the legal actions are listed", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)

		out := new(QuinzeCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("quinze.optHit"))
		assert.Contains(t, out, i18n.T("quinze.optStand"))
	})

	t.Run("no actions leaves the line out entirely", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)
		for _, name := range []string{"CanHit", "CanStand"} {
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, name)
			g.On(name).Return(false)
		}
		assert.NotContains(t, new(QuinzeCuiPresenter).Output(g, nil), i18n.T("quinze.actionsLine"))
	})

	for _, tc := range []struct {
		name        string
		phase       int
		humanBanker bool
		want        string
	}{
		{"betting", domain.QuinzePhaseBet, false, i18n.T("quinze.placeBet")},
		{"betting while banking", domain.QuinzePhaseBet, true, i18n.T("quinze.dealAsBanker")},
		{"banker turn", domain.QuinzePhaseBankerTurn, true, i18n.T("quinze.bankerTurn")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockQuinzeGame)
			setupQuinzeCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanBanker").Return(tc.humanBanker)

			assert.Contains(t, new(QuinzeCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("the bank passing is announced", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetNextBanker")
		g.On("GetPhase").Return(domain.QuinzePhaseEnd)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetNextBanker").Return(0)

		assert.Contains(t, new(QuinzeCuiPresenter).Output(g, nil), "あなた")
	})

	t.Run("the human banker is named as such", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerIdx")
		g.On("IsHumanBanker").Return(true)
		g.On("GetBankerIdx").Return(0)

		assert.Contains(t, new(QuinzeCuiPresenter).Output(g, nil), i18n.T("quinze.bankerIsYou"))
	})

	t.Run("error block", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		setupQuinzeCuiMockDefaults(g)
		assert.Contains(t, new(QuinzeCuiPresenter).Output(g, assertError{}), "boom")
	})
}

func TestQuinzeCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round hides the log", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		g.On("GetGameEndFlag").Return(false)
		assert.NotContains(t, new(QuinzeCuiPresenter).ActionLogOutput(g), "deal")
	})

	t.Run("a settled round shows the log", func(t *testing.T) {
		g := new(interfaces.MockQuinzeGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test detail"},
		})
		assert.Contains(t, new(QuinzeCuiPresenter).ActionLogOutput(g), "test detail")
	})
}

// #5566: 相手がいつ引くのをやめるかは、賭け続けるかの判断材料。12 点という
// 数字はどの画面にも出ていなかった。
func TestQuinzeCuiPresenter_ShowsTheCpuStandThreshold(t *testing.T) {
	g := new(interfaces.MockQuinzeGame)
	setupQuinzeCuiMockDefaults(g)
	// **半点をそのまま出さない。**閾値は 11 だが、画面に出るのは 12 点。
	// ドメインの FormatPoints を通していることを、引数ごとの戻り値で確かめる
	// (既定の総括登録を外してから、引数ごとに積む)。
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "FormatPoints")
	g.On("FormatPoints", domain.QuinzeCpuStandPoints).Return("12").Maybe()
	g.On("FormatPoints", domain.QuinzeTarget).Return("15").Maybe()
	g.On("FormatPoints", mock.Anything).Return("4.5").Maybe()

	out := new(QuinzeCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.Tf("quinze.cpuStandLine", "total", "12", "target", "15"))
	assert.NotContains(t, out, i18n.Tf("quinze.cpuStandLine", "total", "11", "target", "15"))
}

// 打てる手が無い局面では出さない。手番でもないのに停止ラインだけ残ると、
// 誰の話をしているのか分からない。
func TestQuinzeCuiPresenter_HidesTheThresholdWithNoLegalAction(t *testing.T) {
	g := new(interfaces.MockQuinzeGame)
	setupQuinzeCuiMockDefaults(g)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "FormatPoints")
	g.On("FormatPoints", domain.QuinzeCpuStandPoints).Return("12").Maybe()
	g.On("FormatPoints", domain.QuinzeTarget).Return("15").Maybe()
	g.On("FormatPoints", mock.Anything).Return("4.5").Maybe()
	for _, name := range []string{"CanHit", "CanStand"} {
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, name)
		g.On(name).Return(false)
	}
	assert.NotContains(t, new(QuinzeCuiPresenter).Output(g, nil),
		i18n.Tf("quinze.cpuStandLine", "total", "12", "target", "15"))
}
