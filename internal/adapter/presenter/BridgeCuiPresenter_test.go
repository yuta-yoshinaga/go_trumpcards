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

// setupBridgeCuiMock creates a MockBridgeGame with sensible defaults for CUI tests.
func setupBridgeCuiMock() *interfaces.MockBridgeGame {
	m := new(interfaces.MockBridgeGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BridgePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1) // Spade
	m.On("GetContractLevel").Return(1)
	m.On("GetContractSuit").Return(3)
	m.On("GetDoubled").Return(0)
	m.On("BridgeMinLegalBid").Return(1, 1, true).Maybe()
	m.On("BridgeCanDouble", mock.Anything).Return(false).Maybe()
	m.On("BridgeCanRedouble", mock.Anything).Return(false).Maybe()
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetDummyIdx").Return(2)
	m.On("GetBidHistory").Return([]*domain.BridgeBidEntry(nil))
	m.On("GetVulnerability", 0).Return(false)
	m.On("GetVulnerability", 1).Return(false)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetGamesWon", 0).Return(0)
	m.On("GetGamesWon", 1).Return(0)
	m.On("GetBelowLine", 0).Return(0)
	m.On("GetBelowLine", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("IsOpeningLeadDone").Return(false)
	m.On("GetDummyHand").Return(([]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultBridgeConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupBridgeCuiMockWithPlayers() (*interfaces.MockBridgeGame, []*domain.BridgePlayer) {
	m := setupBridgeCuiMock()
	players := makeBridgePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBridgeCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BridgeCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupBridgeCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Contract Bridge (コントラクトブリッジ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "あなた: チーム0 獲得0トリック 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: チーム1 獲得0トリック 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "p <i> (play)")
	})

	t.Run("trump suit shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: SPADE")
	})

	t.Run("trump suit notrump", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(-1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: ノートランプ")
	})

	t.Run("trump suit undecided", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("contract info shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		// Contract suit 3 is the Heart bid suit, localized rather than printed raw.
		assert.Contains(t, result, "コントラクト: 1レベル")
		assert.Contains(t, result, "HEART")
		assert.NotContains(t, result, "スート3")
	})

	t.Run("contract no trump shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractSuit")
		m.On("GetContractSuit").Return(domain.BridgeBidSuitNT)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ノートランプ")
	})

	t.Run("contract suit localized for each bid suit", func(t *testing.T) {
		cases := map[int]string{
			domain.BridgeBidSuitClub:    "CLOVER",
			domain.BridgeBidSuitDiamond: "DIAMOND",
			domain.BridgeBidSuitSpade:   "SPADE",
			99:                          "UNKNOWN", // out-of-range falls through to UNKNOWN
		}
		for suit, name := range cases {
			m, _ := setupBridgeCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractSuit")
			m.On("GetContractSuit").Return(suit)

			result := p.Output(m, nil)
			assert.Contains(t, result, name)
		}
	})

	t.Run("contract doubled", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDoubled")
		m.On("GetDoubled").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ダブル")
	})

	t.Run("contract redoubled", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDoubled")
		m.On("GetDoubled").Return(2)

		result := p.Output(m, nil)
		assert.Contains(t, result, "リダブル")
	})

	t.Run("declarer and dummy shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "デクレアラー: あなた")
		assert.Contains(t, result, "ダミー: CPU 2")
	})

	t.Run("no contract hides contract section", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractLevel")
		m.On("GetContractLevel").Return(0)

		result := p.Output(m, nil)
		assert.NotContains(t, result, "コントラクト:")
	})

	t.Run("vulnerability shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetVulnerability")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetVulnerability")
		m.On("GetVulnerability", 0).Return(true)
		m.On("GetVulnerability", 1).Return(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "バルネラビリティ: チーム0=true チーム1=false")
	})

	t.Run("team scores shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)

		result := p.Output(m, nil)
		assert.Contains(t, result, "チーム0: 5点")
		assert.Contains(t, result, "チーム1: 3点")
	})

	t.Run("dealer shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディーラー: あなた")
	})

	t.Run("dummy hand shown after opening lead", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsOpeningLeadDone")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDummyHand")
		m.On("IsOpeningLeadDone").Return(true)
		m.On("GetDummyHand").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 10, false),
			domain.NewCard(domain.CardDesignSpade, 1, false),
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ダミー手札: HEART 10, SPADE 1")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック: あなた=CLOVER 3, CPU 1=CLOVER 7")
	})

	t.Run("no trick cards hides trick section", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "トリック: あなた")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner team", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "チーム0の勝利です！")
	})

	t.Run("bid phase shows bidder and command", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "b <type>")
		// No auction history before the first bid.
		assert.NotContains(t, result, strings.Split(i18n.T("bridge.bidHistory"), "{{")[0])
	})

	t.Run("bid phase shows the auction history", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBidHistory")
		m.On("GetPhase").Return(domain.BridgePhaseBid)
		m.On("GetBidHistory").Return([]*domain.BridgeBidEntry{
			{PlayerIdx: 0, BidType: domain.BridgeBidNormal, Level: 1, Suit: domain.BridgeBidSuitClub},
			{PlayerIdx: 1, BidType: domain.BridgeBidPass},
			{PlayerIdx: 2, BidType: domain.BridgeBidDouble},
		})
		result := p.Output(m, nil)
		// Auction header, the 1♣ normal bid (level + clover suit), and the labels.
		assert.Contains(t, result, strings.Split(i18n.T("bridge.bidHistory"), "{{")[0])
		assert.Contains(t, result, "1CLOVER")
		assert.Contains(t, result, i18n.T("bridge.bidPass"))
		assert.Contains(t, result, i18n.T("bridge.bidDouble"))
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "n (next trick)")
	})

	t.Run("round end phase shows next round command", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr (next round)")
	})

	t.Run("human with no cards does not print extra cards line", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: チーム0 獲得0トリック 0枚")
	})

	t.Run("player with tricks", func(t *testing.T) {
		m, players := setupBridgeCuiMockWithPlayers()
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: チーム1 獲得1トリック 0枚")
	})
}

func TestBridgeCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BridgeCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewBridgePlayer(true, 0)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた", "棋譜の座席名が他の行と揃っていない")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestBridgeCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return((*domain.BridgeHint)(nil))

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("bid pass hint", func(t *testing.T) {
		bidType := int(domain.BridgeBidPass)
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			BidType: &bidType,
			Reason:  "weak_hand",
		})

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "パス")
		assert.Contains(t, result, "弱い手札")
	})

	t.Run("bid normal hint", func(t *testing.T) {
		bidType := int(domain.BridgeBidNormal)
		bidLevel := 2
		bidSuit := 3
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			BidType:  &bidType,
			BidLevel: &bidLevel,
			BidSuit:  &bidSuit,
			Reason:   "strong_hand",
		})

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "ビッド")
		assert.Contains(t, result, "2レベル")
		assert.Contains(t, result, "HEART")
		assert.NotContains(t, result, "スート3")
		assert.Contains(t, result, "強い手札")
	})

	t.Run("bid double hint", func(t *testing.T) {
		bidType := int(domain.BridgeBidDouble)
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			BidType: &bidType,
			Reason:  "competitive_bid",
		})

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ダブル")
		assert.Contains(t, result, "競り合い")
	})

	t.Run("bid redouble hint", func(t *testing.T) {
		bidType := int(domain.BridgeBidRedouble)
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			BidType: &bidType,
			Reason:  "strong_hand",
		})

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "リダブル")
	})

	t.Run("play hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})
		player := domain.NewBridgePlayer(true, 0)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("hint with nil fields returns no hint", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{
			Reason: "unknown",
		})

		p := new(presenter.BridgeCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"trump_cut":       "切り札でカット",
			"lead_strong":     "強いカードでリード",
			"lead_low":        "低いカードでリード",
			"support_partner": "パートナーをサポート",
			"competitive_bid": "競り合い",
			"unknown_reason":  "unknown_reason",
		}
		for key, expected := range reasons {
			idx := 0
			m := new(interfaces.MockBridgeGame)
			m.On("GetHint").Return(&domain.BridgeHint{
				CardIndex: &idx,
				Reason:    key,
			})
			player := domain.NewBridgePlayer(true, 0)
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)

			p := new(presenter.BridgeCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}

// TestBridgeCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in BridgeCuiPresenter now follows
// the active locale. The default ja path is exercised by the assertions
// above; this suite re-runs Output / HintOutput under LANG=en and checks
// the English keys win out.
func TestBridgeCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.BridgeCuiPresenter)

	t.Run("output uses English headers and prompts", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Contract Bridge")
		assert.Contains(t, result, "Round: 1")
		assert.Contains(t, result, "Trick: 1")
		assert.Contains(t, result, "Dealer: You")
		assert.Contains(t, result, "Trump: SPADE")
		assert.Contains(t, result, "Contract: 1-level HEART")
		assert.Contains(t, result, "Vulnerability:")
		assert.Contains(t, result, "Team 0:")
		assert.NotContains(t, result, "ラウンド") // no Japanese leakage
	})

	t.Run("output uses English game-end banner", func(t *testing.T) {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "Game over!")
		assert.Contains(t, result, "Team 0 wins")
		assert.NotContains(t, result, "ゲーム終了")
	})

	t.Run("hint none uses English", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return((*domain.BridgeHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "No hint available")
	})

	t.Run("hint card with bridge-specific reason", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{CardIndex: &idx, Reason: "support_partner"})
		player := domain.NewBridgePlayer(true, 0)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		m.On("GetPlayer", 0).Return(player)
		result := p.HintOutput(m)
		assert.Contains(t, result, "support partner")
		assert.NotContains(t, result, "パートナー")
	})

	// strategic_bid is intentionally NOT in bridgeHintReasonKeys — it lives
	// in cui_common (Phase 1) so Bridge / Spades / Skat / OhHell / Napoleon
	// share one translation. Pin the shared-fallthrough path under en.
	t.Run("strategic_bid falls through to cui_common in English", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockBridgeGame)
		m.On("GetHint").Return(&domain.BridgeHint{CardIndex: &idx, Reason: "strategic_bid"})
		player := domain.NewBridgePlayer(true, 0)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		m.On("GetPlayer", 0).Return(player)
		result := p.HintOutput(m)
		assert.Contains(t, result, "strategic bid")
	})
}

