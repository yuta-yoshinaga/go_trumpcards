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

func makeFortyFivesPlayers() []*domain.FortyFivesPlayer {
	return []*domain.FortyFivesPlayer{
		domain.NewFortyFivesPlayer(true),
		domain.NewFortyFivesPlayer(false),
		domain.NewFortyFivesPlayer(false),
		domain.NewFortyFivesPlayer(false),
	}
}

func setupFortyFivesCuiMock() *interfaces.MockFortyFivesGame {
	m := new(interfaces.MockFortyFivesGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.FortyFivesPhasePlay)
	m.On("GetContractProgress").Return((*domain.FortyFivesContractProgress)(nil)).Maybe()
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.FortyFivesBidTwenty)
	m.On("GetBids").Return([domain.FortyFivesPlayerCnt]domain.FortyFivesBid{domain.FortyFivesBidTwenty, domain.FortyFivesBidPass, domain.FortyFivesBidPass, domain.FortyFivesBidPass})
	m.On("GetBidDone").Return([domain.FortyFivesPlayerCnt]bool{true, true, false, false})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.FortyFivesTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupFortyFivesCuiMockWithPlayers() (*interfaces.MockFortyFivesGame, []*domain.FortyFivesPlayer) {
	m := setupFortyFivesCuiMock()
	players := makeFortyFivesPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetTopTrumpIndices", mock.Anything).Return(([]int)(nil)).Maybe()
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestFortyFivesCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.FortyFivesCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupFortyFivesCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Auction Forty-Fives")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetPhase").Return(domain.FortyFivesPhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// The bid-history line lists every player: seat 0 bid, seat 2 has not (shown "-").
		assert.Contains(t, result, strings.Split(i18n.T("fortyfives.bidHistory"), "{{")[0])
		assert.Contains(t, result, "=-") // an un-bid player renders as "-"
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.FortyFivesPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.FortyFivesPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestFortyFivesCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.FortyFivesCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.FortyFivesHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupFortyFivesCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.FortyFivesHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.FortyFivesHint{CardIndices: nil, Reason: "take_trick"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestFortyFivesCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.FortyFivesCuiPresenter)
	m := new(interfaces.MockFortyFivesGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewFortyFivesPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// **契約の進捗がラウンド終了まで一切出ていなかった (#4724)。**落札チームが
// あと何点必要かは、押すか降りるかの判断そのもの。
func TestFortyFivesCuiPresenter_ContractProgress(t *testing.T) {
	p := new(presenter.FortyFivesCuiPresenter)
	withProgress := func(pr *domain.FortyFivesContractProgress) *interfaces.MockFortyFivesGame {
		m, _ := setupFortyFivesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractProgress")
		m.On("GetContractProgress").Return(pr)
		return m
	}

	t.Run("shows the points the declaring team still needs", func(t *testing.T) {
		out := p.Output(withProgress(&domain.FortyFivesContractProgress{
			DeclarerTeam: 0, Points: 5, Contract: 20, Remaining: 15,
			Status: domain.FortyFivesContractNeedMore,
		}), nil)
		assert.Contains(t, out, "5/20")
		assert.Contains(t, out, "あと15点")
	})

	t.Run("says when the contract is already made", func(t *testing.T) {
		out := p.Output(withProgress(&domain.FortyFivesContractProgress{
			DeclarerTeam: 1, Points: 25, Contract: 20,
			Status: domain.FortyFivesContractMade,
		}), nil)
		assert.Contains(t, out, "成立")
		assert.Contains(t, out, "チームB")
	})

	// **「もう届かない」と「あと何点」は別の話。**同じ文言だと、投げるべき
	// ラウンドと押すべきラウンドの区別が付かない。
	t.Run("says when the contract can no longer be made", func(t *testing.T) {
		out := p.Output(withProgress(&domain.FortyFivesContractProgress{
			DeclarerTeam: 0, Points: 5, Contract: 25, Remaining: 20,
			Status: domain.FortyFivesContractFailed,
		}), nil)
		assert.Contains(t, out, "不成立")
		assert.NotContains(t, out, "あと20点")
	})

	t.Run("shows nothing before the bid is settled", func(t *testing.T) {
		assert.NotContains(t, p.Output(withProgress(nil), nil), "契約:")
	})
}

// #5643: 上位切り札 (切り札の5・切り札のJ・♥A) は固定で最強で、持っていると
// マストフォローが免除される。**♥A は切り札スート外でも切り札扱い**なので、
// スート記号を眺めても分からない。Web もこれまで何も出していなかった。
func TestFortyFivesCuiPresenter_MarksTheTopTrumps(t *testing.T) {
	p := new(presenter.FortyFivesCuiPresenter)

	handMock := func(tops []int) *interfaces.MockFortyFivesGame {
		m, players := setupFortyFivesCuiMockWithPlayers()
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTopTrumpIndices")
		m.On("GetTopTrumpIndices", mock.Anything).Return(tops)
		return m
	}

	t.Run("marks only the top trumps", func(t *testing.T) {
		out := p.Output(handMock([]int{0, 1}), nil)

		assert.Contains(t, out, "[0]SPADE 5"+presenter.CuiTopTrumpMark)
		assert.Contains(t, out, "[1]"+color.Red("HEART 1")+presenter.CuiTopTrumpMark)
		assert.NotContains(t, out, "[2]CLOVER 9"+presenter.CuiTopTrumpMark)
	})

	t.Run("explains the mark, including the reneging exemption", func(t *testing.T) {
		out := p.Output(handMock([]int{0, 1}), nil)

		assert.Contains(t, out, i18n.T("fortyfives.topTrumpLegend"))
	})

	t.Run("no top trump in hand leaves the list unmarked", func(t *testing.T) {
		out := p.Output(handMock(nil), nil)

		assert.NotContains(t, out, presenter.CuiTopTrumpMark)
		assert.NotContains(t, out, i18n.T("fortyfives.topTrumpLegend"))
	})
}
