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

// setupEuchreCuiMock creates a MockEuchreGame with sensible defaults for CUI tests.
func setupEuchreCuiMock() *interfaces.MockEuchreGame {
	m := new(interfaces.MockEuchreGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.EuchrePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetValidPlayIndices", mock.Anything).Return(([]int)(nil)).Maybe()
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1) // Spade
	m.On("GetFaceUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerTeam").Return(0)
	m.On("GetGoingAlone").Return(false)
	m.On("GetGoingAlonePlayerIdx").Return(-1)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultEuchreConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeEuchrePlayers() []*domain.EuchrePlayer {
	return []*domain.EuchrePlayer{
		domain.NewEuchrePlayer(true, 0),
		domain.NewEuchrePlayer(false, 1),
		domain.NewEuchrePlayer(false, 0),
		domain.NewEuchrePlayer(false, 1),
	}
}

func setupEuchreCuiMockWithPlayers() (*interfaces.MockEuchreGame, []*domain.EuchrePlayer) {
	m := setupEuchreCuiMock()
	players := makeEuchrePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestEuchreCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.EuchreCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupEuchreCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Euchre (ユーカー)")
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
		m, _ := setupEuchreCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: SPADE (メイカー: チーム0)")
	})

	t.Run("trump suit undecided", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("face up card shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetFaceUpCard")
		m.On("GetFaceUpCard").Return(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "表向きカード: HEART 11")
	})

	t.Run("going alone shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGoingAlone")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGoingAlonePlayerIdx")
		m.On("GetGoingAlone").Return(true)
		m.On("GetGoingAlonePlayerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゴーアローン: あなた")
		// Player 0's same-team partner (seat 2) sits out → marker shown exactly once.
		assert.Contains(t, result, i18n.T("euchre.sittingOut"))
		assert.Equal(t, 1, strings.Count(result, i18n.T("euchre.sittingOut")))
	})

	t.Run("no sitting-out marker in a normal hand", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.NotContains(t, result, i18n.T("euchre.sittingOut"))
	})

	t.Run("team scores shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)

		result := p.Output(m, nil)
		assert.Contains(t, result, "チーム0: 5点  チーム1: 3点")
	})

	t.Run("dealer shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディーラー: あなた")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
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
		m, _ := setupEuchreCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "トリック: あなた")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner team", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "チーム0の勝利です！")
	})

	t.Run("pickup phase shows bidder and command", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhasePickUp)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ピックアップフェーズ: あなたの番")
		assert.Contains(t, result, "o (order up)")
		assert.Contains(t, result, "pa (pass)")
	})

	t.Run("call trump phase shows bidder and command", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseCallTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, "コールトランプフェーズ: あなたの番")
		assert.Contains(t, result, "c <suit> (call)")
		assert.Contains(t, result, "pa (pass)")
	})

	t.Run("discard phase shows command", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <i> (discard)")
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "n (next trick)")
	})

	t.Run("round end phase shows next round command", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr (next round)")
	})

	t.Run("human with no cards does not print extra cards line", func(t *testing.T) {
		m, _ := setupEuchreCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: チーム0 獲得0トリック 0枚")
	})

	t.Run("player with tricks", func(t *testing.T) {
		m, players := setupEuchreCuiMockWithPlayers()
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: チーム1 獲得1トリック 0枚")
	})
}

func TestEuchreCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.EuchreCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestEuchreCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return((*domain.EuchreHint)(nil))

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("order up hint", func(t *testing.T) {
		orderUp := true
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			OrderUp: &orderUp,
			Reason:  "strong_hand",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "オーダーアップ")
		assert.Contains(t, result, "強い手札")
	})

	t.Run("order up alone hint", func(t *testing.T) {
		orderUp := true
		goAlone := true
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			OrderUp: &orderUp,
			GoAlone: &goAlone,
			Reason:  "strong_hand",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "オーダーアップ")
		assert.Contains(t, result, "ゴーアローン")
	})

	t.Run("pass hint", func(t *testing.T) {
		orderUp := false
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			OrderUp: &orderUp,
			Reason:  "weak_hand",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "パス")
		assert.Contains(t, result, "弱い手札")
	})

	t.Run("call suit hint", func(t *testing.T) {
		suit := 3
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			Suit:   &suit,
			Reason: "strong_hand",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HEARTをコール")
	})

	t.Run("call suit alone hint", func(t *testing.T) {
		suit := 2
		goAlone := true
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			Suit:    &suit,
			GoAlone: &goAlone,
			Reason:  "strong_hand",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "CLOVERをコール")
		assert.Contains(t, result, "ゴーアローン")
	})

	t.Run("play hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})
		player := domain.NewEuchrePlayer(true, 0)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("hint with nil fields returns no hint", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		m.On("GetHint").Return(&domain.EuchreHint{
			Reason: "unknown",
		})

		p := new(presenter.EuchreCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"trump_cut":       "切り札でカット",
			"discard_weakest": "最弱カードを捨てる",
			"lead_strong":     "強いカードでリード",
			"lead_low":        "低いカードでリード",
			"unknown_reason":  "unknown_reason",
		}
		for key, expected := range reasons {
			idx := 0
			m := new(interfaces.MockEuchreGame)
			m.On("GetHint").Return(&domain.EuchreHint{
				CardIndex: &idx,
				Reason:    key,
			})
			player := domain.NewEuchrePlayer(true, 0)
			player.Reset()
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)

			p := new(presenter.EuchreCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}

// **レフトボーア (同色の別スートの J) が切り札扱いになるという分かりにくい
// ルールを含むので、CUI では出せない札を選んでエラーを受け取るまで気づけな
// かった (#4781)。**Web は同じ判定で合法な札に枠線を付けている。
func TestEuchreCuiPresenter_MarksPlayableCards(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.EuchreCuiPresenter)

	withPlayable := func(valid []int, phase domain.EuchrePhase, turn int) *interfaces.MockEuchreGame {
		m, players := setupEuchreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetValidPlayIndices", mock.Anything).Return(valid).Maybe()
		m.On("GetPhase").Return(phase)
		m.On("GetCurrentPlayerIdx").Return(turn)
		players[0].Reset()
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 11, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 12, false),
		} {
			players[0].AddCard(c)
		}
		return m
	}

	t.Run("stars only the cards that may be played", func(t *testing.T) {
		out := p.Output(withPlayable([]int{0, 2}, domain.EuchrePhasePlay, 0), nil)
		assert.Equal(t, 2, strings.Count(out, "*"), "印が付くのは2枚だけ")
		assert.Contains(t, out, "[0]")
	})

	// **手番でないときは印を付けない。**別のプレイヤーの合法手を人間の手札に
	// 当てると、出せない札に印が付く。
	t.Run("stars nothing while a CPU is to act", func(t *testing.T) {
		out := p.Output(withPlayable([]int{0, 1, 2}, domain.EuchrePhasePlay, 1), nil)
		assert.NotContains(t, out, "*")
	})

	t.Run("stars nothing outside the play phase", func(t *testing.T) {
		out := p.Output(withPlayable([]int{0, 1, 2}, domain.EuchrePhasePickUp, 0), nil)
		assert.NotContains(t, out, "*")
	})
}
