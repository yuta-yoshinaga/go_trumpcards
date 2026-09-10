//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func setupMatrimonyWebMockDefaults(g *mockMatrimonyGame) {
	g.On("GetPhase").Return(domain.MatrimonyPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(88).Maybe()
	g.On("GetRedealCount").Return(0).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.MatrimonyTableauCnt]*domain.Card
	for i := range domain.MatrimonyTableauCnt {
		tableau[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.MatrimonyFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseMatrimonyOutput(t *testing.T, jsonStr string) *controller.MatrimonyWebOutput {
	t.Helper()
	var out controller.MatrimonyWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupMatrimonyOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupMatrimonyOutputMock(g *mockMatrimonyGame) {
	setupMatrimonyWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestMatrimonyWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyOutputMock(g)

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 88, result.StockCount)
		// 1 枠 1 枚なので Tableau はカードの配列（山の配列ではない）。
		assert.Len(t, result.Tableau, domain.MatrimonyTableauCnt)
		assert.Len(t, result.Tableau, domain.MatrimonyTableauCnt)
		// A 始まり / 2 始まりはワイヤに載せる。添字から推測させない。
		assert.Len(t, result.Tableau, domain.MatrimonyTableauCnt)
		assert.Len(t, result.Foundation, domain.MatrimonyFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "matrimony.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).Output(g, nil))
		assert.Equal(t, "matrimony.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyOutputMock(g)

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.MatrimonyPhase
		code string
	}{
		{"game clear", domain.MatrimonyPhaseGameClear, "matrimony.gameClear"},
		{"game over", domain.MatrimonyPhaseGameOver, "matrimony.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(mockMatrimonyGame)
			setupMatrimonyOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestMatrimonyWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.MatrimonyHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestMatrimonyWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyWebMockDefaults(g)
		g.On("GetHint").Return(&domain.MatrimonyHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "matrimony.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyWebMockDefaults(g)
		g.On("GetHint").Return((*domain.MatrimonyHint)(nil))

		result := parseMatrimonyOutput(t, new(MatrimonyWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "matrimony.noHint", result.MessageCode)
	})
}

func TestMatrimonyWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		g.On("GetPhase").Return(domain.MatrimonyPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(MatrimonyWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		g.On("GetPhase").Return(domain.MatrimonyPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(MatrimonyWebPresenter).ActionLogOutput(g), "move")
	})
}
