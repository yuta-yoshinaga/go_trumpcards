//go:build test

package presenter_test

import (
	"errors"
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

func makeTysiacPlayers() []*domain.TysiacPlayer {
	return []*domain.TysiacPlayer{
		domain.NewTysiacPlayer(true),
		domain.NewTysiacPlayer(false),
		domain.NewTysiacPlayer(false),
	}
}

func setupTysiacCuiMock() *interfaces.MockTysiacGame {
	m := new(interfaces.MockTysiacGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TysiacPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(100)
	m.On("GetCurrentBid").Return(100)
	m.On("GetConfig").Return(domain.DefaultTysiacConfig())
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.TysiacPlayerCnt]int{0, 0, 0})
	m.On("GetPlayerScores").Return([domain.TysiacPlayerCnt]int{0, 0, 0})
	m.On("GetMarriageOptions", mock.Anything).Return(([]domain.TysiacMarriageOption)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupTysiacCuiMockWithPlayers() (*interfaces.MockTysiacGame, []*domain.TysiacPlayer) {
	m := setupTysiacCuiMock()
	players := makeTysiacPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestTysiacCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TysiacCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupTysiacCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "トゥシオンツ") // translated helpTitle
		assert.Contains(t, result, "マリッジ")   // play-phase help explains the marriage bonus
		assert.NotEmpty(t, result)
		// The header shows the contract and target during play.
		assert.Contains(t, result, strings.Split(i18n.T("tysiac.headerInfo"), "{{")[0])
		// The live-bid line is only shown while bidding.
		assert.NotContains(t, result, strings.Split(i18n.T("tysiac.headerBid"), "{{")[0])
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "ビッド") // translated bid prompt/help
		// The bid phase adds the current-bid line to the header.
		assert.Contains(t, result, strings.Split(i18n.T("tysiac.headerBid"), "{{")[0])
	})

	t.Run("unconfirmed contract shows placeholder", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
		m.On("GetContract").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "契約: -")
	})

	t.Run("talon phase prompt", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseTalon)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "タロン") // translated talon prompt/help
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTysiacCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TysiacCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TysiacHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupTysiacCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.TysiacHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TysiacHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestTysiacCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TysiacCuiPresenter)
	m := new(interfaces.MockTysiacGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewTysiacPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5687: Web ページは K+Q を揃えたスートをバナーで出しているのに、CUI は
// 「同スートの K と Q を持ち…」という一般論しか出しておらず、いまどのスートで
// 宣言できるかは手札を目で数えるしかなかった。
func TestTysiacCuiPresenter_MarriageOptions(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TysiacCuiPresenter)

	withOptions := func(idx int, opts ...domain.TysiacMarriageOption) *interfaces.MockTysiacGame {
		m, _ := setupTysiacCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(idx)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageOptions")
		m.On("GetMarriageOptions", mock.Anything).Return(opts)
		return m
	}
	spadeAndClub := []domain.TysiacMarriageOption{
		{Suit: domain.CardDesignSpade, Points: 40},
		{Suit: domain.CardDesignClover, Points: 60},
	}

	t.Run("names every declarable suit on the human turn", func(t *testing.T) {
		result := p.Output(withOptions(0, spadeAndClub...), nil)

		assert.Contains(t, result, i18n.Tf("tysiac.promptMarriageReady",
			"suits", "SPADE K-Q (+40), CLOVER K-Q (+60)"))
	})

	t.Run("adds nothing when no suit is paired", func(t *testing.T) {
		result := p.Output(withOptions(0), nil)

		assert.NotContains(t, result, strings.Split(i18n.T("tysiac.promptMarriageReady"), "{{")[0])
	})

	// CPU の手札は伏せたまま。点だけでも「相手が♥のK+Qを持つ」と読めてしまう。
	t.Run("never leaks a CPU hand", func(t *testing.T) {
		result := p.Output(withOptions(1, spadeAndClub...), nil)

		assert.NotContains(t, result, strings.Split(i18n.T("tysiac.promptMarriageReady"), "{{")[0])
	})
}
