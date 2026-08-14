//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupRoyalCotillionWebMockDefaults(g *interfaces.MockRoyalCotillionGame) {
	g.On("GetPhase").Return(domain.RoyalCotillionPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(76).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.RoyalCotillionTableauCnt]*domain.Card
	for i := range domain.RoyalCotillionTableauCnt {
		tableau[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var reserve [domain.RoyalCotillionReserveCnt][]*domain.Card
	for i := range domain.RoyalCotillionReserveCnt {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+2, true)}
	}
	g.On("GetReserve").Return(reserve).Maybe()

	for i := range domain.RoyalCotillionFoundationCnt {
		g.On("IsOddFoundation", i).Return(i < domain.RoyalCotillionOddCnt).Maybe()
	}

	var foundation [domain.RoyalCotillionFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseRoyalCotillionOutput(t *testing.T, jsonStr string) *controller.RoyalCotillionWebOutput {
	t.Helper()
	var out controller.RoyalCotillionWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupRoyalCotillionOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupRoyalCotillionOutputMock(g *interfaces.MockRoyalCotillionGame) {
	setupRoyalCotillionWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestRoyalCotillionWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionOutputMock(g)

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 76, result.StockCount)
		// 1 枠 1 枚なので Tableau はカードの配列（山の配列ではない）。
		assert.Len(t, result.Tableau, domain.RoyalCotillionTableauCnt)
		assert.Len(t, result.Reserve, domain.RoyalCotillionReserveCnt)
		// A 始まり / 2 始まりはワイヤに載せる。添字から推測させない。
		assert.Equal(t,
			[]bool{true, true, true, true, false, false, false, false},
			result.FoundationOdd)
		assert.Len(t, result.Tableau, domain.RoyalCotillionTableauCnt)
		assert.Len(t, result.Foundation, domain.RoyalCotillionFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "royalcotillion.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).Output(g, nil))
		assert.Equal(t, "royalcotillion.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionOutputMock(g)

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.RoyalCotillionPhase
		code string
	}{
		{"game clear", domain.RoyalCotillionPhaseGameClear, "royalcotillion.gameClear"},
		{"game over", domain.RoyalCotillionPhaseGameOver, "royalcotillion.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockRoyalCotillionGame)
			setupRoyalCotillionOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestRoyalCotillionWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.RoyalCotillionHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestRoyalCotillionWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionWebMockDefaults(g)
		g.On("GetHint").Return(&domain.RoyalCotillionHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "royalcotillion.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionWebMockDefaults(g)
		g.On("GetHint").Return((*domain.RoyalCotillionHint)(nil))

		result := parseRoyalCotillionOutput(t, new(RoyalCotillionWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "royalcotillion.noHint", result.MessageCode)
	})
}

func TestRoyalCotillionWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		g.On("GetPhase").Return(domain.RoyalCotillionPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(RoyalCotillionWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		g.On("GetPhase").Return(domain.RoyalCotillionPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(RoyalCotillionWebPresenter).ActionLogOutput(g), "move")
	})
}
