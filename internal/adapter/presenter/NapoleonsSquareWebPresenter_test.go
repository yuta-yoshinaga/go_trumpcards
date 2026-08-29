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

func setupNapoleonsSquareWebMockDefaults(g *interfaces.MockNapoleonsSquareGame) {
	g.On("GetPhase").Return(domain.NapoleonsSquarePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)}).Maybe()

	var tableau [domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard
	for i := range domain.NapoleonsSquareTableauCnt {
		tableau[i] = make([]*domain.NapoleonsSquareTableauCard, domain.NapoleonsSquareColumnLen)
		for j := range domain.NapoleonsSquareColumnLen {
			tableau[i][j] = &domain.NapoleonsSquareTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.NapoleonsSquareFoundationCnt][]*domain.Card
	for i := range domain.NapoleonsSquareFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i%4, 1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseNapoleonsSquareOutput(t *testing.T, jsonStr string) *controller.NapoleonsSquareWebOutput {
	t.Helper()
	var out controller.NapoleonsSquareWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupNapoleonsSquareOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupNapoleonsSquareOutputMock(g *interfaces.MockNapoleonsSquareGame) {
	setupNapoleonsSquareWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestNapoleonsSquareWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareOutputMock(g)

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 48, result.StockCount)
		assert.Len(t, result.Waste, 1)
		assert.Len(t, result.Tableau, domain.NapoleonsSquareTableauCnt)
		assert.Len(t, result.Foundation, domain.NapoleonsSquareFoundationCnt)
		assert.Equal(t, "napoleonssquare.playing", result.MessageCode)
	})

	t.Run("all face up", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareOutputMock(g)

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareOutputMock(g)

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	phases := []struct {
		name string
		val  domain.NapoleonsSquarePhase
		code string
	}{
		{"game clear", domain.NapoleonsSquarePhaseGameClear, "napoleonssquare.gameClear"},
		{"game over", domain.NapoleonsSquarePhaseGameOver, "napoleonssquare.gameOver"},
	}
	for _, tc := range phases {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockNapoleonsSquareGame)
			setupNapoleonsSquareOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, nil))
		assert.Equal(t, "napoleonssquare.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestNapoleonsSquareWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		nsg := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareWebMockDefaults(nsg)
		nsg.On("GetHint").Return(&domain.NapoleonsSquareHint{FromZone: "tableau", FromCol: 1, CardIndex: 0, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(NapoleonsSquareWebPresenter).Output(nsg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		nsg := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareWebMockDefaults(nsg)
		nsg.ExpectedCalls = filterCalls(nsg.ExpectedCalls, "IsStalemate")
		nsg.On("IsStalemate").Return(true)
		nsg.On("GetHint").Return(&domain.NapoleonsSquareHint{FromZone: "tableau", FromCol: 1, CardIndex: 0, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(NapoleonsSquareWebPresenter).Output(nsg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestNapoleonsSquareWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareWebMockDefaults(g)
		g.On("GetHint").Return(&domain.NapoleonsSquareHint{
			FromZone: "tableau", FromCol: 3, CardIndex: 1, ToZone: "tableau", ToCol: 7,
		})

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 3, result.Hint.FromCol)
		assert.Equal(t, 1, result.Hint.CardIndex, "the run head must survive the wire")
		assert.Equal(t, 7, result.Hint.ToCol)
		assert.Equal(t, "napoleonssquare.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareWebMockDefaults(g)
		g.On("GetHint").Return((*domain.NapoleonsSquareHint)(nil))

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "napoleonssquare.noHint", result.MessageCode)
	})
}

func TestNapoleonsSquareWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		g.On("GetPhase").Return(domain.NapoleonsSquarePhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(NapoleonsSquareWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		g.On("GetPhase").Return(domain.NapoleonsSquarePhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(NapoleonsSquareWebPresenter).ActionLogOutput(g), "move")
	})
}

func TestNapoleonsSquareWebPresenter_ErrorCode(t *testing.T) {
	t.Run("code is separated from message", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareOutputMock(g)
		err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "napoleonssquare.errBadColumn", nil)

		result := parseNapoleonsSquareOutput(t, new(NapoleonsSquareWebPresenter).Output(g, err))
		assert.Equal(t, "napoleonssquare.errBadColumn", result.MessageCode, "コードがあれば MessageCode に載る")
		assert.Empty(t, result.Message, "コードがあるなら Message を汚さない (#5553)")
	})
}
