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

func makeMariasPlayers() []*domain.MariasPlayer {
	return []*domain.MariasPlayer{
		domain.NewMariasPlayer(true),
		domain.NewMariasPlayer(false),
		domain.NewMariasPlayer(false),
	}
}

func setupMariasCuiMock() *interfaces.MockMariasGame {
	m := new(interfaces.MockMariasGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MariasPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetSoloistIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.MariasPlayerCnt]int{0, 0, 0})
	m.On("GetRoundMarriage").Return([domain.MariasPlayerCnt]int{0, 0, 0}).Maybe()
	m.On("GetPlayerScores").Return([domain.MariasPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMariasCuiMockWithPlayers() (*interfaces.MockMariasGame, []*domain.MariasPlayer) {
	m := setupMariasCuiMock()
	players := makeMariasPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestMariasCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MariasCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupMariasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Mariáš")
		assert.Contains(t, result, "マリッジ") // play-phase help explains the marriage bonus
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MariasPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows defenders total", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundCardPoints")
		m.On("GetPhase").Return(domain.MariasPhaseRoundEnd)
		// Soloist (idx 0) took 40; the two defenders took 30 + 20 = 50.
		m.On("GetRoundCardPoints").Return([domain.MariasPlayerCnt]int{40, 30, 20})
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("marias.promptRoundEndDefenders"), "{{")[0])
		assert.Contains(t, result, "50")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestMariasCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MariasCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MariasHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupMariasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.MariasHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MariasHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestMariasCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MariasCuiPresenter)
	m := new(interfaces.MockMariasGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewMariasPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5647: 結婚 (K+Q) は配札直後に確定し、**ドメインは roundCardPts + roundMarriage
// で勝敗を決める** (Marias.go の settle)。ところが CUI のラウンド終了行はカード点
// しか合算しておらず、結婚点で勝ったソロイストを負けたように読ませていた。
func TestMariasCuiPresenter_RoundEndCountsTheMarriagePoints(t *testing.T) {
	p := new(presenter.MariasCuiPresenter)

	roundEndMock := func(card, marriage [domain.MariasPlayerCnt]int) *interfaces.MockMariasGame {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundCardPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundMarriage")
		m.On("GetPhase").Return(domain.MariasPhaseRoundEnd)
		m.On("GetRoundCardPoints").Return(card)
		m.On("GetRoundMarriage").Return(marriage)
		return m
	}

	// ソロイスト(席0) はカード 45 + 結婚 40 = 85、守備は 45+10=55。カード点だけ
	// 見ると 45 対 45 で引き分けに読めるが、実際はソロイストの勝ち。
	t.Run("adds each side's marriage points to the totals", func(t *testing.T) {
		out := p.Output(roundEndMock(
			[domain.MariasPlayerCnt]int{45, 30, 15},
			[domain.MariasPlayerCnt]int{40, 10, 0},
		), nil)

		assert.Contains(t, out, i18n.Tf("marias.promptRoundEnd",
			"soloist", color.Bold(i18n.T("cuiPlayerYou")), "pts", "85"))
		assert.Contains(t, out, i18n.Tf("marias.promptRoundEndDefenders", "pts", "55"))
	})

	t.Run("unchanged when nobody had a marriage", func(t *testing.T) {
		out := p.Output(roundEndMock(
			[domain.MariasPlayerCnt]int{45, 30, 15},
			[domain.MariasPlayerCnt]int{0, 0, 0},
		), nil)

		assert.Contains(t, out, i18n.Tf("marias.promptRoundEnd",
			"soloist", color.Bold(i18n.T("cuiPlayerYou")), "pts", "45"))
		assert.Contains(t, out, i18n.Tf("marias.promptRoundEndDefenders", "pts", "45"))
	})
}

// Play 中も結婚が成立したことを知らせる。Web は marias-marriage バナーで
// 「{{points}}点獲得」と常時出している。
func TestMariasCuiPresenter_PlayShowsTheMarriageEarned(t *testing.T) {
	p := new(presenter.MariasCuiPresenter)

	playMock := func(marriage [domain.MariasPlayerCnt]int) *interfaces.MockMariasGame {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundMarriage")
		m.On("GetRoundMarriage").Return(marriage)
		return m
	}

	t.Run("announces the human's marriage points", func(t *testing.T) {
		out := p.Output(playMock([domain.MariasPlayerCnt]int{40, 0, 0}), nil)

		assert.Contains(t, out, i18n.Tf("marias.marriageEarned", "points", "40"))
	})

	t.Run("says nothing without a marriage", func(t *testing.T) {
		out := p.Output(playMock([domain.MariasPlayerCnt]int{0, 20, 0}), nil)

		// **点数を伏せた前置きで見る。**別の数字だけを除外すると、0 点のバナーを
		// 出す実装が素通りする (実際に一度素通りした)。
		prefix, _, ok := strings.Cut(i18n.Tf("marias.marriageEarned", "points", "\x00"), "\x00")
		require.True(t, ok)
		assert.NotContains(t, out, prefix)
	})
}
