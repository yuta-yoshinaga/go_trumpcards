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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// setupNapoleonCuiMock creates a MockNapoleonGame with sensible defaults for CUI tests.
func setupNapoleonCuiMock() *interfaces.MockNapoleonGame {
	m := new(interfaces.MockNapoleonGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetAdjutantCard").Return((*domain.Card)(nil))
	m.On("GetAdjutantRevealed").Return(false)
	m.On("GetHighestBid").Return(0)
	m.On("GetNapoleonIdx").Return(-1)
	m.On("GetAdjutantIdx").Return(-1)
	m.On("GetCurrentTrick").Return([]*domain.NapoleonTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.NapoleonPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(domain.NapoleonWinnerUndecided)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultNapoleonConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeNapoleonPlayers() []*domain.NapoleonPlayer {
	return []*domain.NapoleonPlayer{
		domain.NewNapoleonPlayer(true),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
	}
}

func setupNapoleonCuiMockWithPlayers() (*interfaces.MockNapoleonGame, []*domain.NapoleonPlayer) {
	m := setupNapoleonCuiMock()
	players := makeNapoleonPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

func TestNapoleonCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.NapoleonCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Napoleon (ナポレオン)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック 絵札0枚 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: ビッド=未ビッド 獲得0トリック 絵札0枚 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "p <idx>")
	})

	t.Run("trump suit shown when set", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: ♠")
	})

	t.Run("adjutant card shown when set, not revealed", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantCard")
		m.On("GetAdjutantCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "副官カード: HEART 13")
		assert.Contains(t, result, "(非公開)")
	})

	t.Run("adjutant card shown when revealed", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantCard")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantRevealed")
		m.On("GetAdjutantCard").Return(domain.NewCard(domain.CardDesignDiamond, 1, false))
		m.On("GetAdjutantRevealed").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "副官カード: DIAMOND 1")
		assert.Contains(t, result, "(公開済み)")
	})

	t.Run("highest bid shown when greater than zero", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBid")
		m.On("GetHighestBid").Return(14)

		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド: 14")
	})

	// napoleonArmyMock configures Napoleon(0) + adjutant(1) with the given face
	// captures, bid, and reveal state for the army-progress tests.
	napoleonArmyMock := func(napFaces, adjFaces, bid int, revealed bool) *interfaces.MockNapoleonGame {
		m, players := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetNapoleonIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBid")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantRevealed")
		m.On("GetNapoleonIdx").Return(0)
		m.On("GetAdjutantIdx").Return(1)
		m.On("GetHighestBid").Return(bid)
		m.On("GetAdjutantRevealed").Return(revealed)
		players[0].SetPictureCards(napFaces)
		players[1].SetPictureCards(adjFaces)
		return m
	}

	t.Run("army progress excludes adjutant faces while hidden", func(t *testing.T) {
		m := napoleonArmyMock(5, 3, 14, false)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ナポレオン軍: 5/14枚")
	})

	t.Run("army progress adds adjutant faces once revealed", func(t *testing.T) {
		m := napoleonArmyMock(5, 3, 14, true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ナポレオン軍: 8/14枚")
	})

	t.Run("army progress turns green once the bid is met", func(t *testing.T) {
		origNo := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(origNo)
		m := napoleonArmyMock(14, 0, 14, false)
		result := p.Output(m, nil)
		assert.Contains(t, result, color.Green("ナポレオン軍: 14/14枚"))
	})

	t.Run("player with scores and tricks and pictureCards", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		players[1].SetCumulativeScore(50)
		players[1].SetRoundScore(10)
		players[1].SetPictureCards(3)
		players[1].SetBid(14)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=14 獲得1トリック 絵札3枚 累積50点 ラウンド10点 0枚")
	})

	t.Run("player bid pass shown", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		players[1].SetBid(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=パス")
	})

	t.Run("napoleon role shown", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		players[0].SetIsNapoleon(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "[ナポレオン]")
	})

	t.Run("adjutant role shown when revealed", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAdjutantRevealed")
		m.On("GetAdjutantRevealed").Return(true)
		players[1].SetIsAdjutant(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "[副官]")
	})

	t.Run("adjutant role hidden when not revealed", func(t *testing.T) {
		m, players := setupNapoleonCuiMockWithPlayers()
		players[1].SetIsAdjutant(true)

		result := p.Output(m, nil)
		assert.NotContains(t, result, "[副官]")
	})

	t.Run("human with no cards does not print extra cards line", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック 絵札0枚 累積0点 ラウンド0点 0枚")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック: あなた=CLOVER 3, CPU 1=CLOVER 7")
	})

	t.Run("no trick cards hides trick section", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "トリック: あなた")
		assert.NotContains(t, result, "トリック: CPU")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended napoleon wins", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerNapoleon)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "ナポレオン軍の勝利です！")
	})

	t.Run("game ended allied wins", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerAllied)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "連合軍の勝利です！")
	})

	t.Run("bid phase shows bidder and command", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "b <n>")
	})

	t.Run("bid phase CPU bidder", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBidPlayerIdx")
		m.On("GetPhase").Return(domain.NapoleonPhaseBid)
		m.On("GetBidPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: CPU 1の番")
	})

	t.Run("trump declaration phase", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseTrumpDeclaration)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札宣言フェーズ")
		assert.Contains(t, result, "t <suit> <adjSuit> <adjVal>")
	})

	t.Run("kitty exchange phase", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseKittyExchange)

		result := p.Output(m, nil)
		assert.Contains(t, result, "場札交換フェーズ")
		assert.Contains(t, result, "e <idx>")
	})

	t.Run("play phase shows current player CPU", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
		assert.Contains(t, result, "p <idx>")
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "n / next・・・次のトリックへ")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})

	t.Run("joker card in trick shown as Joker", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック: あなた=Joker")
	})
}

func TestNapoleonCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.NapoleonCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewNapoleonPlayer(true)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた", "棋譜の座席名が他の行と揃っていない")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestNapoleonCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return((*domain.NapoleonHint)(nil))

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("bid hint", func(t *testing.T) {
		bid := 14
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{
			Bid:    &bid,
			Reason: "strategic_bid",
		})

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "ビッド 14")
		assert.Contains(t, result, "戦略的なビッド")
	})

	t.Run("trump suit hint", func(t *testing.T) {
		suit := 1
		adjSuit := 3
		adjVal := 13
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{
			TrumpSuit:     &suit,
			AdjutantSuit:  &adjSuit,
			AdjutantValue: &adjVal,
			Reason:        "strategic_declare",
		})

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "切り札 ♠")
		assert.Contains(t, result, "戦略的な宣言")
	})

	t.Run("discard hint", func(t *testing.T) {
		idx := 2
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{
			DiscardIndex: &idx,
			Reason:       "strategic_discard",
		})
		player := domain.NewNapoleonPlayer(true)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		m.On("GetPlayer", 0).Return(player)
		m.On("GetHumanIdx").Return(0)

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "を捨てる")
		assert.Contains(t, result, "戦略的な捨て")
	})

	// **人間が席 0 とは限らない。**ヒントの数値自体はドメインが findHumanIdx で
	// 解決しているのに、表示だけが GetPlayer(0) を決め打ちしていた (#4689)。
	// コンストラクタは任意の並び順を受け付けるので、席がずれると別人の手札を
	// 見て「これを捨てろ」と言うことになる。
	t.Run("resolves the human seat instead of assuming index 0", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{DiscardIndex: &idx, Reason: "strategic_discard"})

		// 席 2 が人間。席 0 には別人の手札を置く。
		human := domain.NewNapoleonPlayer(true)
		human.Reset()
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		human.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		other := domain.NewNapoleonPlayer(false)
		other.Reset()
		other.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		other.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))

		m.On("GetHumanIdx").Return(2)
		m.On("GetPlayer", 2).Return(human)
		m.On("GetPlayer", 0).Return(other).Maybe()

		result := new(presenter.NapoleonCuiPresenter).HintOutput(m)
		// 人間の index 1 は ♥7。席 0 を見ていると ♦4 と言ってしまう。
		assert.Contains(t, result, "7")
		assert.NotContains(t, result, "♦4")
	})

	t.Run("play hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})
		player := domain.NewNapoleonPlayer(true)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)
		m.On("GetHumanIdx").Return(0)

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("hint with nil bid and nil card index and nil trump and nil discard", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		m.On("GetHint").Return(&domain.NapoleonHint{
			Reason: "unknown",
		})

		p := new(presenter.NapoleonCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"lead_strong":    "強いカードでリード",
			"lead_low":       "低いカードでリード",
			"trump_cut":      "切り札でカット",
			"play_joker":     "ジョーカーをプレイ",
			"discard_low":    "低いカードを捨てる",
			"unknown_reason": "unknown_reason",
		}
		for key, expected := range reasons {
			idx := 0
			m := new(interfaces.MockNapoleonGame)
			m.On("GetHint").Return(&domain.NapoleonHint{
				CardIndex: &idx,
				Reason:    key,
			})
			player := domain.NewNapoleonPlayer(true)
			player.Reset()
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)
			m.On("GetHumanIdx").Return(0)

			p := new(presenter.NapoleonCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}

// TestNapoleonCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in NapoleonCuiPresenter now follows
// the active locale. The default ja path is exercised by the assertions
// above; this suite re-runs Output / HintOutput under LANG=en and checks
// the English keys win out.
func TestNapoleonCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.NapoleonCuiPresenter)

	t.Run("hint none uses English", func(t *testing.T) {
		m := setupNapoleonCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.NapoleonHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "No hint available")
	})

	t.Run("hint reasons render in English", func(t *testing.T) {
		bid := 13
		m := setupNapoleonCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{Bid: &bid, Reason: "strategic_declare"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "bid 13")
		assert.Contains(t, result, "strategic declaration")
	})

	// strategic_bid is intentionally NOT in napoleonHintReasonKeys — it's
	// shared with Bridge, Spades, Skat, OhHell and lives in
	// sharedHintReasonKeys (cui_common). The fallthrough path in
	// hintReasonStr (per-game miss → sharedHintReasonKeys → cui_common)
	// is what makes it render. This test pins that path under LANG=en.
	t.Run("strategic_bid falls through to shared key in English", func(t *testing.T) {
		bid := 14
		m := setupNapoleonCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{Bid: &bid, Reason: "strategic_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "bid 14")
		assert.Contains(t, result, "strategic bid")
	})

	t.Run("output uses English game-end banner (Napoleon side)", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerNapoleon)
		result := p.Output(m, nil)
		assert.Contains(t, result, "The Napoleon side wins")
		assert.NotContains(t, result, "ナポレオン軍") // no Japanese leakage
	})

	t.Run("output uses English headers and prompts", func(t *testing.T) {
		m, _ := setupNapoleonCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Round: 1")
		assert.Contains(t, result, "Trick: 1")
		// Default mock leaves phase as Play, so the play prompt should render.
		assert.Contains(t, result, "Turn: ")
		assert.Contains(t, result, "p <idx>")
	})
}

