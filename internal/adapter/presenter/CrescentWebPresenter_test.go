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

func setupCrescentWebMockDefaults(cg *interfaces.MockCrescentGame) {
	cg.On("GetPhase").Return(domain.CrescentPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("GetRedealsRemaining").Return(domain.CrescentMaxRedeals).Maybe()
	cg.On("CanUndo").Return(false).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()
	cg.On("HasNoLegalMoves").Return(false).Maybe()
	cg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.CrescentTableauCnt][]*domain.CrescentTableauCard
	for i := range domain.CrescentTableauCnt {
		tableau[i] = make([]*domain.CrescentTableauCard, domain.CrescentTableauInitialSize)
		for j := range domain.CrescentTableauInitialSize {
			tableau[i][j] = &domain.CrescentTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CrescentFoundationCnt][]*domain.Card
	for i := range domain.CrescentAscendingFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(testCrescentFoundationSuit(i), 1, false)}
	}
	for i := domain.CrescentAscendingFoundationCnt; i < domain.CrescentFoundationCnt; i++ {
		foundation[i] = []*domain.Card{domain.NewCard(testCrescentFoundationSuit(i), domain.CardValueMax, false)}
	}
	cg.On("GetFoundation").Return(foundation).Maybe()
	for col := 0; col < domain.CrescentTableauCnt; col++ {
		cg.On("LegalTargets", col).Return([]int{}, []int{}).Maybe()
	}
}

func parseCrescentOutput(t *testing.T, jsonStr string) *controller.CrescentWebOutput {
	t.Helper()
	var out controller.CrescentWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupCrescentOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupCrescentOutputMock(g *interfaces.MockCrescentGame) {
	setupCrescentWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCrescentWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		p := new(CrescentWebPresenter)

		result := parseCrescentOutput(t, p.Output(cg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, domain.CrescentMaxRedeals, result.RedealsRemaining)
		assert.False(t, result.NoLegalMoves)
		assert.Len(t, result.Tableau, domain.CrescentTableauCnt)
		assert.Len(t, result.Foundation, domain.CrescentFoundationCnt)
		assert.Empty(t, result.LegalTargets)
		assert.Equal(t, "crescent.playing", result.MessageCode)
	})

	t.Run("no legal moves with redeals remaining", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "HasNoLegalMoves")
		cg.On("HasNoLegalMoves").Return(true)
		result := parseCrescentOutput(t, new(CrescentWebPresenter).Output(cg, nil))
		assert.True(t, result.NoLegalMoves)
	})

	t.Run("error message", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.Output(cg, errors.New("boom")))
		assert.Equal(t, "boom", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CrescentPhaseGameClear)
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.Output(cg, nil))
		assert.Equal(t, "crescent.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CrescentPhaseGameOver)
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.Output(cg, nil))
		assert.Equal(t, "crescent.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.Output(cg, nil))
		assert.Equal(t, "crescent.stalemate", result.MessageCode)
	})
}

func TestCrescentWebPresenter_LegalTargets(t *testing.T) {
	for _, output := range []struct {
		name string
		call func(*CrescentWebPresenter, interfaces.CrescentGame) string
	}{
		{"output", func(p *CrescentWebPresenter, g interfaces.CrescentGame) string { return p.Output(g, nil) }},
		{"hint output with hint", func(p *CrescentWebPresenter, g interfaces.CrescentGame) string { return p.HintOutput(g) }},
		{"hint output without hint", func(p *CrescentWebPresenter, g interfaces.CrescentGame) string { return p.HintOutput(g) }},
	} {
		t.Run(output.name, func(t *testing.T) {
			cg := new(interfaces.MockCrescentGame)
			setupCrescentWebMockDefaults(cg)
			cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "LegalTargets")
			// testify uses the first matching expectation, so drop the shared
			// GetHint default before each case registers its own.
			cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetHint")
			cg.On("LegalTargets", 0).Return([]int(nil), []int{2}).Maybe()
			cg.On("LegalTargets", 1).Return([]int{4}, []int(nil)).Maybe()
			for col := 2; col < domain.CrescentTableauCnt; col++ {
				cg.On("LegalTargets", col).Return([]int{}, []int{}).Maybe()
			}
			switch output.name {
			case "hint output with hint":
				cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 0, ToZone: "tableau", ToCol: 4})
			default:
				cg.On("GetHint").Return((*domain.CrescentHint)(nil)).Maybe()
			}
			result := parseCrescentOutput(t, output.call(new(CrescentWebPresenter), cg))
			if output.name == "hint output with hint" {
				require.NotNil(t, result.Hint, "the with-hint case must actually carry a hint")
			}
			assert.Equal(t, map[string]controller.CrescentWebOutputTargets{
				"tableau-0": {Tableau: []int{}, Foundation: []int{2}},
				"tableau-1": {Tableau: []int{4}, Foundation: []int{}},
			}, result.LegalTargets)
		})
	}
}

func TestCrescentWebPresenter_LegalTargetsOmitsEmptySourcesAndDestinations(t *testing.T) {
	cg := new(interfaces.MockCrescentGame)
	setupCrescentWebMockDefaults(cg)
	cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "LegalTargets")

	var tableau [domain.CrescentTableauCnt][]*domain.CrescentTableauCard
	tableau[0] = []*domain.CrescentTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 2, false), FaceUp: true}}
	cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetTableau")
	cg.On("GetTableau").Return(tableau)
	// The empty source at column 1 must not be queried into the result. The
	// non-empty source at column 0 has no destinations and must also be omitted.
	cg.On("LegalTargets", 0).Return([]int{}, []int{})
	for col := 2; col < domain.CrescentTableauCnt; col++ {
		cg.On("LegalTargets", col).Return([]int{}, []int{}).Maybe()
	}
	cg.On("GetHint").Return((*domain.CrescentHint)(nil)).Maybe()

	result := parseCrescentOutput(t, new(CrescentWebPresenter).Output(cg, nil))
	assert.Empty(t, result.LegalTargets)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCrescentWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 2, ToZone: "foundation", ToCol: 1, Redeal: false}).Maybe()

		result := new(CrescentWebPresenter).Output(cg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentWebMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 2, ToZone: "foundation", ToCol: 1, Redeal: false}).Maybe()

		result := new(CrescentWebPresenter).Output(cg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestCrescentWebPresenter_HintOutput(t *testing.T) {
	t.Run("with tableau→foundation hint", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 2, ToZone: "foundation", ToCol: 0})
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.HintOutput(cg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 2, result.Hint.FromCol)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "crescent.hintAvailable", result.MessageCode)
	})

	t.Run("with redeal hint", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: -1, ToCol: -1, Redeal: true})
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.HintOutput(cg))
		assert.NotNil(t, result.Hint)
		assert.True(t, result.Hint.Redeal)
	})

	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentWebMockDefaults(cg)
		cg.On("GetHint").Return((*domain.CrescentHint)(nil))
		p := new(CrescentWebPresenter)
		result := parseCrescentOutput(t, p.HintOutput(cg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "crescent.noHint", result.MessageCode)
	})
}

func TestCrescentWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetPhase").Return(domain.CrescentPhasePlaying)
		cg.On("GetGameEndFlag").Return(false)
		p := new(CrescentWebPresenter)
		assert.Contains(t, p.ActionLogOutput(cg), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetPhase").Return(domain.CrescentPhaseGameOver)
		cg.On("GetGameEndFlag").Return(true)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "redeal", DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"}},
		})
		p := new(CrescentWebPresenter)
		assert.Contains(t, p.ActionLogOutput(cg), "redeal")
	})
}
