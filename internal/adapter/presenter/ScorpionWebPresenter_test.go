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

func setupScorpionWebMockDefaults(sg *interfaces.MockScorpionGame) {
	sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("CanUndo").Return(false).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()
	sg.On("GetStockCount").Return(3).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetCompletedSuitMask").Return(0).Maybe()

	var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.ScorpionTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	sg.On("GetTableau").Return(tableau).Maybe()
	sg.On("LegalTargetsFrom", mock.Anything, mock.Anything).Return([]int(nil), []int(nil)).Maybe()
}

func parseScorpionOutput(t *testing.T, jsonStr string) *controller.ScorpionWebOutput {
	t.Helper()
	var out controller.ScorpionWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupScorpionOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupScorpionOutputMock(g *interfaces.MockScorpionGame) {
	setupScorpionWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestScorpionWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionOutputMock(sg)
		p := new(ScorpionWebPresenter)

		result := parseScorpionOutput(t, p.Output(sg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 3, result.StockCount)
		assert.Equal(t, "scorpion.playing", result.MessageCode)
		assert.Empty(t, result.LegalTargets)
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		sg.On("GetMoveCount").Return(5).Maybe()
		sg.On("CanUndo").Return(true).Maybe()
		sg.On("IsStalemate").Return(true).Maybe()
		sg.On("UndoToEscape").Return(3).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetCompletedSuitMask").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionWebPresenter)
		result := parseScorpionOutput(t, p.Output(sg, nil))
		assert.Equal(t, "scorpion.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameClear).Maybe()
		sg.On("GetMoveCount").Return(42).Maybe()
		sg.On("CanUndo").Return(false).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(4).Maybe()
		sg.On("GetCompletedSuitMask").Return(15).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionWebPresenter)
		result := parseScorpionOutput(t, p.Output(sg, nil))
		assert.Equal(t, "scorpion.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver).Maybe()
		sg.On("GetMoveCount").Return(10).Maybe()
		sg.On("CanUndo").Return(false).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetCompletedSuitMask").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionWebPresenter)
		result := parseScorpionOutput(t, p.Output(sg, nil))
		assert.Equal(t, "scorpion.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionOutputMock(sg)
		p := new(ScorpionWebPresenter)

		result := parseScorpionOutput(t, p.Output(sg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

func TestScorpionWebPresenter_LegalMovesOutput(t *testing.T) {
	sg := new(interfaces.MockScorpionGame)
	setupScorpionOutputMock(sg)
	p := new(ScorpionWebPresenter)

	// Web delegates legal-move previews to the normal state JSON (targets are
	// computed client-side), so the output mirrors Output.
	result := parseScorpionOutput(t, p.LegalMovesOutput(sg, 0))
	assert.Equal(t, "scorpion.playing", result.MessageCode)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestScorpionWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 5}).Maybe()

		result := new(ScorpionWebPresenter).Output(sg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionWebMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 5}).Maybe()

		result := new(ScorpionWebPresenter).Output(sg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestScorpionWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		p := new(ScorpionWebPresenter)
		result := parseScorpionOutput(t, p.HintOutput(sg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 3, result.Hint.ToCol)
		assert.Equal(t, "scorpion.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionWebMockDefaults(sg)
		sg.On("GetHint").Return((*domain.ScorpionHint)(nil))

		p := new(ScorpionWebPresenter)
		result := parseScorpionOutput(t, p.HintOutput(sg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "scorpion.noHint", result.MessageCode)
	})
}

func TestScorpionWebPresenter_LegalTargetsOnEveryOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		output func(*ScorpionWebPresenter, interfaces.ScorpionGame) string
	}{
		{"output", func(p *ScorpionWebPresenter, g interfaces.ScorpionGame) string { return p.Output(g, nil) }},
		{"hint output without hint", func(p *ScorpionWebPresenter, g interfaces.ScorpionGame) string { return p.HintOutput(g) }},
		{"legal moves output", func(p *ScorpionWebPresenter, g interfaces.ScorpionGame) string { return p.LegalMovesOutput(g, 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sg := new(interfaces.MockScorpionGame)
			setupScorpionWebMockDefaults(sg)
			sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "LegalTargetsFrom")
			// Face-down cards are not valid sources and must be skipped before
			// asking the domain for their destinations.
			var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
			tableau[0] = []*domain.KlondikeTableauCard{
				{Card: domain.NewCard(domain.CardDesignHeart, 5, false), FaceUp: true},
				{Card: domain.NewCard(domain.CardDesignSpade, 6, false), FaceUp: true},
				{Card: domain.NewCard(domain.CardDesignClover, 8, false), FaceUp: true},
				{Card: domain.NewCard(domain.CardDesignDiamond, 9, false), FaceUp: false},
			}
			sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
			sg.On("GetTableau").Return(tableau).Maybe()
			sg.On("LegalTargetsFrom", 0, 0).Return([]int(nil), []int{2}).Maybe()
			sg.On("LegalTargetsFrom", 0, 1).Return([]int{3}, []int(nil)).Maybe()
			sg.On("LegalTargetsFrom", 0, 2).Return([]int(nil), []int(nil)).Maybe()
			sg.On("GetHint").Return((*domain.ScorpionHint)(nil)).Maybe()
			out := parseScorpionOutput(t, tc.output(new(ScorpionWebPresenter), sg))
			assert.Equal(t, controller.ScorpionWebOutputTargets{Tableau: []int{}, Foundation: []int{2}}, out.LegalTargets["tableau-0-0"])
			assert.Equal(t, controller.ScorpionWebOutputTargets{Tableau: []int{3}, Foundation: []int{}}, out.LegalTargets["tableau-0-1"])
			assert.NotContains(t, out.LegalTargets, "tableau-0-2")
			assert.NotContains(t, out.LegalTargets, "tableau-0-3")
			assert.NotNil(t, out.LegalTargets)
		})
	}
}

func TestScorpionWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying)

		sg.On("GetGameEndFlag").Return(false)
		p := new(ScorpionWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})

		p := new(ScorpionWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "entries")
	})
}
