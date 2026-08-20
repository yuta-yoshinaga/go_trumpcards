package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// EgyptianRatscrewCuiPresenter renders the Egyptian Ratscrew CUI view.
type EgyptianRatscrewCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *EgyptianRatscrewCuiPresenter) Output(g interfaces.EgyptianRatscrewGame, lastErr error) string {
	return buildCuiOutput(i18n.T("egyptianratscrew.helpTitle"), func(b *strings.Builder) {
		cpu := g.GetPlayer(1)
		human := g.GetPlayer(0)
		if cpu == nil || human == nil {
			return
		}

		b.WriteString(i18n.Tf("egyptianratscrew.cpuStockLine",
			"count", strconv.Itoa(cpu.GetStockSize())) + "\n")
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("egyptianratscrew.boardLabel")) + " ")
		centerSize := strconv.Itoa(g.GetCenterPileSize())
		if c := g.GetTopCard(); c != nil {
			b.WriteString(i18n.Tf("egyptianratscrew.boardCard",
				"card", cuiCardStr(c),
				"count", centerSize) + "\n")
		} else {
			b.WriteString(i18n.Tf("egyptianratscrew.boardEmpty",
				"count", centerSize) + "\n")
		}
		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("egyptianratscrew.humanStockLine",
			"count", strconv.Itoa(human.GetStockSize())) + "\n")

		if g.GetChanceRemaining() > 0 {
			b.WriteString(color.Yellow(i18n.Tf("egyptianratscrew.promptChance",
				"count", strconv.Itoa(g.GetChanceRemaining()))) + "\n")
			// Name who must answer the chance (web shows chanceResponder) and, when
			// a face card currently sits on top, which card is demanding it.
			responder := i18n.T("egyptianratscrew.responderCpu")
			if g.GetCurrentTurnIdx() == 0 {
				responder = i18n.T("egyptianratscrew.responderHuman")
			}
			b.WriteString(i18n.Tf("egyptianratscrew.chanceResponder", "name", responder) + "\n")
			if c := g.GetTopCard(); c != nil && g.IsTopFaceCard() {
				b.WriteString(i18n.Tf("egyptianratscrew.chanceTrigger", "card", cuiCardStr(c)) + "\n")
			}
		}
		if g.IsSlappable() {
			b.WriteString(color.Yellow(i18n.T("egyptianratscrew.promptSlappable")) + "\n")
		} else if g.GetCurrentTurnIdx() == 0 {
			b.WriteString(i18n.T("egyptianratscrew.promptHumanTurn") + "\n")
		} else {
			b.WriteString(i18n.T("egyptianratscrew.promptCpuTurn") + "\n")
		}

		switch g.GetLastEvent().Kind {
		case domain.EgyptianRatscrewEventSlapCorrect:
			label := i18n.T("egyptianratscrew.slapLabelDefault")
			switch g.GetLastEvent().SlapReason {
			case domain.EgyptianRatscrewSlapReasonPair:
				label = i18n.T("egyptianratscrew.slapLabelPair")
			case domain.EgyptianRatscrewSlapReasonSandwich:
				label = i18n.T("egyptianratscrew.slapLabelSandwich")
			}
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Green(i18n.Tf("egyptianratscrew.slapCorrectHuman",
					"label", label)) + "\n")
			} else {
				b.WriteString(color.Red(i18n.Tf("egyptianratscrew.slapCorrectCpu",
					"label", label)) + "\n")
			}
		case domain.EgyptianRatscrewEventSlapWrong:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Red(i18n.T("egyptianratscrew.slapWrongHuman")) + "\n")
			} else {
				b.WriteString(color.Green(i18n.T("egyptianratscrew.slapWrongCpu")) + "\n")
			}
		case domain.EgyptianRatscrewEventChanceWin:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Green(i18n.T("egyptianratscrew.chanceWinHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("egyptianratscrew.chanceWinCpu")) + "\n")
			}
		}

		if g.GetGameEndFlag() {
			switch g.GetWinnerIdx() {
			case 0:
				b.WriteString(color.Green(i18n.T("egyptianratscrew.winHuman")) + "\n")
			case -1:
				b.WriteString(color.Yellow(i18n.T("egyptianratscrew.winDraw")) + "\n")
			default:
				b.WriteString(color.Red(i18n.T("egyptianratscrew.winCpu")) + "\n")
			}
		} else {
			// Game over takes precedence; otherwise show the error.
			cuiErrorBlock(b, lastErr)
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EgyptianRatscrewCuiPresenter) ActionLogOutput(g interfaces.EgyptianRatscrewGame) string {
	return actionLogOutputTextForSeats[*domain.EgyptianRatscrewPlayer](g)
}
