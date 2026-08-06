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

func setupCallBreakCuiMock() *interfaces.MockCallBreakGame {
	m := new(interfaces.MockCallBreakGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetSpadesBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CallBreakPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultCallBreakConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeCallBreakPlayers() []*domain.CallBreakPlayer {
	return []*domain.CallBreakPlayer{
		domain.NewCallBreakPlayer(true),
		domain.NewCallBreakPlayer(false),
		domain.NewCallBreakPlayer(false),
		domain.NewCallBreakPlayer(false),
	}
}

func setupCallBreakCuiMockWithPlayers() (*interfaces.MockCallBreakGame, []*domain.CallBreakPlayer) {
	m := setupCallBreakCuiMock()
	players := makeCallBreakPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCallBreakCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CallBreakCuiPresenter)

	t.Run("initial header and player info", func(t *testing.T) {
		m, players := setupCallBreakCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Call Break (コールブレイク)")
		assert.Contains(t, result, "ラウンド: 1 / 5")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "スペードブレイク: なし")
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック バッグ0 累積0.0点 ラウンド0.0点 2枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
	})

	t.Run("formatted decimal score", func(t *testing.T) {
		m, players := setupCallBreakCuiMockWithPlayers()
		players[1].SetCumulativeScore(41)
		players[1].SetRoundScore(41)
		players[1].SetBid(4)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=4 獲得1トリック バッグ0 累積4.1点 ラウンド4.1点")
	})

	t.Run("formatted negative score", func(t *testing.T) {
		m, players := setupCallBreakCuiMockWithPlayers()
		players[1].SetCumulativeScore(-30)
		players[1].SetRoundScore(-30)
		players[1].SetBid(3)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=3 獲得0トリック バッグ0 累積-3.0点 ラウンド-3.0点")
	})

	t.Run("spades broken shows yes", func(t *testing.T) {
		m, _ := setupCallBreakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSpadesBroken")
		m.On("GetSpadesBroken").Return(true)
		assert.Contains(t, p.Output(m, nil), "スペードブレイク: あり")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupCallBreakCuiMockWithPlayers()
		err := errors.New("invalid play")
		assert.Contains(t, p.Output(m, err), "invalid play")
	})

	t.Run("game ended human win", func(t *testing.T) {
		m, _ := setupCallBreakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("bid phase", func(t *testing.T) {
		m, _ := setupCallBreakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CallBreakPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "b <n>")
	})

	t.Run("trick end and round end", func(t *testing.T) {
		m, _ := setupCallBreakCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CallBreakPhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})
}

// **Web は cb-bags-counter で常時出しているのに CUI には無かった (#4752)。**
// バッグ (宣言超過トリック) の蓄積は長期スコアに直結する。
func TestCallBreakCuiPresenter_ShowsBags(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CallBreakCuiPresenter)

	m, players := setupCallBreakCuiMockWithPlayers()
	// 宣言 2 に対して 5 トリック = バッグ 3。
	players[1].SetBid(2)
	for i := 0; i < 5; i++ {
		players[1].AddTrick([]*domain.Card{nil})
	}
	// 宣言ちょうど = バッグ 0 (超過だけを数えていることの確認)。
	players[2].SetBid(1)
	players[2].AddTrick([]*domain.Card{nil})

	result := p.Output(m, nil)
	assert.Contains(t, result, "CPU 1: ビッド=2 獲得5トリック バッグ3")
	assert.Contains(t, result, "CPU 2: ビッド=1 獲得1トリック バッグ0")
}

func TestCallBreakCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CallBreakCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCallBreakGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
	})

	t.Run("game not ended yields placeholder", func(t *testing.T) {
		m := new(interfaces.MockCallBreakGame)
		m.On("GetGameEndFlag").Return(false)
		assert.Contains(t, p.ActionLogOutput(m), "棋譜はありません")
	})
}

func TestCallBreakCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockCallBreakGame)
		m.On("GetHint").Return((*domain.CallBreakHint)(nil))
		assert.Contains(t, (&presenter.CallBreakCuiPresenter{}).HintOutput(m), "ヒントはありません")
	})

	t.Run("bid hint", func(t *testing.T) {
		bid := 3
		m := new(interfaces.MockCallBreakGame)
		m.On("GetHint").Return(&domain.CallBreakHint{Bid: &bid, Reason: "strategic_bid"})
		result := (&presenter.CallBreakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "ビッド 3")
		assert.Contains(t, result, "戦略的なビッド")
	})

	t.Run("nil bid and nil card index", func(t *testing.T) {
		m := new(interfaces.MockCallBreakGame)
		m.On("GetHint").Return(&domain.CallBreakHint{Reason: "unknown"})
		assert.Contains(t, (&presenter.CallBreakCuiPresenter{}).HintOutput(m), "ヒントはありません")
	})

	t.Run("play hint with trump_cut reason", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockCallBreakGame)
		m.On("GetHint").Return(&domain.CallBreakHint{CardIndex: &idx, Reason: "trump_cut"})
		player := domain.NewCallBreakPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		m.On("GetPlayer", 0).Return(player)

		result := (&presenter.CallBreakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "スペードでカット")
	})

	t.Run("play hint shared reason fallback", func(t *testing.T) {
		idx := 0
		m := new(interfaces.MockCallBreakGame)
		m.On("GetHint").Return(&domain.CallBreakHint{CardIndex: &idx, Reason: "follow_suit"})
		player := domain.NewCallBreakPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		m.On("GetPlayer", 0).Return(player)
		result := (&presenter.CallBreakCuiPresenter{}).HintOutput(m)
		assert.Contains(t, result, "リードスートに追随")
	})
}
