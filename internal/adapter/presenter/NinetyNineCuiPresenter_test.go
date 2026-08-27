//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupNinetyNineCuiMock() *interfaces.MockNinetyNineGame {
	m := new(interfaces.MockNinetyNineGame)
	m.On("GetDealNumber").Return(1)
	m.On("GetTargetScore").Return(100)
	m.On("GetHandSize").Return(9)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.NinetyNinePhasePlay)
	m.On("GetRoundSuccessBonus").Return(0).Maybe()
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultNinetyNineConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupNinetyNineCuiMockWithPlayers() (*interfaces.MockNinetyNineGame, []*domain.NinetyNinePlayer) {
	m := setupNinetyNineCuiMock()
	players := makeNinetyNinePlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestNinetyNineCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.NinetyNineCuiPresenter)

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Ninety-Nine")
		assert.Contains(t, result, "ディール: 1")
		assert.Contains(t, result, "切り札:")
		assert.Contains(t, result, "手番:")
	})

	t.Run("bid phase", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディール終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		result := p.Output(m, domain.ErrInvalidPlay)
		assert.Contains(t, result, "invalid play")
	})
}

func TestNinetyNineCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.NinetyNineCuiPresenter)

	t.Run("bury hint", func(t *testing.T) {
		m := setupNinetyNineCuiMock()
		m.On("GetHint").Return(&domain.NinetyNineHint{BuryIndices: []int{0, 1, 2}, Reason: "strategic_bury"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "0 1 2")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupNinetyNineCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		cardIdx := 0
		m.On("GetHint").Return(&domain.NinetyNineHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
	})

	t.Run("nil hint", func(t *testing.T) {
		m := setupNinetyNineCuiMock()
		m.On("GetHint").Return((*domain.NinetyNineHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint with nil bury and nil cardIndex", func(t *testing.T) {
		m := setupNinetyNineCuiMock()
		m.On("GetHint").Return(&domain.NinetyNineHint{Reason: "unknown"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("card hint with no human player", func(t *testing.T) {
		m := setupNinetyNineCuiMock()
		cpuPlayers := []*domain.NinetyNinePlayer{
			domain.NewNinetyNinePlayer(false),
			domain.NewNinetyNinePlayer(false),
			domain.NewNinetyNinePlayer(false),
		}
		m.On("GetPlayerCnt").Return(3)
		m.On("GetPlayer", 0).Return(cpuPlayers[0])
		m.On("GetPlayer", 1).Return(cpuPlayers[1])
		m.On("GetPlayer", 2).Return(cpuPlayers[2])
		cardIdx := 0
		m.On("GetHint").Return(&domain.NinetyNineHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestNinetyNineCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.NinetyNineCuiPresenter)
	m := setupNinetyNineCuiMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You buries 3 and declares 3"},
	})
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

// round is 10+bid+bonus, so the bonus vanishes into it. The sole-success +30 is
// the point of this game's scoring; it was only ever in the action log string.
func TestNinetyNineCuiPresenter_SuccessBonusLine(t *testing.T) {
	p := new(presenter.NinetyNineCuiPresenter)

	build := func(bonus int) *interfaces.MockNinetyNineGame {
		m, _ := setupNinetyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundSuccessBonus")
		m.On("GetRoundSuccessBonus").Return(bonus).Maybe()
		return m
	}

	t.Run("names the bonus that was applied", func(t *testing.T) {
		assert.Contains(t, p.Output(build(30), nil), i18n.Tf("ninetynine.successBonusLine", "bonus", "30"))
	})

	t.Run("says nothing when nobody made their bid", func(t *testing.T) {
		// A bonus of 0 means no one succeeded; printing "+0" would imply someone did.
		assert.NotContains(t, p.Output(build(0), nil), i18n.T("ninetynine.successBonusLine"))
	})
}
