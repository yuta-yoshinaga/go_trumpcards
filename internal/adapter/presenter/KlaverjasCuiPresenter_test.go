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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeKlaverjasPlayers() []*domain.KlaverjasPlayer {
	return []*domain.KlaverjasPlayer{
		domain.NewKlaverjasPlayer(true),
		domain.NewKlaverjasPlayer(false),
		domain.NewKlaverjasPlayer(false),
		domain.NewKlaverjasPlayer(false),
	}
}

func setupKlaverjasCuiMock() *interfaces.MockKlaverjasGame {
	m := new(interfaces.MockKlaverjasGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetRoundRoem").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetTeamScores").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupKlaverjasCuiMockWithPlayers() (*interfaces.MockKlaverjasGame, []*domain.KlaverjasPlayer) {
	m := setupKlaverjasCuiMock()
	players := makeKlaverjasPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKlaverjasCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KlaverjasCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupKlaverjasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Klaverjas")
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KlaverjasPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows roem breakdown and total", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundCardPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundRoem")
		m.On("GetPhase").Return(domain.KlaverjasPhaseRoundEnd)
		m.On("GetRoundCardPoints").Return([domain.KlaverjasTeamCnt]int{62, 40})
		m.On("GetRoundRoem").Return([domain.KlaverjasTeamCnt]int{20, 0})
		result := p.Output(m, nil)
		// Team A: 62 card points + 20 Roem = 82 total.
		assert.Contains(t, result, i18n.Tf("klaverjas.promptRoundEndRoem",
			"roemA", "20", "roemB", "0", "totalA", "82", "totalB", "40"))
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestKlaverjasCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KlaverjasCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.KlaverjasHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupKlaverjasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.KlaverjasHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.KlaverjasHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestKlaverjasCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KlaverjasCuiPresenter)
	m := new(interfaces.MockKlaverjasGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// **切り札と非切り札で強さの順序が丸ごと違う。** Web は
// klaverjas-strength-legend の 2 表を常時出しているのに、CUI には説明が一切
// なかった (#5645)。姉妹の Manille は rankHelp で解決済み。
func TestKlaverjasCuiPresenter_ShowsBothStrengthOrders(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KlaverjasCuiPresenter)

	t.Run("the play prompt carries both orders", func(t *testing.T) {
		m, players := setupKlaverjasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, nil)

		// **ドメインの配点と一致していること。** 表は写さず、Klaverjas.cardPoints /
		// trumpStrength を読んで書いた値をここでも明示する。
		assert.Contains(t, out, "J > 9 > A > 10 > K > Q > 8 > 7") // 切り札
		assert.Contains(t, out, "A > 10 > K > Q > J > 9 > 8 > 7") // 非切り札
		assert.Contains(t, out, "J=20")
		assert.Contains(t, out, "9=14")
		assert.Contains(t, out, "J=2") // 非切り札の J は 2 点

		// 2 系統あることが伝わらないと意味がない。片方だけでは不足。
		assert.Contains(t, out, i18n.T("klaverjas.rankHelpTrump"))
		assert.Contains(t, out, i18n.T("klaverjas.rankHelpPlain"))
		assert.NotEqual(t, i18n.T("klaverjas.rankHelpTrump"), i18n.T("klaverjas.rankHelpPlain"))
	})

	// **負のコントロール: 打つ局面でないなら出さない。** ラウンド終了の集計に
	// 強さ表が混ざると読みにくいだけ。
	t.Run("other phases do not repeat the tables", func(t *testing.T) {
		m, _ := setupKlaverjasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KlaverjasPhaseTrickEnd)
		out := p.Output(m, nil)
		assert.NotContains(t, out, i18n.T("klaverjas.rankHelpTrump"))
	})

	t.Run("english locale keeps both orders", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m, players := setupKlaverjasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, nil)
		assert.Contains(t, out, "Trump order:")
		assert.Contains(t, out, "Non-trump:")
		// ja が漏れていないこと (T() は未翻訳ならキーを返すので、キー比較だけでは
		// 翻訳の有無を検査できない)。
		assert.NotContains(t, out, "切り札の強さ")
	})
}
