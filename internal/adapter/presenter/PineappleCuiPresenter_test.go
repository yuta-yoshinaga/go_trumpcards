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
			{PlayerIdx: 0, HandName: "Flush", WonAmount: 100},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: Flush")
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

	t.Run("discard prompt says 2 cards for a 4-card deal (Irish Poker)", func(t *testing.T) {
		m, _ := setupPineappleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetInitialDealCount")
		m.On("GetPhase").Return(domain.PineapplePhaseDiscard)
		m.On("GetInitialDealCount").Return(4)
		result := p.Output(m, nil)
		assert.Contains(t, result, "手札から2枚選んで")
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
