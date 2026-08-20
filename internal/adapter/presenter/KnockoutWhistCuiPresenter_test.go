//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeKnockoutWhistPlayers() []*domain.KnockoutWhistPlayer {
	return []*domain.KnockoutWhistPlayer{
		domain.NewKnockoutWhistPlayer(true),
		domain.NewKnockoutWhistPlayer(false),
		domain.NewKnockoutWhistPlayer(false),
		domain.NewKnockoutWhistPlayer(false),
	}
}

func setupKnockoutWhistCuiMock() *interfaces.MockKnockoutWhistGame {
	m := new(interfaces.MockKnockoutWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetHandSize").Return(7)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetLeadPlayerIdx").Return(-1).Maybe()
	m.On("GetRoundWinnerIdx").Return(-1).Maybe()
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetActiveCount").Return(4)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupKnockoutWhistCuiMockWithPlayers() (*interfaces.MockKnockoutWhistGame, []*domain.KnockoutWhistPlayer) {
	m := setupKnockoutWhistCuiMock()
	players := makeKnockoutWhistPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKnockoutWhistCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KnockoutWhistCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupKnockoutWhistCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("eliminated player line", func(t *testing.T) {
		m, players := setupKnockoutWhistCuiMockWithPlayers()
		players[1].SetEliminated(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KnockoutWhistPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KnockoutWhistPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestKnockoutWhistCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KnockoutWhistCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.KnockoutWhistHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupKnockoutWhistCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.KnockoutWhistHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.KnockoutWhistHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestKnockoutWhistCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KnockoutWhistCuiPresenter)
	m := new(interfaces.MockKnockoutWhistGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewKnockoutWhistPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

func TestKnockoutWhistCuiPresenter_TrumpSelectPrompt(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KnockoutWhistCuiPresenter)

	m, _ := setupKnockoutWhistCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.KnockoutWhistPhaseTrumpSelect)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
	m.On("GetLeadPlayerIdx").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "st <1-4>") // help line advertising the select-trump command
}

// **リード権と直前ラウンドの勝者が CUI に出ていなかった (#4762)。**Web は
// leader / roundWinner バッジを常時出しており、SpoilFive の CUI にも同じ
// 目的のマークがある。
func TestKnockoutWhistCuiPresenter_LeaderAndWinnerMarks(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KnockoutWhistCuiPresenter)

	withSeats := func(lead, winner int) *interfaces.MockKnockoutWhistGame {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetLeadPlayerIdx").Return(lead)
		m.On("GetRoundWinnerIdx").Return(winner)
		return m
	}

	t.Run("marks the lead seat once", func(t *testing.T) {
		out := p.Output(withSeats(1, -1), nil)
		assert.Equal(t, 1, strings.Count(out, "▶(リード)"), "印が付くのは1席だけ")
	})

	t.Run("marks the round winner once", func(t *testing.T) {
		out := p.Output(withSeats(-1, 2), nil)
		assert.Equal(t, 1, strings.Count(out, "★(勝者)"))
	})

	// **同じ席がリードかつ勝者になることがある。**片方で上書きしない。
	t.Run("marks a seat that is both lead and round winner", func(t *testing.T) {
		out := p.Output(withSeats(0, 0), nil)
		assert.Contains(t, out, "▶(リード)")
		assert.Contains(t, out, "★(勝者)")
	})

	t.Run("marks nothing before either is decided", func(t *testing.T) {
		out := p.Output(withSeats(-1, -1), nil)
		assert.NotContains(t, out, "▶(リード)")
		assert.NotContains(t, out, "★(勝者)")
	})
}

// #5650: Knockout Whist は毎ラウンド手札が1枚ずつ減り、最後は1枚勝負になる。
// Web は kw-next-round-preview で「次は何枚か / 切り札を選ぶのは誰か」を予告して
// いるのに、CUI は生存者数しか出しておらず、次ラウンドの形が分からなかった。
func TestKnockoutWhistCuiPresenter_RoundEndPreviewsTheNextRound(t *testing.T) {
	p := new(presenter.KnockoutWhistCuiPresenter)

	roundEndMock := func(handSize, winner int) *interfaces.MockKnockoutWhistGame {
		m, _ := setupKnockoutWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandSize")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetPhase").Return(domain.KnockoutWhistPhaseRoundEnd)
		m.On("GetHandSize").Return(handSize)
		m.On("GetRoundWinnerIdx").Return(winner)
		return m
	}

	t.Run("announces the smaller hand and who picks trump", func(t *testing.T) {
		out := p.Output(roundEndMock(7, 0), nil)

		assert.Contains(t, out, i18n.Tf("knockoutwhist.nextRoundPreview",
			"count", "6", "name", color.Bold(i18n.T("cuiPlayerYou"))))
	})

	// **手札1枚は最終ラウンド。**同じ文言だと「次もまだ続く」と読めてしまう。
	t.Run("calls the one-card round final", func(t *testing.T) {
		out := p.Output(roundEndMock(2, 1), nil)

		assert.Contains(t, out, i18n.Tf("knockoutwhist.finalRoundPreview",
			"count", "1", "name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))))
		assert.NotContains(t, out, i18n.Tf("knockoutwhist.nextRoundPreview",
			"count", "1", "name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))))
	})

	// 手札は1枚より下がらない。1枚ラウンドの後も1枚のまま予告する。
	t.Run("never previews fewer than one card", func(t *testing.T) {
		out := p.Output(roundEndMock(1, 0), nil)

		assert.Contains(t, out, i18n.Tf("knockoutwhist.finalRoundPreview",
			"count", "1", "name", color.Bold(i18n.T("cuiPlayerYou"))))
	})

	// 勝者が未確定なら誰が切り札を選ぶか言えない。
	t.Run("says nothing before a round winner is known", func(t *testing.T) {
		out := p.Output(roundEndMock(7, -1), nil)

		prefix, _, ok := strings.Cut(i18n.Tf("knockoutwhist.nextRoundPreview",
			"count", "\x00", "name", "x"), "\x00")
		require.True(t, ok)
		assert.NotContains(t, out, prefix)
	})
}
