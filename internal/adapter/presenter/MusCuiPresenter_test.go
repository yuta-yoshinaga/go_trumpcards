//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeMusPlayers() []*domain.MusPlayer {
	return []*domain.MusPlayer{
		domain.NewMusPlayer(true),
		domain.NewMusPlayer(false),
		domain.NewMusPlayer(false),
		domain.NewMusPlayer(false),
	}
}

func setupMusCuiMock() *interfaces.MockMusGame {
	m := new(interfaces.MockMusGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetAmarrakos").Return([domain.MusTeamCnt]int{0, 0})
	m.On("GetPhase").Return(domain.MusPhaseMus)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetMusTurn").Return(0)
	m.On("GetDiscardTurn").Return(0)
	m.On("GetBetTeam").Return(0)
	m.On("GetPendingStake").Return(0)
	m.On("GetLastBettorTeam").Return(-1)
	m.On("GetManoIdx").Return(0)
	m.On("GetMusCycle").Return(0)
	for ri := 0; ri < domain.MusRoundCnt; ri++ {
		m.On("GetResult", ri).Return(domain.MusRoundResult{Team: -1})
	}
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMusCuiMockWithPlayers() (*interfaces.MockMusGame, []*domain.MusPlayer) {
	m := setupMusCuiMock()
	players := makeMusPlayers()
	m.On("GetPlayerCnt").Return(domain.MusPlayerCnt)
	m.On("GetHandSummary", mock.Anything).Return((*domain.MusHandSummary)(nil)).Maybe()
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestMusCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MusCuiPresenter)

	t.Run("mus phase shows prompt", func(t *testing.T) {
		m, players := setupMusCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Mus")
		assert.NotEmpty(t, result)
	})

	t.Run("discard phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("grande betting phase shows bet prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("juego betting phase shows bet prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseJuego)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("showdown phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseShowdown)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("betting results show localized kind names, not raw ints", func(t *testing.T) {
		// Each result kind renders by its localized name, covering every branch.
		for _, tc := range []struct {
			kind  int
			label string
		}{
			{domain.MusResultDeferred, i18n.T("mus.resultDeferred")},
			{domain.MusResultAccepted, i18n.T("mus.resultAccepted")},
			{domain.MusResultAwarded, i18n.T("mus.resultAwarded")},
			{domain.MusResultOrdago, i18n.T("mus.resultOrdago")},
		} {
			m, _ := setupMusCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetResult")
			m.On("GetPhase").Return(domain.MusPhaseChica)
			for ri := 0; ri < domain.MusRoundCnt; ri++ {
				m.On("GetResult", ri).Return(domain.MusRoundResult{Kind: tc.kind, Stake: 1, Team: -1})
			}
			result := p.Output(m, nil)
			assert.Contains(t, result, tc.label)
			// A raw kind int must not leak into the result line.
			assert.NotContains(t, result, "種別="+strconv.Itoa(tc.kind))
		}
	})
}

func TestMusCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MusCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MusHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("mus phase hint - exchange", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MusHint{Mus: true, Reason: "mus_exchange"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("mus phase hint - cut", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MusHint{Mus: false, Reason: "mus_cut"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("discard phase hint with indices", func(t *testing.T) {
		m, players := setupMusCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		m.On("GetHint").Return(&domain.MusHint{Indices: []int{1}, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("discard phase hint no cards to discard", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		m.On("GetHint").Return(&domain.MusHint{Indices: []int{}, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("bet hint - paso", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		m.On("GetHint").Return(&domain.MusHint{Action: domain.MusActionPaso, Reason: "bet_paso"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("bet hint - envido", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhasePares)
		m.On("GetHint").Return(&domain.MusHint{Action: domain.MusActionEnvido, Amount: 2, Reason: "bet_envido"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestMusCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MusCuiPresenter)
	m := new(interfaces.MockMusGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "mus", Detail: "You wants mus"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "mus")
}

// #5640: 4 つの賭けラウンドはそれぞれ手札の別々の側面を見ている。Web は
// mus-hand-summary で 4 項目を常時出しているのに、CUI は札を並べるだけで、
// Mus 独自のランク付け (A/K が高位、2/3 が低位) を暗算させていた。
func TestMusCuiPresenter_ShowsTheHandSummary(t *testing.T) {
	p := new(presenter.MusCuiPresenter)
	summary := &domain.MusHandSummary{
		HighestRank:   10,
		LowestRank:    9,
		ParesCategory: domain.MusParesDuples,
		Points:        40,
		HasJuego:      true,
	}

	t.Run("prints all four rounds worth of evaluation", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandSummary")
		m.On("GetHandSummary", mock.Anything).Return(summary)

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("mus.summaryLine",
			"label", i18n.T("mus.summaryLabel"),
			"grande", strconv.Itoa(summary.HighestRank),
			"chica", strconv.Itoa(summary.LowestRank),
			"pares", i18n.T("mus.paresDuples"),
			"juego", i18n.Tf("mus.juegoYes", "points", strconv.Itoa(summary.Points))))
	})

	t.Run("says Punto when the hand is short of the threshold", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandSummary")
		m.On("GetHandSummary", mock.Anything).Return(&domain.MusHandSummary{
			HighestRank: 4, LowestRank: 1, ParesCategory: domain.MusParesNone, Points: 10,
		})

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("mus.juegoPunto", "points", "10"))
		assert.Contains(t, out, i18n.T("mus.paresNone"))
	})

	// ちょうど 31 は Juego で最強。Web も juegoBest として別扱いにしている。
	t.Run("calls exactly 31 the best juego", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandSummary")
		m.On("GetHandSummary", mock.Anything).Return(&domain.MusHandSummary{
			HighestRank: 10, LowestRank: 1, ParesCategory: domain.MusParesMedias,
			Points: domain.MusJuegoThreshold, HasJuego: true,
		})

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.T("mus.juegoBest"))
		assert.NotContains(t, out, i18n.Tf("mus.juegoYes", "points", "31"))
		// 3 枚同ランクは medias。分類ごとにラベルが違うことも同時に固定する。
		assert.Contains(t, out, i18n.T("mus.paresMedias"))
		assert.NotContains(t, out, i18n.T("mus.paresPar")+" ")
	})

	t.Run("prints nothing when the hand cannot be evaluated", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers() // default GetHandSummary = nil

		out := p.Output(m, nil)

		assert.NotContains(t, out, i18n.T("mus.summaryLabel"))
	})
}
