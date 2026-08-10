//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// setupPineappleCuiMock returns a MockPineappleGame primed with sensible
// defaults. Tests override individual expectations as needed via
// removeMockCall + m.On(...).
func setupPineappleCuiMock() *interfaces.MockPineappleGame {
	m := new(interfaces.MockPineappleGame)
	cfg := domain.DefaultPineappleConfig()
	m.On("GetConfig").Return(cfg)
	m.On("GetHandCount").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetCommunityCards").Return([]*domain.Card(nil))
	m.On("GetPot").Return(0)
	m.On("GetCpuActions").Return([]domain.HoldemCpuAction(nil))
	m.On("GetRoundResults").Return([]domain.HoldemResult(nil))
	m.On("GetPhase").Return(domain.PineapplePhasePreFlop)
	m.On("GetRebuyPhaseType").Return(domain.PineappleRebuyPhaseNone)
	m.On("GetRebuyCounts").Return([]int{0, 0, 0, 0})
	m.On("GetGameEndFlag").Return(false)
	m.On("IsMuckAvailable").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetInitialDealCount").Return(3).Maybe()
	m.On("IsDiscardAfterFlopBetting").Return(false).Maybe()
	m.On("GetDiscardDone").Return(([]bool)(nil)).Maybe()
	m.On("GetHumanDiscardPairPreviews").Return(([]domain.PineappleDiscardPairPreview)(nil)).Maybe()
	m.On("GetHumanDiscardPreviews").Return(([]domain.PineappleDiscardPreview)(nil)).Maybe()
	return m
}

