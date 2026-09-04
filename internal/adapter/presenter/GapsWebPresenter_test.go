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

func setupGapsWebMockDefaults(g *interfaces.MockGapsGame) {
	g.On("GetPhase").Return(domain.GapsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetRedealsUsed").Return(0).Maybe()
	g.On("GetRedealsRemaining").Return(3).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			grid[r][c] = domain.NewCard(s, c+2, true)
		}
	}
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
}

func parseGapsOutput(t *testing.T, s string) *controller.GapsWebOutput {
	t.Helper()
	var out controller.GapsWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// setupGapsOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupGapsOutputMock(g *interfaces.MockGapsGame) {
	setupGapsWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestGapsWebPresenter_Output_Playing(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsOutputMock(g)
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, nil))
	assert.Equal(t, "gaps.playing", out.MessageCode)
	assert.Equal(t, 3, out.RedealsRemaining)
	assert.Len(t, out.Grid, domain.GapsRowCnt)
}

func TestGapsWebPresenter_Output_Stalemate(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsOutputMock(g)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.GapsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetRedealsUsed").Return(3).Maybe()
	g.On("GetRedealsRemaining").Return(0).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("UndoToEscape").Return(2).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, nil))
	assert.Equal(t, "gaps.stalemate", out.MessageCode)
}

func TestGapsWebPresenter_Output_GameClear(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsOutputMock(g)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.GapsPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(42).Maybe()
	g.On("GetRedealsUsed").Return(1).Maybe()
	g.On("GetRedealsRemaining").Return(2).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, nil))
	assert.Equal(t, "gaps.gameClear", out.MessageCode)
	assert.Equal(t, "42", out.MessageParams["moveCount"])
}

func TestGapsWebPresenter_Output_GameOver(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsOutputMock(g)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.GapsPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetRedealsUsed").Return(3).Maybe()
	g.On("GetRedealsRemaining").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("UndoToEscape").Return(-1).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, nil))
	assert.Equal(t, "gaps.gameOver", out.MessageCode)
}

func TestGapsWebPresenter_Output_Error(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsOutputMock(g)
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, errors.New("oops")))
	assert.Equal(t, "oops", out.Message)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGapsWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.GapsHint{FromRow: 1, FromCol: 2, ToRow: 0, ToCol: 3}

	g := new(interfaces.MockGapsGame)
	setupGapsWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := new(GapsWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

func TestGapsWebPresenter_HintOutput_None(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsWebMockDefaults(g)
	g.On("GetHint").Return((*domain.GapsHint)(nil))
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.HintOutput(g))
	assert.Nil(t, out.Hint)
	assert.Equal(t, "gaps.noHint", out.MessageCode)
}

func TestGapsWebPresenter_HintOutput_WithHint(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	setupGapsWebMockDefaults(g)
	g.On("GetHint").Return(&domain.GapsHint{FromRow: 1, FromCol: 0, ToRow: 0, ToCol: 0})
	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.HintOutput(g))
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "gaps.hintAvailable", out.MessageCode)
}

func TestGapsWebPresenter_Output_GapNeeds(t *testing.T) {
	g := new(interfaces.MockGapsGame)
	g.On("GetPhase").Return(domain.GapsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetRedealsUsed").Return(0).Maybe()
	g.On("GetRedealsRemaining").Return(3).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetHint").Return(nil).Maybe()

	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	grid[0][3] = domain.NewCard(domain.CardDesignSpade, 10, true)
	g.On("GetGrid").Return(grid).Maybe()

	// 1. anySuit (col 0, rank 2)
	g.On("GetGapNeed", 0, 0).Return(&domain.GapsGapNeed{
		Kind:  domain.GapsNeedAnySuit,
		Value: domain.GapsAnchorRank,
	}).Maybe()

	// 2. blocked (left is K)
	g.On("GetGapNeed", 0, 1).Return(&domain.GapsGapNeed{
		Kind: domain.GapsNeedBlocked,
	}).Maybe()

	// 3. card / needed
	g.On("GetGapNeed", 1, 2).Return(&domain.GapsGapNeed{
		Kind:   domain.GapsNeedCard,
		Design: domain.CardDesignSpade,
		Value:  5,
	}).Maybe()

	// Catch-all for remaining cells (occupied, undecided, etc.)
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()

	p := &GapsWebPresenter{}
	out := parseGapsOutput(t, p.Output(g, nil))

	assert.Len(t, out.GapNeeds, domain.GapsRowCnt)
	for r := 0; r < domain.GapsRowCnt; r++ {
		assert.Len(t, out.GapNeeds[r], domain.GapsColCnt)
	}

	// 1. anySuit
	assert.NotNil(t, out.GapNeeds[0][0])
	assert.Equal(t, domain.GapsNeedAnySuit, out.GapNeeds[0][0].Kind)
	assert.Equal(t, domain.GapsAnchorRank, out.GapNeeds[0][0].Value)
	assert.Empty(t, out.GapNeeds[0][0].Design)

	// 2. blocked
	assert.NotNil(t, out.GapNeeds[0][1])
	assert.Equal(t, domain.GapsNeedBlocked, out.GapNeeds[0][1].Kind)
	assert.Zero(t, out.GapNeeds[0][1].Value)
	assert.Empty(t, out.GapNeeds[0][1].Design)

	// 3. needed card
	assert.NotNil(t, out.GapNeeds[1][2])
	assert.Equal(t, domain.GapsNeedCard, out.GapNeeds[1][2].Kind)
	assert.Equal(t, "SPADE", out.GapNeeds[1][2].Design)
	assert.Equal(t, 5, out.GapNeeds[1][2].Value)

	// 4. occupied cell is null
	assert.Nil(t, out.GapNeeds[0][3])

	// 5. undecided cell is null
	assert.Nil(t, out.GapNeeds[0][4])
}

func TestGapsWebPresenter_gapNeedToOutput_Nil(t *testing.T) {
	assert.Nil(t, gapNeedToOutput(nil))
}
