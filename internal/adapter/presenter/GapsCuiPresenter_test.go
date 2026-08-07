//go:build test

package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupGapsCuiMock() *interfaces.MockGapsGame {
	g := new(interfaces.MockGapsGame)
	g.On("GetPhase").Return(domain.GapsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetRedealsUsed").Return(0).Maybe()
	g.On("GetRedealsRemaining").Return(3).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CanUndo").Return(false).Maybe()
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
	g.On("GetLockedPrefixLengths").Return([domain.GapsRowCnt]int{3, 0, 0, 0}).Maybe()
	g.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	return g
}

func TestGapsCuiPresenter_Output_Playing(t *testing.T) {
	g := setupGapsCuiMock()
	p := &GapsCuiPresenter{}
	out := p.Output(g, nil)
	assert.Contains(t, out, "Gaps")
	assert.Contains(t, out, "再配り 0/3")
	assert.Contains(t, out, "[ . ]")
	// Locked legend plus 3 markers on row 0 (one per locked card) → 4 asterisks.
	assert.Contains(t, out, i18n.T("gaps.lockedLegend"))
	assert.Equal(t, 4, strings.Count(out, "*"))
}

func TestGapsCuiPresenter_Output_GameClear(t *testing.T) {
	g := setupGapsCuiMock()
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.GapsPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(10).Maybe()
	g.On("GetRedealsUsed").Return(1).Maybe()
	g.On("GetRedealsRemaining").Return(2).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
	g.On("GetLockedPrefixLengths").Return([domain.GapsRowCnt]int{}).Maybe()
	g.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	p := &GapsCuiPresenter{}
	out := p.Output(g, nil)
	assert.Contains(t, out, "ゲームクリア")
}

func TestGapsCuiPresenter_Output_GameOver(t *testing.T) {
	g := setupGapsCuiMock()
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.GapsPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetRedealsUsed").Return(3).Maybe()
	g.On("GetRedealsRemaining").Return(0).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("UndoToEscape").Return(-1).Maybe()
	var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	g.On("GetGrid").Return(grid).Maybe()
	g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
	g.On("GetLockedPrefixLengths").Return([domain.GapsRowCnt]int{}).Maybe()
	g.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	p := &GapsCuiPresenter{}
	out := p.Output(g, nil)
	assert.Contains(t, out, "ゲームオーバー")
}

func TestGapsCuiPresenter_Output_WithError(t *testing.T) {
	g := setupGapsCuiMock()
	p := &GapsCuiPresenter{}
	out := p.Output(g, errors.New("oops"))
	assert.Contains(t, out, "oops")
}

func TestGapsCuiPresenter_HintOutput_None(t *testing.T) {
	g := setupGapsCuiMock()
	g.On("GetHint").Return((*domain.GapsHint)(nil))
	p := &GapsCuiPresenter{}
	out := p.HintOutput(g)
	assert.NotEmpty(t, out)
}

func TestGapsCuiPresenter_HintOutput_Move(t *testing.T) {
	g := setupGapsCuiMock()
	g.On("GetHint").Return(&domain.GapsHint{FromRow: 1, FromCol: 0, ToRow: 0, ToCol: 0})
	p := &GapsCuiPresenter{}
	out := p.HintOutput(g)
	assert.Contains(t, out, "(1,0)")
	assert.Contains(t, out, "(0,0)")
}

func TestGapsCuiPresenter_ActionLogOutput(t *testing.T) {
	g := setupGapsCuiMock()
	p := &GapsCuiPresenter{}
	// Phase==Playing returns empty action log
	out := p.ActionLogOutput(g)
	assert.NotNil(t, out)
}

// **どのカードが入るかも詰みかも CUI から分からなかった (#4800)。**Web は
// ゴーストカードと 🚫 で常時プレビューしている。
func TestGapsCuiPresenter_GapNeeds(t *testing.T) {
	p := new(GapsCuiPresenter)
	withNeed := func(need *domain.GapsGapNeed) *interfaces.MockGapsGame {
		g := new(interfaces.MockGapsGame)
		var grid [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
		g.On("GetGrid").Return(grid).Maybe()
		g.On("GetGapNeed", 0, 0).Return(need).Maybe()
		g.On("GetGapNeed", mock.Anything, mock.Anything).Return((*domain.GapsGapNeed)(nil)).Maybe()
		g.On("GetLockedPrefixLengths").Return([domain.GapsRowCnt]int{}).Maybe()
		g.On("GetPhase").Return(domain.GapsPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("GetRedealsLeft").Return(2).Maybe()
		g.On("GetRedealsUsed").Return(0).Maybe()
		g.On("GetMaxRedeals").Return(2).Maybe()
		g.On("GetRedealsRemaining").Return(2).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
		return g
	}

	t.Run("names the exact card a gap needs", func(t *testing.T) {
		out := p.Output(withNeed(&domain.GapsGapNeed{
			Kind: domain.GapsNeedCard, Design: domain.CardDesignHeart, Value: 6,
		}), nil)
		assert.Contains(t, out, "HEART 6")
	})

	t.Run("marks a blocked gap differently from an open one", func(t *testing.T) {
		blocked := p.Output(withNeed(&domain.GapsGapNeed{Kind: domain.GapsNeedBlocked}), nil)
		anySuit := p.Output(withNeed(&domain.GapsGapNeed{Kind: domain.GapsNeedAnySuit, Value: 2}), nil)
		assert.NotEqual(t, blocked, anySuit, "詰みと空きが同じ見た目になっている")
		assert.Contains(t, blocked, "[ x ]")
	})

	// **決まらないマスは従来どおり。**左隣も空きのときは何が入るか決まらない。
	t.Run("leaves an undetermined gap as the plain placeholder", func(t *testing.T) {
		assert.Contains(t, p.Output(withNeed(nil), nil), "[ . ]")
	})
}