func setupPineappleCuiMockWithPlayers() (*interfaces.MockPineappleGame, []*domain.PineapplePlayer) {
	m := setupPineappleCuiMock()
	players := []*domain.PineapplePlayer{
		domain.NewPineapplePlayer(true, domain.HoldemStyleTAG),
		domain.NewPineapplePlayer(false, domain.HoldemStyleLAP),
		domain.NewPineapplePlayer(false, domain.HoldemStyleTAP),
		domain.NewPineapplePlayer(false, domain.HoldemStyleGTO),
	}
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPineappleCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PineappleCuiPresenter)

	t.Run("default ja header and player line", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Pineapple Poker")
		assert.Contains(t, result, "テーブル: 4-max")
		assert.Contains(t, result, "ディーラー: Player 0")
		assert.Contains(t, result, "コミュニティ: (なし)")
		assert.Contains(t, result, "ポット: 0")
	})

	t.Run("community cards rendered", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCommunityCards")
		m.On("GetCommunityCards").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "♠5")
		assert.Contains(t, result, "♥9")
	})

	t.Run("folded badge has its own leading space", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		players[1].SetFolded(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, " [フォールド]")
	})

	t.Run("CPU actions rendered with localized line key", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.PineappleActionRaise, Amount: 30},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("showdown result hand uses resultHand key", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundResults")
		m.On("GetPhase").Return(domain.PineapplePhaseEnd)
		m.On("GetRoundResults").Return([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("error message rendered via cuiErrorBlock", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid bet"))
		assert.Contains(t, result, "invalid bet")
	})

	t.Run("game end banner rendered", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("discard prompt says 1 card for a 3-card deal (Pineapple)", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		result := p.Output(m, nil)
		assert.Contains(t, result, "手札から1枚選んで")
	})

	// **Web は残す2枚の性質を候補ごとに出しているのに、CUI はインデックス付きの
	// 手札一覧だけだった (#4685)。**ボードがまだ無い段階で唯一の判断材料になる。
	t.Run("discard phase names what each keep would leave", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		players[0].Reset()
		// ♠1 ♠13 ♥5: [2] を捨てるとスーテッド かつ コネクター (A-K)。
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "スーテッド")
		// **エースは 1 と 14 の両方で数える。**A-K がコネクターと出ること。
		assert.Contains(t, result, "コネクター")
	})

	t.Run("discard phase reports a pair when the keep is paired", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		assert.Contains(t, p.Output(m, nil), "ペア")
	})

	// **Crazy Pineapple は捨てるのがフロップベットの「後」。**Web は事前告知の
	// バナーを出しているのに、CUI は無言だった (#4686)。
	t.Run("flop betting warns that a discard is still coming", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsDiscardAfterFlopBetting")
		m.On("GetPhase").Return(domain.PineapplePhaseFlop)
		m.On("IsDiscardAfterFlopBetting").Return(true)

		assert.Contains(t, p.Output(m, nil), "フロップベット終了後にカードを1枚捨てます")
	})

	t.Run("flop betting says two cards for a four-card deal (Irish)", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsDiscardAfterFlopBetting")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.On("GetPhase").Return(domain.PineapplePhaseFlop)
		m.On("IsDiscardAfterFlopBetting").Return(true)
		m.On("GetInitialDealCount").Return(4)

		assert.Contains(t, p.Output(m, nil), "カードを2枚捨てます")
	})

	// **プレーンな Pineapple は告知してはいけない。**フロップ前にもう捨て終えて
	// いるので、「この後捨てる」は嘘になる。
	t.Run("plain Pineapple gets no upcoming-discard notice", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PineapplePhaseFlop)

		assert.NotContains(t, p.Output(m, nil), "捨てます")
	})

	// **ボードがある局面では完成役そのものを名指しできる (#4686)。**
	t.Run("discard phase names the hand each discard would leave", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanDiscardPreviews")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetHumanDiscardPreviews").Return([]domain.PineappleDiscardPreview{
			{CardIdx: 0, HandRank: domain.PokerHandOnePair},
			{CardIdx: 1, HandRank: domain.PokerHandOnePair},
			{CardIdx: 2, HandRank: domain.PokerHandFlush, Recommended: true},
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0] これを捨てると: ワンペア")
		assert.Contains(t, result, "[2] これを捨てると: フラッシュ")
	})

	t.Run("discard phase marks only the recommended discard", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanDiscardPreviews")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetHumanDiscardPreviews").Return([]domain.PineappleDiscardPreview{
			{CardIdx: 0, HandRank: domain.PokerHandOnePair},
			{CardIdx: 1, HandRank: domain.PokerHandOnePair},
			{CardIdx: 2, HandRank: domain.PokerHandFlush, Recommended: true},
		})

		result := p.Output(m, nil)
		assert.Equal(t, 1, strings.Count(result, "おすすめ"), "推奨は1枚だけのはず")
		for _, line := range strings.Split(result, "\n") {
			if strings.Contains(line, "おすすめ") {
				assert.Contains(t, line, "フラッシュ", "推奨が付くのは最強の役が残る捨て札")
			}
		}
	})

	// **完成役が出せるときに性質だけの助言を重ねない。**役が言えるならそちらが
	// 上位互換で、2種類が並ぶと読み手はどちらに従うのか分からなくなる。
	t.Run("keep features give way to the hand preview", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanDiscardPreviews")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetHumanDiscardPreviews").Return([]domain.PineappleDiscardPreview{
			{CardIdx: 0, HandRank: domain.PokerHandOnePair},
			{CardIdx: 1, HandRank: domain.PokerHandOnePair},
			{CardIdx: 2, HandRank: domain.PokerHandFlush, Recommended: true},
		})
		players[0].Reset()
		// ♠A ♠K ♥5 — #4685 の性質表示ならスーテッド/コネクターが出る手札。
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		assert.NotContains(t, p.Output(m, nil), "スーテッド")
	})

	// **4枚配り (Irish) は2枚まとめて捨てる。**Web は1枚目を選んだ後の3択しか
	// 出せないが、CUI には選択途中が無いので6通りを最初から並べる (#4687)。
	t.Run("four-card deal lists every discard pair", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanDiscardPairPreviews")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetInitialDealCount").Return(4)
		m.On("GetHumanDiscardPairPreviews").Return([]domain.PineappleDiscardPairPreview{
			{DiscardIdx0: 0, DiscardIdx1: 1, HandRank: domain.PokerHandOnePair},
			{DiscardIdx0: 0, DiscardIdx1: 2, HandRank: domain.PokerHandOnePair},
			{DiscardIdx0: 0, DiscardIdx1: 3, HandRank: domain.PokerHandTwoPair, Recommended: true},
			{DiscardIdx0: 1, DiscardIdx1: 2, HandRank: domain.PokerHandHighCard},
			{DiscardIdx0: 1, DiscardIdx1: 3, HandRank: domain.PokerHandOnePair},
			{DiscardIdx0: 2, DiscardIdx1: 3, HandRank: domain.PokerHandOnePair},
		})

		result := p.Output(m, nil)
		// **6通り全部。**一部だけ出すと「載っていない組み合わせは弱い」と
		// 読めてしまう。
		assert.Equal(t, 6, strings.Count(result, "を捨てると:"))
		assert.Contains(t, result, "[0] [3] を捨てると: ツーペア")
		assert.Equal(t, 1, strings.Count(result, "おすすめ"))
	})

	// **4枚配り (Irish) では出さない。**2枚捨てなので「1枚捨てたら残る2枚」
	// という前提が成り立たない。
	t.Run("no keep features for a four-card deal", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetInitialDealCount").Return(4)
		players[0].Reset()
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
		} {
			players[0].AddCard(c)
		}

		assert.NotContains(t, p.Output(m, nil), "を捨てる →")
	})

	t.Run("discard prompt says 2 cards for a 4-card deal (Irish Poker)", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetInitialDealCount").Return(4)
		result := p.Output(m, nil)
		assert.Contains(t, result, "手札から2枚選んで")
	})

	t.Run("hand is indexed during the discard phase before discarding", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "[2]")
	})

	t.Run("hand is not indexed outside the discard phase", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		result := p.Output(m, nil) // default phase is pre-flop
		assert.NotContains(t, result, "[0]")
	})

	t.Run("title is plain Pineapple for a 3-card pre-flop-discard deal", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "パイナップルポーカー")
		assert.NotContains(t, result, "クレイジー")
		assert.NotContains(t, result, "アイリッシュ")
	})

	t.Run("title is Crazy Pineapple when discarding after flop betting", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsDiscardAfterFlopBetting")
		m.On("IsDiscardAfterFlopBetting").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "クレイジーパイナップル")
	})

	t.Run("title is Irish Poker for a 4-card deal", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.On("GetInitialDealCount").Return(4)
		result := p.Output(m, nil)
		assert.Contains(t, result, "アイリッシュポーカー")
	})
}

// TestPineappleCuiPresenter_English verifies the migration's en path.
// Mirrors the suite added to OmahaCuiPresenter_test.go in #1763.
func TestPineappleCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.PineappleCuiPresenter)

	t.Run("output uses English headers", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Pineapple Poker")
		assert.Contains(t, result, "Table: 4-max")
		assert.Contains(t, result, "Dealer: Player 0")
		assert.Contains(t, result, "Community: (none)")
		assert.NotContains(t, result, "テーブル")
	})

	t.Run("folded badge renders English with leading space", func(t *testing.T) {
		m, players := setupPineappleCuiMockWithPlayers()
		players[1].SetFolded(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, " [FOLDED]")
		assert.NotContains(t, result, "[フォールド]")
	})

	t.Run("game end banner uses English", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "Game over")
		assert.NotContains(t, result, "ゲーム終了")
	})
}

func TestPineappleCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PineappleCuiPresenter)

	t.Run("nil entries", func(t *testing.T) {
		m := setupPineappleCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("with entries", func(t *testing.T) {
		m := setupPineappleCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raise 30"},
		})
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "raise")
	})
}
