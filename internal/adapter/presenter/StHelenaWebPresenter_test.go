//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupStHelenaWebMockDefaults(cg *interfaces.MockStHelenaGame) {
	cg.On("GetPhase").Return(domain.StHelenaPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("GetRedealsRemaining").Return(domain.StHelenaMaxRedeals).Maybe()
	cg.On("RestrictionsActive").Return(true).Maybe()
	cg.On("CanUndo").Return(false).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()
	cg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	for i := range domain.StHelenaTableauCnt {
		tableau[i] = make([]*domain.StHelenaTableauCard, domain.StHelenaTableauInitialSize)
		for j := range domain.StHelenaTableauInitialSize {
			tableau[i][j] = &domain.StHelenaTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.StHelenaFoundationCnt][]*domain.Card
	for i := range domain.StHelenaAscendingFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.StHelenaFoundationSuit(i), 1, false)}
	}
	for i := domain.StHelenaAscendingFoundationCnt; i < domain.StHelenaFoundationCnt; i++ {
		foundation[i] = []*domain.Card{domain.NewCard(domain.StHelenaFoundationSuit(i), domain.CardValueMax, false)}
	}
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func parseStHelenaOutput(t *testing.T, jsonStr string) *controller.StHelenaWebOutput {
	t.Helper()
	var out controller.StHelenaWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupStHelenaOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupStHelenaOutputMock(g *interfaces.MockStHelenaGame) {
	setupStHelenaWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestStHelenaWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaOutputMock(cg)
		p := new(StHelenaWebPresenter)

		result := parseStHelenaOutput(t, p.Output(cg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, domain.StHelenaMaxRedeals, result.RedealsRemaining)
		assert.Len(t, result.Tableau, domain.StHelenaTableauCnt)
		assert.Len(t, result.Foundation, domain.StHelenaFoundationCnt)
		assert.Equal(t, "sthelena.playing", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaOutputMock(cg)
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.Output(cg, errors.New("boom")))
		assert.Equal(t, "boom", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameClear)
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.Output(cg, nil))
		assert.Equal(t, "sthelena.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameOver)
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.Output(cg, nil))
		assert.Equal(t, "sthelena.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.Output(cg, nil))
		assert.Equal(t, "sthelena.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestStHelenaWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: 2, ToZone: "foundation", ToCol: 1, Redeal: false}).Maybe()

		result := new(StHelenaWebPresenter).Output(cg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: 2, ToZone: "foundation", ToCol: 1, Redeal: false}).Maybe()

		result := new(StHelenaWebPresenter).Output(cg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestStHelenaWebPresenter_HintOutput(t *testing.T) {
	t.Run("with tableau→foundation hint", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: 2, ToZone: "foundation", ToCol: 0})
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.HintOutput(cg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 2, result.Hint.FromCol)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "sthelena.hintAvailable", result.MessageCode)
	})

	t.Run("with redeal hint", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: -1, ToCol: -1, Redeal: true})
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.HintOutput(cg))
		assert.NotNil(t, result.Hint)
		assert.True(t, result.Hint.Redeal)
	})

	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.On("GetHint").Return((*domain.StHelenaHint)(nil))
		p := new(StHelenaWebPresenter)
		result := parseStHelenaOutput(t, p.HintOutput(cg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "sthelena.noHint", result.MessageCode)
	})
}

func TestStHelenaWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetPhase").Return(domain.StHelenaPhasePlaying)
		cg.On("GetGameEndFlag").Return(false)
		p := new(StHelenaWebPresenter)
		assert.Contains(t, p.ActionLogOutput(cg), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameOver)
		cg.On("GetGameEndFlag").Return(true)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "redeal", Detail: "test"},
		})
		p := new(StHelenaWebPresenter)
		assert.Contains(t, p.ActionLogOutput(cg), "redeal")
	})
}

// **送り先の制限を API に載せる。**載せないと、ページは「どの組札が押せるか」を
// 決められず、押した瞬間にサーバが拒む組札を並べることになる。
func TestStHelenaWebPresenter_CarriesTheFirstDealRestriction(t *testing.T) {
	for _, active := range []bool{true, false} {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaWebMockDefaults(cg)
		cg.On("GetHint").Return(nil).Maybe()
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "RestrictionsActive")
		cg.On("RestrictionsActive").Return(active)

		var out controller.StHelenaWebOutput
		require.NoError(t, json.Unmarshal([]byte(new(StHelenaWebPresenter).Output(cg, nil)), &out))
		assert.Equal(t, active, out.RestrictionsActive)
	}
}
