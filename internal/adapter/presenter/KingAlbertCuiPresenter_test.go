//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKingAlbertCuiMockDefaults(bg *interfaces.MockKingAlbertGame) {
	bg.On("GetPhase").Return(domain.KingAlbertPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	for i := range domain.KingAlbertTableauCnt {
		tableau[i] = make([]*domain.KingAlbertTableauCard, i+1)
		for j := range i + 1 {
			tableau[i][j] = &domain.KingAlbertTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	reserve := make([]*domain.Card, domain.KingAlbertReserveCnt)
	for i := range reserve {
		reserve[i] = domain.NewCard(domain.CardDesignHeart, i+2, false)
	}
	bg.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.KingAlbertFoundationCnt][]*domain.Card
	for i := range domain.KingAlbertFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestKingAlbertCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		p := new(KingAlbertCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "King Albert")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		p := new(KingAlbertCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameClear)

		p := new(KingAlbertCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameOver)

		p := new(KingAlbertCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate shows the undo-to-escape guidance", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(3)

		p := new(KingAlbertCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
		assert.Contains(t, result, "脱出には undo を 3 回")
	})

	t.Run("empty column and depleted reserve", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		bg.On("GetTableau").Return(emptyTableau)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetReserve")
		bg.On("GetReserve").Return([]*domain.Card{nil, nil})

		p := new(KingAlbertCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty foundation", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var emptyFoundation [domain.KingAlbertFoundationCnt][]*domain.Card
		bg.On("GetFoundation").Return(emptyFoundation)

		p := new(KingAlbertCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestKingAlbertCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint from tableau", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetHint").Return(&domain.KingAlbertHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(KingAlbertCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetHint").Return(&domain.KingAlbertHint{
			FromZone:  "tableau",
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(KingAlbertCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("reserve hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetHint").Return(&domain.KingAlbertHint{
			FromZone:  "reserve",
			FromCol:   2,
			CardIndex: -1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(KingAlbertCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "リザーブ2")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetHint").Return((*domain.KingAlbertHint)(nil))

		p := new(KingAlbertCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestKingAlbertCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetPhase").Return(domain.KingAlbertPhasePlaying)

		p := new(KingAlbertCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockKingAlbertGame)
		bg.On("GetPhase").Return(domain.KingAlbertPhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(KingAlbertCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
