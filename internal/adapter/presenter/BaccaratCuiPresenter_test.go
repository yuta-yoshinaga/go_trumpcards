package presenter

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBaccaratCuiMockDefaults(m *interfaces.MockBaccaratGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHandValue").Return(0).Maybe()
	m.On("GetBankerHandValue").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBetType").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()
}

// baccaratStreakMock は履歴だけを差し替えた最小のモックを返す。
func baccaratStreakMock(t *testing.T, history []int) *interfaces.MockBaccaratGame {
	t.Helper()
	m := new(interfaces.MockBaccaratGame)
	m.On("GetPhase").Return(domain.BaccaratPhaseBet).Maybe()
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHandValue").Return(0).Maybe()
	m.On("GetBankerHandValue").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHistory").Return(history).Maybe()
	return m
}

func TestBaccaratCuiPresenter_Output_BetPhase(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestBaccaratCuiPresenter_Output_History(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)

	t.Run("shows P/B/T history symbols when present", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetChips").Return(1000).Maybe()
		m.On("GetPhase").Return(domain.BaccaratPhaseBet).Maybe()
		m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
		m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
		m.On("GetPlayerHandValue").Return(0).Maybe()
		m.On("GetBankerHandValue").Return(0).Maybe()
		m.On("GetGameEndFlag").Return(false).Maybe()
		m.On("GetHistory").Return([]int{
			domain.BaccaratResultPlayer,
			domain.BaccaratResultBanker,
			domain.BaccaratResultTie,
		}).Maybe()
		result := p.Output(m, nil)
		assert.Contains(t, result, "履歴: P B T")
		assert.Contains(t, result, "集計: P:1 B:1 T:1")
	})

	// **Web の ShoeStatsPanel は連勝数も出しているのに CUI には無かった (#4688)。**
	// ロードマップと並んでシューの流れを読む材料になる。
	t.Run("reports the current streak", func(t *testing.T) {
		m := baccaratStreakMock(t, []int{
			domain.BaccaratResultPlayer,
			domain.BaccaratResultBanker,
			domain.BaccaratResultBanker,
		})
		assert.Contains(t, p.Output(m, nil), "バンカー が 2 連勝中")
	})

	// **タイは連勝を切らない。**フロントの computeBaccaratShoeStats と同じ規則。
	// ここがずれると同じ履歴で CUI と Web が違う連勝数を出す。
	t.Run("a tie does not break the streak", func(t *testing.T) {
		m := baccaratStreakMock(t, []int{
			domain.BaccaratResultBanker,
			domain.BaccaratResultTie,
			domain.BaccaratResultBanker,
		})
		assert.Contains(t, p.Output(m, nil), "バンカー が 2 連勝中")
	})

	t.Run("the opposite side ends the streak", func(t *testing.T) {
		m := baccaratStreakMock(t, []int{
			domain.BaccaratResultBanker,
			domain.BaccaratResultBanker,
			domain.BaccaratResultPlayer,
		})
		assert.Contains(t, p.Output(m, nil), "プレイヤー が 1 連勝中")
	})

	// タイだけの履歴には連勝が無い。行ごと出さない。
	t.Run("ties only means no streak line", func(t *testing.T) {
		m := baccaratStreakMock(t, []int{domain.BaccaratResultTie, domain.BaccaratResultTie})
		assert.NotContains(t, p.Output(m, nil), "連勝中")
	})

	t.Run("omits the history line when empty", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		setupBaccaratCuiMockDefaults(m)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "履歴:")
	})

	t.Run("renders unknown result values as ? and caps the trailing run", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetChips").Return(1000).Maybe()
		m.On("GetPhase").Return(domain.BaccaratPhaseBet).Maybe()
		m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
		m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
		m.On("GetPlayerHandValue").Return(0).Maybe()
		m.On("GetBankerHandValue").Return(0).Maybe()
		m.On("GetGameEndFlag").Return(false).Maybe()
		// 40 entries (> cap) ending in an unexpected value to hit both branches.
		history := make([]int, 0, 40)
		for range 39 {
			history = append(history, domain.BaccaratResultPlayer)
		}
		history = append(history, 99)
		m.On("GetHistory").Return(history).Maybe()
		result := p.Output(m, nil)
		// Unknown value maps to '?'.
		assert.Contains(t, result, "?")
		// Only the last baccaratHistoryMaxShown symbols are shown (29 P + 1 ?);
		// without the cap there would be 39 P. Count within the symbols line only
		// (the separate totals line also contains P/B/T letters).
		var histLine string
		for _, ln := range strings.Split(result, "\n") {
			if strings.HasPrefix(ln, "履歴:") {
				histLine = ln
				break
			}
		}
		assert.Equal(t, baccaratHistoryMaxShown, strings.Count(histLine, "P")+strings.Count(histLine, "?"))
	})
}

func TestBaccaratCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(2).Maybe()
	m.On("GetBankerHandValue").Return(7).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "BANKER")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "払戻し: 200")
	assert.Contains(t, result, "SPADE 9")
	// No side bet placed -> no side-bet outcome lines.
	assert.NotContains(t, result, "的中")
	assert.NotContains(t, result, "外れ")
}

func TestBaccaratCuiPresenter_Output_EndPhase_SideBets(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 4, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(8).Maybe()
	m.On("GetBankerHandValue").Return(7).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return([]*domain.BacSideBetResult{
		{BetType: domain.BacSideBetPlayerPair, BetAmount: 20, Payout: 240}, // hit
		{BetType: domain.BacSideBetBankerPair, BetAmount: 20, Payout: 0},   // miss
	}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player Pair: 的中 +240")
	assert.Contains(t, result, "Banker Pair: 外れ")
}

func TestBaccaratCuiPresenter_Output_EndPhase_BankerWins(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 8, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(3).Maybe()
	m.On("GetBankerHandValue").Return(8).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetBanker).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "バンカーの勝ち")
	assert.Contains(t, result, "BANKER")
}

func TestBaccaratCuiPresenter_Output_EndPhase_Tie(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1900).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(5).Maybe()
	m.On("GetBankerHandValue").Return(5).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetTie).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetPayout").Return(900).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "タイ")
	assert.Contains(t, result, "TIE")
}

func TestBaccaratCuiPresenter_Output_Error(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrInvalidAmount, "Invalid bet amount."))
	assert.Contains(t, result, "Invalid bet amount.")
}

func TestBaccaratCuiPresenter_Output_UnknownPhase(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(99).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestBaccaratCuiPresenter_Output_EndPhase_UnknownResult(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHandValue").Return(0).Maybe()
	m.On("GetBankerHandValue").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetResult").Return(domain.GameResult(99)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "払戻し: 0")
}

func TestBaccaratCuiPresenter_Output_UnknownBetType(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(5).Maybe()
	m.On("GetBankerHandValue").Return(5).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(99).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestBaccaratCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(BaccaratCuiPresenter)

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game ended with log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 100 on player"},
		})
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "bet 100 on player")
	})

	t.Run("game ended without log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// #5497: Web はベット前に payoutRef パネルで配当倍率を出しているのに、CUI には
// 配当を出すコマンドも表示も無く、プレイヤーは倍率を知らないままベットタイプを
// 選ぶことになる。BlackJack は #4677 で同じ対応が入っている。
func TestBaccaratCuiPresenter_PayoutTable(t *testing.T) {
	bp := new(BaccaratCuiPresenter)

	t.Run("lists every payout during the betting phase", func(t *testing.T) {
		b := domain.NewDefaultBaccarat()
		out := bp.Output(b, nil)
		assert.Contains(t, out, i18n.T("baccarat.payoutRefTitle"))
		// 倍率はドメインの定数から作る。文言に直接書くと定数と乖離する。
		assert.Contains(t, out, i18n.Tf("baccarat.payoutRefTie",
			"rate", strconv.Itoa(domain.BaccaratTiePayoutRate)))
		assert.Contains(t, out, i18n.Tf("baccarat.payoutRefPair",
			"rate", strconv.Itoa(domain.BacPairPayoutRate)))
		assert.Contains(t, out, i18n.Tf("baccarat.payoutRefBanker",
			"commission", strconv.Itoa(domain.BaccaratCommissionRate)))
		assert.Contains(t, out, i18n.T("baccarat.payoutRefPlayer"))
	})

	// **結果表示中は重ねない。** 決着した卓に配当表を並べると、いま起きたことが
	// 読み取りにくくなる。次のベット前にまた出る。
	t.Run("stays quiet while a finished round is on screen", func(t *testing.T) {
		b := domain.NewDefaultBaccarat()
		b.SetGameEndFlag(true)
		assert.NotContains(t, bp.Output(b, nil), i18n.T("baccarat.payoutRefTitle"))
	})
}
