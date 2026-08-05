//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSpiteAndMaliceCuiMockDefaults(g *interfaces.MockSpiteAndMaliceGame) {
	g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetCurrent").Return(0).Maybe()
	g.On("GetWinner").Return(-1).Maybe()
	g.On("GetStockSize").Return(50).Maybe()
	g.On("GetCompletedSize").Return(0).Maybe()

	var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
	foundations[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	g.On("GetFoundations").Return(foundations).Maybe()

	player := domain.NewSpiteAndMalicePlayer(false)
	player.AddToHand(domain.NewCard(domain.CardDesignSpade, 5, false))
	player.AddToGoal(domain.NewCard(domain.CardDesignHeart, 9, false))
	cpu := domain.NewSpiteAndMalicePlayer(true)
	cpu.AddToHand(domain.NewCard(domain.CardDesignDiamond, 3, false))
	cpu.AddToGoal(domain.NewCard(domain.CardDesignClover, 7, false))
	g.On("IsGoalTopPlayable", 0).Return(false).Maybe()
	g.On("IsGoalTopPlayable", 1).Return(false).Maybe()
	g.On("GetPlayer", 0).Return(player).Maybe()
	g.On("GetPlayer", 1).Return(cpu).Maybe()
}

// **ゴール札を先に空にした方が勝ち。**Web は出せる状態のゴール札にリングを出す
// のに、CUI は札と残り枚数だけで毎ターン全基礎札と見比べさせていた (#4876)。
func TestSpiteAndMaliceCuiPresenter_MarksPlayableGoalTop(t *testing.T) {
	t.Run("marks the human's goal top when it can go out", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		// **defaults より先に登録する。**testify は最初に一致した期待値を使う。
		g.On("IsGoalTopPlayable", 0).Return(true)
		setupSpiteAndMaliceCuiMockDefaults(g)
		result := new(SpiteAndMaliceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "今出せます")
		// 印は 1 つだけ = CPU 側には付かない。
		assert.Equal(t, 1, strings.Count(result, "今出せます"))
	})

	t.Run("no mark when it cannot", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceCuiMockDefaults(g)
		assert.NotContains(t, new(SpiteAndMaliceCuiPresenter).Output(g, nil), "今出せます")
	})

	t.Run("no mark for the cpu even when its goal top is playable", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("IsGoalTopPlayable", 1).Return(true)
		setupSpiteAndMaliceCuiMockDefaults(g)
		assert.NotContains(t, new(SpiteAndMaliceCuiPresenter).Output(g, nil), "今出せます")
	})
}

func TestSpiteAndMaliceCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceCuiMockDefaults(g)
		result := new(SpiteAndMaliceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Spite and Malice")
		assert.Contains(t, result, "[F0")
		assert.Contains(t, result, "[P0")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceCuiMockDefaults(g)
		result := new(SpiteAndMaliceCuiPresenter).Output(g, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	// human hand visible / cpu hand hidden — verifies the bug from PR #1503
	// review (presenters used to leak the CPU's hand on every turn).
	t.Run("hides cpu hand and shows human hand on human turn", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		setupSpiteAndMaliceCuiMockDefaults(g)
		result := new(SpiteAndMaliceCuiPresenter).Output(g, nil)
		// 人間の手札 (SPADE 5) は公開
		assert.Contains(t, result, "SPADE 5")
		// CPU の手札 (DIAMOND 3) は非公開、枚数のみ
		assert.NotContains(t, result, "DIAMOND 3")
		assert.Contains(t, result, "手札: 1枚")
	})

	t.Run("hides cpu hand even on cpu turn", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("GetCurrent").Return(domain.SpiteAndMaliceCpuIdx).Maybe()
		g.On("GetWinner").Return(-1).Maybe()
		g.On("GetStockSize").Return(50).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		human := domain.NewSpiteAndMalicePlayer(false)
		human.AddToHand(domain.NewCard(domain.CardDesignSpade, 5, false))
		cpu := domain.NewSpiteAndMalicePlayer(true)
		cpu.AddToHand(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.On("GetPlayer", 0).Return(human).Maybe()
		g.On("GetPlayer", 1).Return(cpu).Maybe()
		result := new(SpiteAndMaliceCuiPresenter).Output(g, nil)
		// CPU のターンであっても、人間の手札は公開され続ける
		assert.Contains(t, result, "SPADE 5")
		// CPU の手札は依然として非公開
		assert.NotContains(t, result, "DIAMOND 3")
	})

	t.Run("human wins", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("GetCurrent").Return(0).Maybe()
		g.On("GetWinner").Return(domain.SpiteAndMaliceHumanIdx).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		human := domain.NewSpiteAndMalicePlayer(false)
		cpu := domain.NewSpiteAndMalicePlayer(true)
		g.On("GetPlayer", 0).Return(human).Maybe()
		g.On("GetPlayer", 1).Return(cpu).Maybe()
		assert.Contains(t, new(SpiteAndMaliceCuiPresenter).Output(g, nil), "勝ち")
	})

	t.Run("cpu wins", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("GetCurrent").Return(1).Maybe()
		g.On("GetWinner").Return(domain.SpiteAndMaliceCpuIdx).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		human := domain.NewSpiteAndMalicePlayer(false)
		cpu := domain.NewSpiteAndMalicePlayer(true)
		g.On("GetPlayer", 0).Return(human).Maybe()
		g.On("GetPlayer", 1).Return(cpu).Maybe()
		assert.Contains(t, new(SpiteAndMaliceCuiPresenter).Output(g, nil), "CPU")
	})

	t.Run("nil player skipped gracefully", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("GetCurrent").Return(0).Maybe()
		g.On("GetWinner").Return(-1).Maybe()
		g.On("GetStockSize").Return(0).Maybe()
		g.On("GetCompletedSize").Return(0).Maybe()
		var foundations [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		g.On("GetPlayer", 0).Return((*domain.SpiteAndMalicePlayer)(nil)).Maybe()
		g.On("GetPlayer", 1).Return((*domain.SpiteAndMalicePlayer)(nil)).Maybe()
		assert.NotEmpty(t, new(SpiteAndMaliceCuiPresenter).Output(g, nil))
	})
}

func TestSpiteAndMaliceCuiPresenter_HintOutput(t *testing.T) {
	tests := []struct {
		name     string
		hint     *domain.SpiteAndMaliceHint
		contains string
	}{
		{"goal", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceGoal, FoundationIdx: 1}, "ゴール"},
		{"hand", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceHand, Index: 2, FoundationIdx: 0}, "手札2"},
		{"side", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceSide, Index: 3, FoundationIdx: 1}, "サイド3"},
		{"discard", &domain.SpiteAndMaliceHint{Source: domain.SpiteAndMaliceSourceHand, Index: 0, FoundationIdx: 2, Discard: true}, "ディスカード"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := new(interfaces.MockSpiteAndMaliceGame)
			g.On("GetHint").Return(tt.hint)
			assert.Contains(t, new(SpiteAndMaliceCuiPresenter).HintOutput(g), tt.contains)
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetHint").Return((*domain.SpiteAndMaliceHint)(nil))
		assert.Contains(t, new(SpiteAndMaliceCuiPresenter).HintOutput(g), "ヒントはありません")
	})

	t.Run("unknown source falls through", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetHint").Return(&domain.SpiteAndMaliceHint{Source: 99})
		assert.Contains(t, new(SpiteAndMaliceCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestSpiteAndMaliceCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing returns empty", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhasePlaying)
		assert.NotEmpty(t, new(SpiteAndMaliceCuiPresenter).ActionLogOutput(g))
	})
	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSpiteAndMaliceGame)
		g.On("GetPhase").Return(domain.SpiteAndMalicePhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "playHand", Detail: "test"}})
		assert.NotEmpty(t, new(SpiteAndMaliceCuiPresenter).ActionLogOutput(g))
	})
}
