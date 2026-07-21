//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newCourtPieceForCuiTest() *domain.CourtPiece {
	cp := domain.NewDefaultCourtPiece()
	cp.Reset()
	return cp
}

func TestCourtPieceCuiPresenter_Output_PhaseLabels(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	t.Run("trump phase prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
		cp.SetCallerIdx(0)
		out := p.Output(cp, nil)
		assert.Contains(t, out, "courtpiece")
	})

	t.Run("play phase prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		out := p.Output(cp, nil)
		// i18n placeholders are emitted literally in tests; key the assertion to a
		// section we know is present regardless of locale.
		assert.Contains(t, out, "courtpiece.promptPlay")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrickEnd)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("round end prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block included", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		out := p.Output(cp, errors.New("bad input"))
		assert.Contains(t, out, "bad input")
	})

	t.Run("game end banner", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseGameEnd)
		cp.SetTeamScore(0, domain.CourtPieceDefaultPointLimit)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("caller and player lines render with team labels", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		cp.SetCallerIdx(0)
		out := p.Output(cp, nil)
		// The caller line (which now carries the caller's team) and one player
		// line per seat are rendered. Assertions key on the section markers
		// because i18n templates resolve to their literal keys in tests.
		assert.Contains(t, out, "courtpiece.callerLine")
		assert.Contains(t, out, "courtpiece.playerLine")
	})
}

func TestCourtPieceCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	t.Run("trump hint", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
		cp.SetCallerIdx(0)
		out := p.HintOutput(cp)
		assert.NotEmpty(t, out)
	})

	t.Run("play hint", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		cp.SetCurrentPlayerIdx(0)
		out := p.HintOutput(cp)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint when out of turn", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetCurrentPlayerIdx(1) // CPU's turn
		out := p.HintOutput(cp)
		assert.Contains(t, strings.ToLower(out), "hint")
	})
}

func TestCourtPieceCuiPresenter_ActionLogOutput(t *testing.T) {
	cp := newCourtPieceForCuiTest()
	p := new(presenter.CourtPieceCuiPresenter)
	out := p.ActionLogOutput(cp)
	assert.NotNil(t, out)
}
