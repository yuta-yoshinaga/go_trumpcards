//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BassetCuiPresenter renders Basset state for the terminal.
type BassetCuiPresenter struct{}

// Output renders the current Basset state.
func (p *BassetCuiPresenter) Output(g interfaces.BassetGame, lastErr error) string {
	var s strings.Builder
	s.WriteString("----------\n")
	s.WriteString(i18n.Tf("basset.chipsLine", "chips", strconv.Itoa(g.GetChips())) + "\n")
	s.WriteString(i18n.Tf("basset.phaseLine", "phase", strconv.Itoa(g.GetPhase())) + "\n")
	if bet, rank := g.GetBet(); bet != nil {
		s.WriteString(i18n.Tf("basset.betLine", "rank", strconv.Itoa(rank), "amount", strconv.Itoa(bet.Amount)) + "\n")
	}
	if turn := g.GetLastTurn(); turn != nil {
		s.WriteString("banker: " + cuiCardStr(turn.BankerCard) + " player: " + cuiCardStr(turn.PlayerCard) + "\n")
	}
	if lastErr != nil {
		s.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}
	if g.GetPhase() == domain.BassetPhaseDecision {
		s.WriteString(i18n.T("basset.promptDecision") + "\n")
	}
	return s.String()
}

// ActionLogOutput renders the action log.
func (p *BassetCuiPresenter) ActionLogOutput(g interfaces.BassetGame) string {
	return actionLogOutputText(g)
}