// #5504: 目標点数は開始前の設定でしか見えず、対局中は round/trick は出るのに
// 到達条件だけが出ていなかった。**あと何点で決着するのかを知るには Settings を
// 開き直すしかない。**
func TestNapoleonCuiPresenter_PointLimitLine(t *testing.T) {
	p := new(presenter.NapoleonCuiPresenter)

	t.Run("shows the configured target", func(t *testing.T) {
		n := domain.NewDefaultNapoleon()
		n.Reset()
		cfg := n.GetConfig()
		cfg.PointLimit = 75
		n.SetConfig(cfg)

		assert.Contains(t, p.Output(n, nil), i18n.Tf("napoleon.pointLimitLine", "limit", "75"))
	})

	// **設定値を出していること。** 定数を書いているだけなら、変えても表示が動かない。
	t.Run("follows a settings change", func(t *testing.T) {
		n := domain.NewDefaultNapoleon()
		n.Reset()
		cfg := n.GetConfig()
		cfg.PointLimit = 30
		n.SetConfig(cfg)

		out := p.Output(n, nil)
		assert.Contains(t, out, i18n.Tf("napoleon.pointLimitLine", "limit", "30"))
		assert.NotContains(t, out, i18n.Tf("napoleon.pointLimitLine", "limit", "75"))
	})
}
