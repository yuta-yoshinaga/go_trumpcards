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
)

func makeGermanSoloPlayers() []*domain.GermanSoloPlayer {
	return []*domain.GermanSoloPlayer{
		domain.NewGermanSoloPlayer(true),
		domain.NewGermanSoloPlayer(false),
		domain.NewGermanSoloPlayer(false),
		domain.NewGermanSoloPlayer(false),
	}
}

func setupGermanSoloCuiMock() *interfaces.MockGermanSoloGame {
	m := new(interfaces.MockGermanSoloGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetWinningBid").Return(domain.GermanSoloBidFrage)
	m.On("GetHighestBid").Return(domain.GermanSoloBidNone)
	m.On("IsPlayingAlone").Return(false)
	m.On("GetCalledAceSuit").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCallableAceSuits").Return([]int(nil))
	m.On("IsHumanAceCallTurn").Return(false)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetOutcome").Return(domain.GermanSoloOutcomeMade)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.GermanSoloPlayerCnt]int{})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetSideTrickCounts").Return(2, 3)
	m.On("RequiredTricks").Return(domain.GermanSoloMakeTricks)
	m.On("GetBiddableBids").Return([]int{int(domain.GermanSoloBidFrage), int(domain.GermanSoloBidSolo), int(domain.GermanSoloBidTout)})
	return m
}

func setupGermanSoloCuiMockWithPlayers() (*interfaces.MockGermanSoloGame, []*domain.GermanSoloPlayer) {
	m := setupGermanSoloCuiMock()
	players := makeGermanSoloPlayers()
	m.On("GetPlayerCnt").Return(domain.GermanSoloPlayerCnt)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGermanSoloCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GermanSoloCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupGermanSoloCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "ジャーマン・ソロ") // translated helpTitle / role
		assert.Contains(t, result, "マストフォロー")  // play-phase help mentions must-follow
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GermanSoloPhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "ビッド") // translated bid prompt/help
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GermanSoloPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GermanSoloPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "成功") // outcome label
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

// **マタドールは常にトリックに勝つ。**Web はバッジで示すのに、CUI は素の一覧
// しか出しておらず、序列を覚えていないと気づけなかった (#4919)。
func TestGermanSoloCuiPresenter_AnnotatesMatadors(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GermanSoloCuiPresenter)

	t.Run("marks all three matadors once trump is decided", func(t *testing.T) {
		m, players := setupGermanSoloCuiMockWithPlayers()
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 12, false)) // スパディーユ (♣Q)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))   // マニーユ (♥ が切り札)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))  // バスタ (♠Q)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))  // ただの切り札

		out := p.Output(m, nil)
		assert.Contains(t, out, "[0]CLOVER 12(スパディーユ)")
		assert.Contains(t, out, "[1]HEART 7(マニーユ)")
		assert.Contains(t, out, "[2]SPADE 12(バスタ)")
		// 切り札の K は平の切り札。注記は付かない。
		assert.Contains(t, out, "[3]HEART 13")
		assert.NotContains(t, out, "[3]HEART 13(")
	})

	// **マニーユは切り札スート次第。**♠ が切り札なら ♥7 はただの平札。
	t.Run("the manille follows the trump suit", func(t *testing.T) {
		m, players := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

		out := p.Output(m, nil)
		assert.NotContains(t, out, "[0]HEART 7(")
		assert.Contains(t, out, "[1]SPADE 7(マニーユ)")
	})

	// 切り札未確定なら注記なし (受け入れ条件2)。
	t.Run("no annotation before trump is decided", func(t *testing.T) {
		m, players := setupGermanSoloCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(-1)
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))

		out := p.Output(m, nil)
		// 手札行に注記が付かないことを見る。序列の説明は promptPlayHelp が
		// 常に出しているので、文言そのものの有無では判定できない。
		assert.Contains(t, out, "[0]SPADE 1  [1]CLOVER 1")
		assert.NotContains(t, out, "[0]SPADE 1(")
		assert.NotContains(t, out, "[1]CLOVER 1(")
	})
}

func TestGermanSoloCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GermanSoloCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.GermanSoloHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupGermanSoloCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("bid decision hint shows the action, not an empty card list", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: nil, Reason: "bid_solo"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ソロ")  // recommended action name
		assert.Contains(t, result, "を推奨") // hintDecision format
		assert.NotContains(t, result, "HINT: -")
	})

	// **エース呼びのヒントはスートを名指しする。** 札を指さないので、スートを
	// 落とすと「呼べ」としか言っていない行になる。
	t.Run("ace-call hint names the suit", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.GermanSoloHint{Reason: "call_ace", SuitHint: domain.CardDesignClover})
		result := p.HintOutput(m)
		assert.Contains(t, result, "クラブ")
		assert.Contains(t, result, "味方の助けが大きい")
		assert.NotContains(t, result, "HINT: -")
	})

	t.Run("non-bid empty-card hint falls back to the card line", func(t *testing.T) {
		m, _ := setupGermanSoloCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.GermanSoloHint{CardIndices: nil, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "-") // hintCard fallback keeps the placeholder
		assert.NotContains(t, result, "を推奨")
	})
}

func TestGermanSoloCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GermanSoloCuiPresenter)
	m := new(interfaces.MockGermanSoloGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewGermanSoloPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
