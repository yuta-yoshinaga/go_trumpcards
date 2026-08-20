package presenter_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeIndianPokerForPresenter() (*domain.IndianPoker, []*domain.IndianPokerPlayer) {
	tc := domain.NewTrumpCards(0)
	players := domain.NewIndianPokerPlayers()
	ip := domain.NewIndianPoker(tc, players, domain.DefaultIndianPokerConfig())
	return ip, players
}

func TestIndianPokerCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.IndianPokerCuiPresenter)

	t.Run("initial state header", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()

		result := p.Output(ip, nil)
		assert.Contains(t, result, "Indian Poker")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "リミット:")
		assert.Contains(t, result, "アンティ:")
	})

	t.Run("betting phase human card hidden", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "カード: ??")
		assert.NotContains(t, result, "♠10")
	})

	t.Run("betting phase shows human estimated equity", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		// A low visible opponent card leaves most ranks above it.
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "推定勝率:")
	})

	t.Run("betting shows check-available on human turn with no outstanding bet", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCurrentTurn(0)
		ip.SetLastBet(0)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "チェック可能")
		assert.NotContains(t, result, "コール:")
	})

	t.Run("betting shows call amount and min raise on human turn", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCurrentTurn(0)
		ip.SetLastBet(50)
		ip.SetMinRaise(20)
		players[0].SetCurrentBet(10)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "コール: 40") // 50 - 10 already committed
		assert.Contains(t, result, "ミニマムレイズ: 20")
		assert.NotContains(t, result, "チェック可能")
	})

	t.Run("betting hides call info when it is not the human's turn", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCurrentTurn(1) // a CPU is to act
		ip.SetLastBet(50)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "推定勝率:")
		assert.NotContains(t, result, "コール:")
		assert.NotContains(t, result, "チェック可能")
	})

	t.Run("showdown phase hides equity", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "推定勝率:")
	})

	t.Run("showdown phase human card visible", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "♠10")
		assert.NotContains(t, result, "カード: ??")
	})

	t.Run("end phase human card visible", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "♥5")
	})

	t.Run("CPU card always visible", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(ip, nil)
		assert.Contains(t, result, "♣7")
	})

	t.Run("folded player", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[1].SetFolded(true)

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[1].SetAllIn(true)

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].SetCurrentBet(50)

		result := p.Output(ip, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		ip, players := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].SetCurrentBet(0)

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("player without cards", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		// No cards added, so GetCardsSize() == 0

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "カード:")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCpuActions([]domain.IndianPokerCpuAction{
			{PlayerIdx: 1, Action: domain.IndianPokerActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.IndianPokerActionRaise, Amount: 30},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("CPU action without amount", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCpuActions([]domain.IndianPokerCpuAction{
			{PlayerIdx: 1, Action: domain.IndianPokerActionFold, Amount: 0},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCpuActions([]domain.IndianPokerCpuAction{})

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results displayed", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{
				PlayerIdx: 0,
				Card:      domain.NewCard(domain.CardDesignSpade, 14, false),
				CardRank:  14,
				WonAmount: 100,
			},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{
				PlayerIdx: 1,
				Card:      domain.NewCard(domain.CardDesignHeart, 13, false),
				CardRank:  13,
				WonAmount: 50,
			},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown result without card (folded winner)", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{
				PlayerIdx: 1,
				Card:      nil,
				WonAmount: 80,
			},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "80チップ獲得")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignDiamond, 2, false), WonAmount: 0},
		})

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-showdown phase", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{PlayerIdx: 0, WonAmount: 100},
		})

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("empty results not shown", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		ip.SetRoundResults([]domain.IndianPokerResult{})

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetGameEndFlag(true)

		result := p.Output(ip, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "ゲーム終了")
	})

	t.Run("betting limit out of range hides limit line", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		cfg := domain.IndianPokerConfig{Ante: 10, InitChips: 1000, BettingLimit: 99}
		ip.SetConfig(cfg)
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		assert.NotContains(t, result, "リミット:")
	})

	t.Run("all action names in CPU actions", func(t *testing.T) {
		ip, _ := makeIndianPokerForPresenter()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCpuActions([]domain.IndianPokerCpuAction{
			{PlayerIdx: 1, Action: domain.IndianPokerActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.IndianPokerActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.IndianPokerActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.IndianPokerActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.IndianPokerActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.IndianPokerActionAllIn, Amount: 100},
			{PlayerIdx: 1, Action: 99, Amount: 0},
		})

		result := p.Output(ip, nil)
		assert.Contains(t, result, "フォールド")
		assert.Contains(t, result, "チェック")
		assert.Contains(t, result, "コール")
		assert.Contains(t, result, "ベット")
		assert.Contains(t, result, "レイズ")
		assert.Contains(t, result, "オールイン")
		assert.Contains(t, result, "不明")
	})
}

func TestIndianPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.IndianPokerCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewIndianPokerPlayer(true, domain.HoldemPlayStyle(0))).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "bet")
		assert.Contains(t, result, "bet 100")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewIndianPokerPlayer(true, domain.HoldemPlayStyle(0))).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// #5505: Web は computeIndianPokerEquity という別実装で勝率を出しており、
// ドメインの estimateOwnStrength とズレていた。**同ランクのスートタイブレークを
// 数えていない**ぶん低めに出る。CUI と Web が同じ局面で違う数字を出さないよう、
// 同じ値であることをここで固定する。
func TestIndianPoker_CuiAndWebShowTheSameEquity(t *testing.T) {
	cui := new(presenter.IndianPokerCuiPresenter)
	web := new(presenter.IndianPokerWebPresenter)

	ip := domain.NewDefaultIndianPoker()
	require.NoError(t, ip.Reset())
	ip.SetPhase(domain.IndianPokerPhaseBetting)

	var out controller.IndianPokerWebOutput
	require.NoError(t, json.Unmarshal([]byte(web.Output(ip, nil)), &out))

	// CUI の行に出ている数字と、Web が送る数字が一致すること。
	assert.Contains(t, cui.Output(ip, nil),
		i18n.Tf("indianpoker.equityLine", "pct", strconv.Itoa(out.EstimatedStrength)))
	// 0 固定のフィールドを比べているだけ、にならないよう範囲も確かめる。
	assert.GreaterOrEqual(t, out.EstimatedStrength, 0)
	assert.LessOrEqual(t, out.EstimatedStrength, 100)
}
