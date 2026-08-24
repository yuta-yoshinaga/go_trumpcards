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
)

func makeGleekPlayers() []*domain.GleekPlayer {
	return []*domain.GleekPlayer{
		domain.NewGleekPlayer(true),
		domain.NewGleekPlayer(false),
		domain.NewGleekPlayer(false),
	}
}

func setupGleekCuiMock() *interfaces.MockGleekGame {
	m := new(interfaces.MockGleekGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTurnUp").Return(domain.NewCard(domain.CardDesignHeart, 4, false))
	m.On("GetBuyerIdx").Return(0)
	m.On("GetWinningBid").Return(14)
	m.On("HighestBid").Return(14)
	m.On("NextBidAmount").Return(16)
	m.On("GetCurrentBidderIdx").Return(0)
	m.On("GetRuffWinnerIdx").Return(1)
	m.On("GetRuffs").Return([]*domain.GleekRuff{
		{PlayerIdx: 0, Suit: domain.CardDesignHeart, Total: 20},
		{PlayerIdx: 1, Suit: domain.CardDesignSpade, Total: 31},
		{PlayerIdx: 2, Suit: domain.CardDesignClover, Total: 18},
	})
	m.On("GetMelds").Return([]*domain.GleekMeld{})
	m.On("GetTrickPoints").Return([domain.GleekPlayerCnt]int{})
	m.On("GetPlayerScores").Return([domain.GleekPlayerCnt]int{})
	m.On("DealPoints").Return(78)
	m.On("Par").Return(26)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GleekPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return((*domain.GleekHint)(nil)).Maybe()
	return m
}

func setupGleekCuiMockWithPlayers() (*interfaces.MockGleekGame, []*domain.GleekPlayer) {
	m := setupGleekCuiMock()
	players := makeGleekPlayers()
	m.On("GetPlayerCnt").Return(domain.GleekPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestGleekCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GleekCuiPresenter)

	t.Run("play phase shows the hand and the honour values", func(t *testing.T) {
		m, players := setupGleekCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false)) // Tib = 15
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "グリーク")
		// **名札の点は手札に出す。** 出さないと 15 点の Tib を捨てた理由が読めない。
		assert.Contains(t, result, "HEART 1(15)")
		assert.NotContains(t, result, "SPADE 13(", "平札に点は付かない")
	})

	t.Run("stock line names the buyer once the auction closes", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "14 で落札")
	})

	t.Run("stock line says the auction is open before it closes", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBuyerIdx")
		m.On("GetBuyerIdx").Return(-1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "競り中")
	})

	// **段階の点は盤に出さないと見えない。** ラフとメルドで動いた点が無いと、
	// 累積点だけが理由なく動いているように見える。
	t.Run("ruff and meld lines report the stage payments", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMelds")
		m.On("GetMelds").Return([]*domain.GleekMeld{
			{PlayerIdx: 0, Rank: 13, Count: 3, Value: 3},
			{PlayerIdx: 1, Rank: 11, Count: 4, Value: 2},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラフ")
		assert.Contains(t, result, "スペード")
		assert.Contains(t, result, "キング3枚")
		assert.Contains(t, result, "ジャック4枚")
		assert.Contains(t, result, "マーニヴァル")
	})

	t.Run("bid phase names the amount that may still be bid", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GleekPhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "競り中")
		assert.Contains(t, result, "16")
	})

	// **上限に達したら競り上げを勧めない。** サーバが弾く選択肢を案内すると、
	// その通り打った人間だけが弾かれる。
	t.Run("bid phase says so once the ceiling is reached", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GleekPhaseBid)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "NextBidAmount")
		m.On("NextBidAmount").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "これ以上")
	})

	t.Run("discard phase asks for the right number of cards", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GleekPhaseDiscard)
		result := p.Output(m, nil)
		assert.Contains(t, result, "7枚")
		assert.Contains(t, result, "discard")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GleekPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "次のトリックへ")
	})

	// **基準点はそのディールから数える。** 上限を出すと、名札が場外に落ちた
	// ディールで説明が合わなくなる。
	t.Run("round end reports this deal's total and its par", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GleekPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "78")
		assert.Contains(t, result, "26")
	})

	t.Run("game end banner names the winner and drops the phase prompt", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "あなた")
		assert.NotContains(t, result, "マストフォロー")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("boom")), "boom")
	})
}

func TestGleekCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GleekCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.GleekHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	// **競りのヒントは札でなく額を指す。** 索引を出しても意味が無い。
	t.Run("bid hint names the amount", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GleekHint{Bid: 16, Reason: "bid_raise"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "16")
		assert.Contains(t, result, "を推奨")
		assert.NotContains(t, result, "HINT: -")
	})

	t.Run("dropping out is named as an action, not a card", func(t *testing.T) {
		m, _ := setupGleekCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GleekHint{Bid: 0, Reason: "bid_pass"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "降りる")
		assert.NotContains(t, result, "HINT: -")
	})

	t.Run("play hint lists the recommended card", func(t *testing.T) {
		m, players := setupGleekCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GleekHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "SPADE 13")
	})

	// **捨て札のヒントは落札者の手札を指す。** 現在の手番の席を読むと、
	// 別人の手札の索引を出すことになる。
	t.Run("discard hint reads the buyer's hand", func(t *testing.T) {
		m, players := setupGleekCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GleekHint{CardIndices: []int{0}, Reason: "discard_stock"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "CLOVER 4")
	})
}
