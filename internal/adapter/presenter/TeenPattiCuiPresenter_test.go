//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func tpMakePlayers() []*domain.TeenPattiPlayer {
	return []*domain.TeenPattiPlayer{
		domain.NewTeenPattiPlayer(true, 30),
		domain.NewTeenPattiPlayer(false, 30),
		domain.NewTeenPattiPlayer(false, 30),
		domain.NewTeenPattiPlayer(false, 30),
	}
}

func tpSetupBaseMock() *interfaces.MockTeenPattiGame {
	m := new(interfaces.MockTeenPattiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPot").Return(4)
	m.On("GetStake").Return(1)
	m.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetMatchWinnerIdx").Return(-1)
	m.On("IsShowdown").Return(false)
	m.On("GetSideShowRequester").Return(-1)
	m.On("GetSideShowTarget").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// #4729: レイズ可能域をドメインから引くようになった。
	m.On("GetRaiseRange", 0).Return(2, 100, true).Maybe()

	return m
}

func tpSetupMockWithPlayers() (*interfaces.MockTeenPattiGame, []*domain.TeenPattiPlayer) {
	m := tpSetupBaseMock()
	players := tpMakePlayers()
	m.On("GetPlayerCnt").Return(domain.TeenPattiPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestTeenPattiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TeenPattiCuiPresenter)

	t.Run("betting phase shows prompt and human cards", func(t *testing.T) {
		m, players := tpSetupMockWithPlayers()
		players[0].SetSeen(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "SPADE")
	})

	t.Run("betting raise range: blind human sees full-chip ceiling", func(t *testing.T) {
		m, players := tpSetupMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// **計算はドメインの仕事になった (#4729)。**ここで確かめるのは
		// 「返ってきた範囲をそのまま出すこと」だけ。Seen で上限が半分になる等の
		// 式の正しさは TestTeenPatti_GetRaiseRange が実際のプレイヤーで見ている。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRaiseRange")
		m.On("GetRaiseRange", 0).Return(2, 30, true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "レイズ可能額: 2〜30")
	})

	t.Run("betting raise range: seen human ceiling is halved", func(t *testing.T) {
		m, players := tpSetupMockWithPlayers()
		players[0].SetSeen(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRaiseRange")
		m.On("GetRaiseRange", 0).Return(2, 15, true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "レイズ可能額: 2〜15")
	})

	t.Run("betting raise range: unavailable when chips are too low", func(t *testing.T) {
		m := tpSetupBaseMock()
		poor := domain.NewTeenPattiPlayer(true, 1) // 1 chip, stake 1 → min 2 > max 1
		players := []*domain.TeenPattiPlayer{poor, tpMakePlayers()[1], tpMakePlayers()[2], tpMakePlayers()[3]}
		m.On("GetPlayerCnt").Return(domain.TeenPattiPlayerCnt)
		for i, pl := range players {
			m.On("GetPlayer", i).Return(pl)
		}
		// **base の .Maybe() より先に評価されない。**testify は最初に一致した
		// 登録を使うので、tpSetupBaseMock 側の既定を外してから入れ直す。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRaiseRange")
		m.On("GetRaiseRange", 0).Return(2, 1, false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "チップ不足のためレイズできません")
	})

	t.Run("betting raise range is omitted on a CPU turn", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1) // a CPU seat
		result := p.Output(m, nil)
		assert.NotContains(t, result, "レイズ可能額")
		assert.NotContains(t, result, "チップ不足")
	})

	t.Run("side show phase shows player names, not raw indices", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m, _ := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseSideShow)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowRequester")
		m.On("GetSideShowRequester").Return(2)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowTarget")
		m.On("GetSideShowTarget").Return(1) // CPU target
		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 2") // requester rendered by name
		assert.Contains(t, result, "CPU 1") // target rendered by name
		// Target is a CPU, so the "you are challenged" clarification is omitted.
		assert.NotContains(t, result, "accept or decline")
	})

	t.Run("side show targeting the human shows a respond clarification", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m, _ := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseSideShow)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowRequester")
		m.On("GetSideShowRequester").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowTarget")
		m.On("GetSideShowTarget").Return(0) // human target
		result := p.Output(m, nil)
		assert.Contains(t, result, "accept or decline")
	})

	t.Run("showdown reveals all non-folded hands", func(t *testing.T) {
		m, players := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseShowdown)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsShowdown")
		m.On("IsShowdown").Return(true)
		for i := range players {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		}
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows winner prompt", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("folded and out statuses render", func(t *testing.T) {
		m, players := tpSetupMockWithPlayers()
		players[1].SetFolded(true)
		players[2].SetOut(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTeenPattiCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TeenPattiCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.On("GetHint").Return((*domain.TeenPattiHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("see hint", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.TeenPattiHint{Action: "see", Reason: "see_first"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "see")
	})

	t.Run("raise hint", func(t *testing.T) {
		m, _ := tpSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.TeenPattiHint{Action: "raise", Reason: "strong_hand"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "raise")
	})
}

func TestTeenPattiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TeenPattiCuiPresenter)
	m := new(interfaces.MockTeenPattiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "You bets 1"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewTeenPattiPlayer(true, 0)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "bet")
}
