//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBigBenWebMockDefaults(g *interfaces.MockBigBenGame) {
	g.On("GetPhase").Return(domain.BigBenPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(52).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(false).Maybe()

	var tableau [domain.BigBenTableauCnt][]*domain.BigBenTableauCard
	for i := range domain.BigBenTableauCnt {
		tableau[i] = make([]*domain.BigBenTableauCard, domain.BigBenColumnLen)
		for j := range domain.BigBenColumnLen {
			tableau[i][j] = &domain.BigBenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BigBenFoundationCnt][]*domain.Card
	for i := range domain.BigBenFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseBigBenOutput(t *testing.T, jsonStr string) *controller.BigBenWebOutput {
	t.Helper()
	var out controller.BigBenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupBigBenOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupBigBenOutputMock(g *interfaces.MockBigBenGame) {
	setupBigBenWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestBigBenWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Len(t, result.Tableau, domain.BigBenTableauCnt)
		assert.Len(t, result.Foundation, domain.BigBenFoundationCnt)
		assert.Equal(t, "bigben.playing", result.MessageCode)
	})

	// The target rank is on the wire so the client never has to recompute the
	// clock ordering and drift from the domain.
	t.Run("each face carries its target rank", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
		for i, f := range result.Foundation {
			assert.Equal(t, domain.BigBenTargetRank(i), f.TargetRank, "face %d", i)
			assert.False(t, f.Complete)
		}
		// **9 時始まり。**クローン元は 1 時始まりで添字 0 が A を求めた。
		assert.Equal(t, 9, result.Foundation[0].TargetRank, "index 0 is 9 o'clock")
		assert.Equal(t, 8, result.Foundation[11].TargetRank, "index 11 is 8 o'clock")
	})

	t.Run("completed faces are flagged", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(true)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
		for _, f := range result.Foundation {
			assert.True(t, f.Complete)
		}
	})

	t.Run("all face up", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.BigBenPhase
		code string
	}{
		{"game clear", domain.BigBenPhaseGameClear, "bigben.gameClear"},
		{"game over", domain.BigBenPhaseGameOver, "bigben.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBigBenGame)
			setupBigBenOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
		assert.Equal(t, "bigben.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBigBenWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.BigBenHint{FromCol: 2, ToZone: "foundation", ToIdx: 0}

	g := new(interfaces.MockBigBenGame)
	setupBigBenWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestBigBenWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenWebMockDefaults(g)
		g.On("GetHint").Return(&domain.BigBenHint{FromCol: 3, ToZone: "foundation", ToIdx: 7})

		result := parseBigBenOutput(t, new(BigBenWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 3, result.Hint.FromCol)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, 7, result.Hint.ToIdx)
		assert.Equal(t, "bigben.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenWebMockDefaults(g)
		g.On("GetHint").Return((*domain.BigBenHint)(nil))

		result := parseBigBenOutput(t, new(BigBenWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "bigben.noHint", result.MessageCode)
	})
}

func TestBigBenWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		g.On("GetPhase").Return(domain.BigBenPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(BigBenWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		g.On("GetPhase").Return(domain.BigBenPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(BigBenWebPresenter).ActionLogOutput(g), "move")
	})
}

// **山札の残りを API に載せる。**載せないと、ページは空の山札でも補充ボタンを
// 押せる状態で描き、サーバに拒まれるまで気付けない。
func TestBigBenWebPresenter_CarriesTheStockCount(t *testing.T) {
	g := new(interfaces.MockBigBenGame)
	setupBigBenOutputMock(g)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetStockCount")
	g.On("GetStockCount").Return(37)

	assert.Equal(t, 37, parseBigBenOutput(t, new(BigBenWebPresenter).Output(g, nil)).StockCount)
}
