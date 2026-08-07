package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCaribbeanStudCuiMockDefaults(m *interfaces.MockCaribbeanStudGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestCaribbeanStudCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestCaribbeanStudCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	assert.Contains(t, result, "PLAYER")
	// First dealer card is visible
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "HEART 13")
	// Remaining cards hidden
	assert.Contains(t, result, "??")
}

func TestCaribbeanStudCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(1000).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "合計払戻し: 1000")
}

func TestCaribbeanStudCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestCaribbeanStudCuiPresenter_Output_Error(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestCaribbeanStudCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

// **同じバッチの RedDog / Badugi / DeuceToSeven / SevenCardStud には HintOutput が
// あるのに、カリビアンスタッドだけ CUI に戦略アシストが無かった (#4697)。**
func TestCaribbeanStudCuiPresenter_HintOutput(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	game := func(phase, rank int, hand []*domain.Card) *interfaces.MockCaribbeanStudGame {
		m := new(interfaces.MockCaribbeanStudGame)
		m.On("GetPhase").Return(phase)
		m.On("GetPlayerHand").Return(hand).Maybe()
		m.On("GetPlayerHandRank").Return(rank).Maybe()
		return m
	}
	aceKing := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}
	junk := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}

	t.Run("recommends play with one pair or better", func(t *testing.T) {
		out := p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandOnePair, junk))
		assert.Contains(t, out, "プレイ")
		assert.Contains(t, out, "ワンペア以上")
	})

	// **役が無くても A-K なら降りない。**ここを落とすと期待値上正しい手を
	// フォールドさせる。
	t.Run("recommends play on Ace-King high with no made hand", func(t *testing.T) {
		out := p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandHighCard, aceKing))
		assert.Contains(t, out, "プレイ")
		assert.Contains(t, out, "A と K")
	})

	t.Run("recommends folding a hand below Ace-King", func(t *testing.T) {
		out := p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandHighCard, junk))
		assert.Contains(t, out, "フォールド")
	})

	// **A だけ、K だけでは足りない。**両方揃って初めて A-K ハイ。片方で Play を
	// 勧めると、期待値上フォールドすべき手を打たせる。
	t.Run("folds an ace without a king", func(t *testing.T) {
		aceOnly := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 4, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
		}
		out := p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandHighCard, aceOnly))
		assert.Contains(t, out, "フォールド")
	})

	t.Run("folds a king without an ace", func(t *testing.T) {
		kingOnly := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 4, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
		}
		out := p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandHighCard, kingOnly))
		assert.Contains(t, out, "フォールド")
	})

	t.Run("gives no hint in the bet phase", func(t *testing.T) {
		assert.Contains(t,
			p.HintOutput(game(domain.CaribbeanStudPhaseBet, domain.PokerHandHighCard, nil)),
			"ヒントを出せません")
	})

	t.Run("gives no hint once the round is over", func(t *testing.T) {
		assert.Contains(t,
			p.HintOutput(game(domain.CaribbeanStudPhaseEnd, domain.PokerHandOnePair, junk)),
			"ヒントを出せません")
	})

	t.Run("gives no hint before any card is dealt", func(t *testing.T) {
		assert.Contains(t,
			p.HintOutput(game(domain.CaribbeanStudPhaseAction, domain.PokerHandHighCard, nil)),
			"ヒントを出せません")
	})
}