// **どこから上回れるか・ダブルできるかを出す。**Web はボタンを無効化して理由まで
// 出すのに、CUI は打って拒否されるまで分からなかった (#4903)。
func TestBridgeCuiPresenter_GuidesTheBidding(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BridgeCuiPresenter)

	build := func(minOK bool, canDouble, canRedouble bool) *interfaces.MockBridgeGame {
		m, _ := setupBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseBid)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "BridgeMinLegalBid")
		m.On("BridgeMinLegalBid").Return(3, domain.BridgeBidSuitClub+1, minOK)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "BridgeCanDouble")
		m.On("BridgeCanDouble", mock.Anything).Return(canDouble)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "BridgeCanRedouble")
		m.On("BridgeCanRedouble", mock.Anything).Return(canRedouble)
		return m
	}

	out := p.Output(build(true, false, false), nil)
	assert.Contains(t, out, "上回るには")
	assert.NotContains(t, out, "ダブルできます")
	assert.NotContains(t, out, "7NT まで埋まっています")

	// 7NT まで埋まったら上は無いと言う。
	assert.Contains(t, p.Output(build(false, false, false), nil), "7NT まで埋まっています")

	assert.Contains(t, p.Output(build(true, true, false), nil), "ダブルできます")

	// **リダブルできるならリダブルだけを出す。**両方出すと矛盾して読める。
	both := p.Output(build(true, true, true), nil)
	assert.Contains(t, both, "リダブルできます")
	assert.NotContains(t, both, "相手のコントラクトにダブルできます")
}

// #5516: IsHumanTurn はダミーの手番かつ人間がデクレアラーなら true を返すのに、
// CUI はそれを一度も呼ばず、常に currentIdx の名前を出す。**人間が操作すべき
// 局面で CPU の名前が出る。** Web は yourTurnDummy で明示している。
func TestBridgeCuiPresenter_DummyTurnReadsAsTheHumansTurn(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BridgeCuiPresenter)

	// 人間 (0) がデクレアラー、ダミーは 2。手番はダミー。
	newDummyTurn := func() *domain.Bridge {
		b := domain.NewDefaultBridge()
		b.Reset()
		b.SetPhase(domain.BridgePhasePlay)
		b.SetDeclarerIdx(0)
		b.SetDummyIdx(2)
		b.SetCurrentPlayerIdx(2)
		return b
	}

	t.Run("names the human, not the dummy seat's CPU", func(t *testing.T) {
		b := newDummyTurn()
		require.True(t, b.IsHumanTurn(), "前提: ドメインは人間の手番と判定している")
		out := p.Output(b, nil)
		assert.Contains(t, out, i18n.Tf("bridge.promptPlayDummy", "seat", "CPU 2"))
	})

	// **ダミー席というだけでは足りない。** CPU がデクレアラーなら、ダミーの手番でも
	// 打つのは CPU。席だけで判定すると、人間と無関係な局面で「あなたの手番です」と出る。
	t.Run("stays quiet on a dummy seat when a CPU is the declarer", func(t *testing.T) {
		b := domain.NewDefaultBridge()
		b.Reset()
		b.SetPhase(domain.BridgePhasePlay)
		b.SetDeclarerIdx(1) // CPU がデクレアラー
		b.SetDummyIdx(3)
		b.SetCurrentPlayerIdx(3) // ダミー席の手番
		require.False(t, b.IsHumanTurn(), "前提: 人間の手番ではない")
		assert.NotContains(t, p.Output(b, nil),
			strings.Split(i18n.T("bridge.promptPlayDummy"), "{{")[0])
	})

	// **通常の手番表示は変えない。** CPU が本当に打つ番では従来どおり。
	t.Run("leaves an ordinary CPU turn alone", func(t *testing.T) {
		b := domain.NewDefaultBridge()
		b.Reset()
		b.SetPhase(domain.BridgePhasePlay)
		b.SetDeclarerIdx(1)
		b.SetDummyIdx(3)
		b.SetCurrentPlayerIdx(1)
		require.False(t, b.IsHumanTurn())
		// 席名に依らず、ダミー用の文言の骨格が出ていないこと。
		assert.NotContains(t, p.Output(b, nil),
			strings.Split(i18n.T("bridge.promptPlayDummy"), "{{")[0])
	})
}
