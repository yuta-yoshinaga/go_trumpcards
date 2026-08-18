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

func makeTwentyNinePlayers() []*domain.TwentyNinePlayer {
	return []*domain.TwentyNinePlayer{
		domain.NewTwentyNinePlayer(true),
		domain.NewTwentyNinePlayer(false),
		domain.NewTwentyNinePlayer(false),
		domain.NewTwentyNinePlayer(false),
	}
}

func setupTwentyNineCuiMock() *interfaces.MockTwentyNineGame {
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpRevealed").Return(true)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.TwentyNineBidTwenty)
	m.On("GetBids").Return([domain.TwentyNinePlayerCnt]domain.TwentyNineBid{domain.TwentyNineBidTwenty, domain.TwentyNineBidPass, domain.TwentyNineBidPass, domain.TwentyNineBidPass})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetRoundTeamPoints").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 既定は「制限なし」。目印を確かめるテストは自分で上書きする。
	m.On("GetPlayableIndices", 0).Return([]int(nil)).Maybe()
	m.On("GetContractProgress").Return((*domain.TwentyNineContractProgress)(nil)).Maybe()
	return m
}

func setupTwentyNineCuiMockWithPlayers() (*interfaces.MockTwentyNineGame, []*domain.TwentyNinePlayer) {
	m := setupTwentyNineCuiMock()
	players := makeTwentyNinePlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

// **Web は playableIndices をリング表示しているのに、CUI は素の一覧だけで、
// 番号を入力してエラーを踏むまで合法手が分からなかった (#4725)。**
func TestTwentyNineCuiPresenter_MarksPlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TwentyNineCuiPresenter)

	addThreeCards := func(pl *domain.TwentyNinePlayer) {
		pl.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		pl.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		pl.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	}

	t.Run("human play turn marks only the legal cards", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		addThreeCards(players[0])
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayableIndices")
		m.On("GetPlayableIndices", 0).Return([]int{1})

		result := p.Output(m, nil)
		assert.Contains(t, result, "[1]HEART 5*", "合法手には目印が付く")
		assert.NotContains(t, result, "[0]SPADE 1*", "非合法手には付かない")
		assert.NotContains(t, result, "[2]CLOVER 9*")
	})

	// **目印を出さない側も踏む。**ビッド中は制限そのものが決まっていないので、
	// 常時マークに退化すると「全部出せる」と誤って伝えてしまう。
	t.Run("bid phase leaves the hand unmarked", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		addThreeCards(players[0])
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseBid)
		// **ドメインは合法手を返し得る。**ここを nil にすると、フェーズのガードを
		// 外しても目印が出ず、この一本が何も確かめなくなる。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayableIndices")
		m.On("GetPlayableIndices", 0).Return([]int{1})

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]SPADE 1")
		assert.NotContains(t, result, "HEART 5*", "ビッド中は目印を出さない")
	})

	t.Run("cpu turn leaves the human hand unmarked", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		addThreeCards(players[0])
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)
		// 手番ガードを外したときに確実に落ちるよう、合法手を返させておく。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayableIndices")
		m.On("GetPlayableIndices", 0).Return([]int{1})

		result := p.Output(m, nil)
		assert.NotContains(t, result, "HEART 5*", "相手の手番では目印を出さない")
	})
}

func TestTwentyNineCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TwentyNineCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Twenty-Nine")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt hides trump", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpRevealed")
		m.On("GetPhase").Return(domain.TwentyNinePhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		m.On("GetTrumpSuit").Return(0)
		m.On("GetTrumpRevealed").Return(false)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTwentyNineCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TwentyNineCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TwentyNineHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestTwentyNineCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TwentyNineCuiPresenter)
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5644: 落札チームが契約に届くかどうかは、目標点・現在点・場に残る点の 3 つを
// 突き合わせないと分からない。姉妹ゲームの FortyFives は #4724 でこれを出したの
// に、29 は計算そのものが無かった。
func TestTwentyNineCuiPresenter_ShowsTheContractProgress(t *testing.T) {
	p := new(presenter.TwentyNineCuiPresenter)

	progressMock := func(pr *domain.TwentyNineContractProgress) *interfaces.MockTwentyNineGame {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractProgress")
		m.On("GetContractProgress").Return(pr)
		return m
	}

	t.Run("still reachable shows how many points are missing", func(t *testing.T) {
		out := p.Output(progressMock(&domain.TwentyNineContractProgress{
			DeclarerTeam: 0, Points: 10, Contract: 20, Remaining: 10,
			Status: domain.TwentyNineContractNeedMore,
		}), nil)

		assert.Contains(t, out, i18n.Tf("twentynine.contractProgress",
			"team", i18n.T("twentynine.teamA"),
			"got", "10", "contract", "20",
			"status", i18n.Tf("twentynine.contractNeedMore", "remaining", "10")))
	})

	t.Run("already made", func(t *testing.T) {
		out := p.Output(progressMock(&domain.TwentyNineContractProgress{
			DeclarerTeam: 1, Points: 20, Contract: 16, Remaining: 0,
			Status: domain.TwentyNineContractMade,
		}), nil)

		assert.Contains(t, out, i18n.T("twentynine.contractMade"))
		assert.Contains(t, out, i18n.T("twentynine.teamB"))
	})

	t.Run("no longer reachable", func(t *testing.T) {
		out := p.Output(progressMock(&domain.TwentyNineContractProgress{
			DeclarerTeam: 0, Points: 10, Contract: 20, Remaining: 10,
			Status: domain.TwentyNineContractFailed,
		}), nil)

		assert.Contains(t, out, i18n.T("twentynine.contractFailed"))
	})

	t.Run("nothing to say before a contract exists", func(t *testing.T) {
		out := p.Output(progressMock(nil), nil)

		assert.NotContains(t, out, i18n.T("twentynine.contractMade"))
		assert.NotContains(t, out, i18n.T("twentynine.contractFailed"))
		assert.NotContains(t, out, i18n.Tf("twentynine.contractNeedMore", "remaining", "10"))
	})
}
