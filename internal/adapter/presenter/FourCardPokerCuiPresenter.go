//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FourCardPokerCuiPresenter is the Four Card Poker CUI presenter.
type FourCardPokerCuiPresenter struct{}

// Output renders the game state for the CLI.
func (p *FourCardPokerCuiPresenter) Output(g interfaces.FourCardPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("fourcardpoker.chipsLine", "chips", strconv.Itoa(g.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("fourcardpoker.phaseLine", "phase", p.phaseStr(g.GetPhase())) + "\n")

	playerHand := g.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("fourcardpoker.playerHeader")) + " ---\n")
		rank := g.GetPlayerHandRank()
		if rank > 0 && rank < len(domain.FourCardHandNames) {
			sb.WriteString(i18n.Tf("fourcardpoker.handLine", "hand", domain.FourCardHandNames[rank]) + "\n")
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
		// At End phase, surface the best-4 subset so the player can see
		// why the showdown went the way it did.
		if g.GetPhase() == domain.FourCardPokerPhaseEnd {
			if best := g.GetPlayerBest(); len(best) > 0 {
				p2 := make([]string, len(best))
				for i, card := range best {
					p2[i] = cuiCardStr(card)
				}
				sb.WriteString(i18n.T("fourcardpoker.bestHand") + ": " + strings.Join(p2, ",") + "\n")
			}
		}
	}

	// Dealer display: while in action phase show only the upcard; on end reveal all.
	dealerHand := g.GetDealerHand()
	if len(dealerHand) > 0 {
		switch g.GetPhase() {
		case domain.FourCardPokerPhaseAction:
			up := g.GetDealerUpCard()
			if up != nil {
				sb.WriteString("--- " + color.Bold(i18n.T("fourcardpoker.dealerHeader")) + " ---\n")
				sb.WriteString(i18n.T("fourcardpoker.dealerUpcard") + ": " + cuiCardStr(up) + "\n")
			}
		case domain.FourCardPokerPhaseEnd:
			sb.WriteString("--- " + color.Bold(i18n.T("fourcardpoker.dealerHeader")) + " ---\n")
			rank := g.GetDealerHandRank()
			if rank > 0 && rank < len(domain.FourCardHandNames) {
				sb.WriteString(i18n.Tf("fourcardpoker.handLine", "hand", domain.FourCardHandNames[rank]) + "\n")
			}
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
			if best := g.GetDealerBest(); len(best) > 0 {
				p2 := make([]string, len(best))
				for i, card := range best {
					p2[i] = cuiCardStr(card)
				}
				sb.WriteString(i18n.T("fourcardpoker.bestHand") + ": " + strings.Join(p2, ",") + "\n")
			}
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if g.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("fourcardpoker.anteLine", "ante", strconv.Itoa(g.GetAnteBet())) + "\n")
		if g.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("fourcardpoker.playLine", "play", strconv.Itoa(g.GetPlayBet())) + "\n")
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("fourcardpoker.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("fourcardpoker.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("fourcardpoker.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("fourcardpoker.push")) + "\n")
		default:
		}
		// Break the total down by payout bucket (matching the web breakdown);
		// list only the non-zero buckets, as the web UI does.
		for _, item := range []struct {
			key    string
			amount int
		}{
			{"fourcardpoker.antePayoutLine", g.GetAntePayout()},
			{"fourcardpoker.playPayoutLine", g.GetPlayPayout()},
			{"fourcardpoker.anteBonusPayoutLine", g.GetAnteBonusPayout()},
			{"fourcardpoker.acesUpPayoutLine", g.GetAcesUpPayout()},
		} {
			if item.amount != 0 {
				sb.WriteString(i18n.Tf(item.key, "payout", strconv.Itoa(item.amount)) + "\n")
			}
		}
		sb.WriteString(i18n.Tf("fourcardpoker.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput renders the action log.
func (p *FourCardPokerCuiPresenter) ActionLogOutput(g interfaces.FourCardPokerGame) string {
	return actionLogOutputText(g)
}

func (p *FourCardPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.FourCardPokerPhaseBet:
		return i18n.T("fourcardpoker.phaseBet")
	case domain.FourCardPokerPhaseAction:
		return i18n.T("fourcardpoker.phaseAction")
	case domain.FourCardPokerPhaseEnd:
		return i18n.T("fourcardpoker.phaseEnd")
	default:
		return i18n.T("fourcardpoker.phaseUnknown")
	}
}
