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
)

// setupOmiCuiMock creates a MockOmiGame with sensible defaults for CUI tests.
func setupOmiCuiMock() *interfaces.MockOmiGame {
	m := new(interfaces.MockOmiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OmiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetValidPlayIndices", mock.Anything).Return(([]int)(nil)).Maybe()
	m.On("GetTrumpCallerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1) // Spade
	m.On("GetMakerTeam").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultOmiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeOmiPlayers() []*domain.OmiPlayer {
	return []*domain.OmiPlayer{
		domain.NewOmiPlayer(true, 0),
		domain.NewOmiPlayer(false, 1),
		domain.NewOmiPlayer(false, 0),
		domain.NewOmiPlayer(false, 1),
	}
}

func setupOmiCuiMockWithPlayers() (*interfaces.MockOmiGame, []*domain.OmiPlayer) {
	m := setupOmiCuiMock()
	players := makeOmiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestOmiCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmiCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupOmiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Omi (オミ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "ディーラー: あなた")
		assert.Contains(t, result, "指名者: あなた")
		assert.Contains(t, result, "切り札: SPADE (指名: あなた / チーム0)")
		assert.Contains(t, result, "残り4枚が配られ、全員の手札が8枚になりました。")
		assert.Contains(t, result, "得点規則: 5トリック以上で1点、全取り(8トリック)で2点、4-4引き分けは0点")
		assert.Contains(t, result, "チーム0: 0点 (0トリック)  チーム1: 0点 (0トリック)")
		assert.Contains(t, result, "あなた: チーム0 獲得0トリック 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: チーム1 獲得0トリック 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "p <i> (play)")
	})

	t.Run("trump suit shown", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: SPADE (指名: あなた / チーム0)")
	})

	t.Run("trump suit undecided", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("team scores shown with tricks", func(t *testing.T) {
		m, players := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)
		players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 1, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "チーム0: 5点 (1トリック)  チーム1: 3点 (1トリック)")
	})

	t.Run("dealer shown", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディーラー: あなた")
	})

	t.Run("trump caller shown", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "指名者: あなた")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック: あなた=CLOVER 3, CPU 1=CLOVER 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner team", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "チーム0の勝利です！")
	})

	t.Run("call trump phase shows caller and command", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseCallTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, "各自に4枚配られました。切り札を宣言してください。")
		assert.Contains(t, result, "コールトランプフェーズ: あなたの番")
		assert.Contains(t, result, "t <1-4> (1=♠ 2=♣ 3=♥ 4=♦)")
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupOmiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "n (next trick)")
	})
}

func TestOmiCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmiCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockOmiGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		m.On("GetPlayer", mock.Anything).Return(domain.NewOmiPlayer(true, 0)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})
}

func TestOmiCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockOmiGame)
		m.On("GetHint").Return((*domain.OmiHint)(nil))

		p := new(presenter.OmiCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません。")
	})

	t.Run("call suit hint", func(t *testing.T) {
		suit := 3
		m := new(interfaces.MockOmiGame)
		m.On("GetHint").Return(&domain.OmiHint{
			Suit:   &suit,
			Reason: "strategic_call",
		})

		p := new(presenter.OmiCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HEARTを宣言")
		assert.Contains(t, result, "手札が強いスート")
	})

	t.Run("play hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockOmiGame)
		m.On("GetHint").Return(&domain.OmiHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})
		player := domain.NewOmiPlayer(true, 0)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.OmiCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに従う")
	})

	t.Run("hint with nil fields returns no hint", func(t *testing.T) {
		m := new(interfaces.MockOmiGame)
		m.On("GetHint").Return(&domain.OmiHint{
			Reason: "unknown",
		})

		p := new(presenter.OmiCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません。")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"strategic_call": "手札が強いスート",
			"normal_play":    "通常プレイ",
			"lead_trump":     "切り札リード",
			"lead_strong":    "最強カードでリード",
			"follow_suit":    "リードスートに従う",
			"trump_cut":      "切り札で勝負",
			"discard_weak":   "最弱カードを捨てる",
			"unknown_reason": "unknown_reason",
		}
		for key, expected := range reasons {
			idx := 0
			m := new(interfaces.MockOmiGame)
			m.On("GetHint").Return(&domain.OmiHint{
				CardIndex: &idx,
				Reason:    key,
			})
			player := domain.NewOmiPlayer(true, 0)
			player.Reset()
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)

			p := new(presenter.OmiCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}

func TestOmiCuiPresenter_MarksPlayableCards(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmiCuiPresenter)

	withPlayable := func(valid []int, phase domain.OmiPhase, turn int) *interfaces.MockOmiGame {
		m, players := setupOmiCuiMockWithPlayers()
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
		out := p.Output(withPlayable([]int{0, 2}, domain.OmiPhasePlay, 0), nil)
		assert.Equal(t, 2, strings.Count(out, "*"), "印が付くのは2枚だけ")
		assert.Contains(t, out, "[0]")
	})

	t.Run("stars nothing while a CPU is to act", func(t *testing.T) {
		out := p.Output(withPlayable([]int{0, 1, 2}, domain.OmiPhasePlay, 1), nil)
		assert.NotContains(t, out, "*")
	})

	t.Run("stars nothing outside the play phase", func(t *testing.T) {
		out := p.Output(withPlayable([]int{0, 1, 2}, domain.OmiPhaseCallTrump, 0), nil)
		assert.NotContains(t, out, "*")
	})
}
