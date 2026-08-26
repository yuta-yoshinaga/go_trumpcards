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

func makeSchafkopfPlayers() []*domain.SchafkopfPlayer {
	cfg := domain.DefaultSchafkopfConfig()
	return []*domain.SchafkopfPlayer{
		domain.NewSchafkopfPlayer(true, cfg.StartChips),
		domain.NewSchafkopfPlayer(false, cfg.StartChips),
		domain.NewSchafkopfPlayer(false, cfg.StartChips),
		domain.NewSchafkopfPlayer(false, cfg.StartChips),
		domain.NewSchafkopfPlayer(false, cfg.StartChips),
	}
}

func setupSchafkopfCuiMock() *interfaces.MockSchafkopfGame {
	m := new(interfaces.MockSchafkopfGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetPickerIdx").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCalledSuit").Return(0)
	m.On("GetCallableSuits").Return([]int{domain.CardDesignSpade, domain.CardDesignHeart})
	m.On("GetPassCount").Return(0)
	m.On("GetContract").Return(domain.SchafkopfContractRufspiel)
	m.On("GetSoloSuit").Return(0)
	m.On("GetBeatableContracts").Return([]domain.SchafkopfContract{
		domain.SchafkopfContractRufspiel, domain.SchafkopfContractWenz, domain.SchafkopfContractSolo,
	})
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SchafkopfPhasePick)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsPartnerRevealed").Return(false)
	m.On("GetRoundPickerPoints").Return(0)
	m.On("GetRoundMultiplier").Return(1)
	m.On("GetRoundPickerWon").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSchafkopfCuiMockWithPlayers() (*interfaces.MockSchafkopfGame, []*domain.SchafkopfPlayer) {
	m := setupSchafkopfCuiMock()
	players := makeSchafkopfPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

// **どのスートを呼べるかを出す。**Web は呼べるスートだけボタンを描くのに、
// CUI はコマンド構文しか示さず試行錯誤させていた (#4916)。
func TestSchafkopfCuiPresenter_ListsTheCallableSuits(t *testing.T) {
	p := new(presenter.SchafkopfCuiPresenter)

	callMock := func(suits []int) *interfaces.MockSchafkopfGame {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.On("GetPhase").Return(domain.SchafkopfPhaseCall)
		m.On("GetPickerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallableSuits")
		m.On("GetCallableSuits").Return(suits)
		return m
	}

	t.Run("lists the suits with the number the command takes", func(t *testing.T) {
		// ♥ は切り札なので呼べない。呼べるのは ♠ ♣ ♦ の 3 つ。
		out := p.Output(callMock([]int{domain.CardDesignSpade, domain.CardDesignDiamond}), nil)
		// **行ごと照合する。**promptCallHelp が `1=♠ 2=♣ 4=♦` を常に出しているので、
		// 断片で見ると呼べない ♣ の有無を判定できない。
		// 番号を添えるのは、c コマンドが取るのが記号ではなく数字だから。
		assert.Contains(t, out, "呼べるスート: 1=♠ 4=♦\n")
	})

	// 呼べるスートが 0 件でもクラッシュしない (受け入れ条件2)。
	t.Run("says so when no suit can be called", func(t *testing.T) {
		out := p.Output(callMock(nil), nil)
		assert.Contains(t, out, "呼べるスートがありません")
		assert.NotContains(t, out, "呼べるスート: ")
	})

	// コールフェーズ以外には出さない。
	t.Run("confined to the call phase", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		assert.NotContains(t, p.Output(m, nil), "呼べるスート")
	})
}

func TestSchafkopfCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SchafkopfCuiPresenter)

	t.Run("pick phase shows blind count and prompt", func(t *testing.T) {
		m, players := setupSchafkopfCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Schafkopf")
		assert.NotEmpty(t, result)
	})

	t.Run("call phase shows picker prompt", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPhase").Return(domain.SchafkopfPhaseCall)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("play phase shows current player", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchafkopfPhasePlay)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchafkopfPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows picker won result", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundPickerWon")
		m.On("GetPhase").Return(domain.SchafkopfPhaseRoundEnd)
		m.On("GetRoundPickerWon").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("schafkopf.roundPickerWon"))
	})

	t.Run("round end shows picker lost result", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchafkopfPhaseRoundEnd)
		result := p.Output(m, nil) // default GetRoundPickerWon = false
		assert.Contains(t, result, i18n.T("schafkopf.roundPickerLost"))
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("picker with partner revealed", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCalledSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsPartnerRevealed")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPartnerIdx").Return(2)
		m.On("GetCalledSuit").Return(domain.CardDesignClover)
		m.On("IsPartnerRevealed").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestSchafkopfCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SchafkopfCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SchafkopfHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("pick hint", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SchafkopfHint{Pick: true, Reason: "pick_take"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("call suit hint", func(t *testing.T) {
		m, _ := setupSchafkopfCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SchafkopfHint{Suit: domain.CardDesignClover, Reason: "call_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("play hint", func(t *testing.T) {
		m, players := setupSchafkopfCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.SchafkopfHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSchafkopfCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SchafkopfCuiPresenter)
	m := new(interfaces.MockSchafkopfGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "pick", Detail: "You picks up the blind"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewSchafkopfPlayer(true, 0)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "pick")
}
