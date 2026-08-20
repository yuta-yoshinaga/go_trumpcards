//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeCalabresellaPlayers() []*domain.CalabresellaPlayer {
	return []*domain.CalabresellaPlayer{
		domain.NewCalabresellaPlayer(true),
		domain.NewCalabresellaPlayer(false),
		domain.NewCalabresellaPlayer(false),
	}
}

func setupCalabresellaCuiMock() *interfaces.MockCalabresellaGame {
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetWinningBid").Return(domain.CalabresellaBidChiamo)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetSoloistIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetRoundThirds").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetPlayerScores").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupCalabresellaCuiMockWithPlayers() (*interfaces.MockCalabresellaGame, []*domain.CalabresellaPlayer) {
	m := setupCalabresellaCuiMock()
	players := makeCalabresellaPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestCalabresellaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupCalabresellaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "カラブレセッラ") // translated helpTitle
		assert.Contains(t, result, "マストフォロー") // play-phase help mentions must-follow
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "ビッド") // translated bid prompt/help
	})

	t.Run("discard phase prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseDiscard)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "モンテ") // translated discard prompt/help
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end lists all players thirds with roles", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundThirds")
		m.On("GetPhase").Return(domain.CalabresellaPhaseRoundEnd)
		// Soloist (index 0) takes 5, the two coalition players 3 each → 11 total.
		m.On("GetRoundThirds").Return([domain.CalabresellaPlayerCnt]int{5, 3, 3})
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("calabresella.roleCoalition"))
		// Each player's thirds appear in the breakdown: soloist 5, coalition 3 each.
		assert.Contains(t, result, "5/3")
		assert.Contains(t, result, "3/3")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

// **モンテは公開情報。**取得した時点で全員に見える札なのに、CUI はどのフェーズでも
// 出しておらず、Web だけが monteLabel パネルに出していた (#4843)。
func TestCalabresellaCuiPresenter_ShowsRevealedMonte(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	monte := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}
	log := []*domain.ActionLogEntry{
		{PlayerIdx: 0, ActionType: "bid", Cards: nil},
		{PlayerIdx: 0, ActionType: "monte_take", Cards: monte},
	}

	t.Run("shown once taken", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return(log)
		result := p.Output(m, nil)
		assert.Contains(t, result, "モンテ（公開）:")
		assert.Contains(t, result, "SPADE 1")
		assert.Contains(t, result, "DIAMOND 13")
	})

	t.Run("hidden during the bid phase", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetActionLog").Return(log)
		m.On("GetPhase").Return(domain.CalabresellaPhaseBid)
		assert.NotContains(t, p.Output(m, nil), "モンテ（公開）:")
	})

	t.Run("hidden before anyone takes it", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		assert.NotContains(t, p.Output(m, nil), "モンテ（公開）:")
	})
}

func TestCalabresellaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.CalabresellaHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupCalabresellaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestCalabresellaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CalabresellaCuiPresenter)
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewCalabresellaPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5688: ソリストはモンテを取って 16 枚になった手札を 12 枚まで捨てる。
// Web は残り枚数をバナーとボタン名で毎レンダー更新しているのに、CUI の
// 固定文言 (「4枚を捨てて12枚にする」) は 1 枚捨てても変わらず、
// 利用者は自分が何枚捨てたかを暗算し続けるしかなかった。
func TestCalabresellaCuiPresenter_DiscardRemaining(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	outputWithHand := func(handSize int) string {
		m, players := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseDiscard)
		for i := 0; i < handSize; i++ {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, i%13+1, false))
		}
		return p.Output(m, nil)
	}

	// 16 枚から 1 枚ずつ捨てるあいだ、残りは 4 → 3 → … と減る。
	for handSize, want := range map[int]int{16: 4, 15: 3, 13: 1} {
		assert.Contains(t, outputWithHand(handSize),
			i18n.Tf("calabresella.promptDiscardRemaining",
				"n", strconv.Itoa(want),
				"size", strconv.Itoa(handSize),
				"target", strconv.Itoa(domain.CalabresellaHandSize)),
			"hand of %d should show %d left", handSize, want)
	}

	t.Run("says nothing once the hand is down to twelve", func(t *testing.T) {
		result := outputWithHand(domain.CalabresellaHandSize)

		assert.NotContains(t, result,
			strings.Split(i18n.T("calabresella.promptDiscardRemaining"), "{{")[0])
	})
}
