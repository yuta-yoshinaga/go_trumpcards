package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// prsiPlayerStr returns the display string for a single Prsi player.
func prsiPlayerStr(player *domain.PrsiPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("prsi.playerLine",
		"name", cuiPlayerName(player, i),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// prsiIsLegalPlay reports whether card may be played on top given the discard
// top and pending 7-penalty. Mirrors Prsi.isValidPlay so the CUI can flag
// playable cards without a domain accessor: during a 7-penalty only a 7 stacks;
// otherwise the card must match the top's suit or rank (any card on an empty pile).
func prsiIsLegalPlay(card, top *domain.Card, penalty int) bool {
	if card == nil {
		return false
	}
	if penalty > 0 {
		return card.GetValue() == domain.PrsiSevenValue
	}
	if top == nil {
		return true
	}
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// prsiLegalCardsLine returns the localized "playable cards" line for the human's
// hand, or the draw guidance when no card is legal (the human must draw).
func prsiLegalCardsLine(player *domain.PrsiPlayer, top *domain.Card, penalty int) string {
	var legal []string
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if prsiIsLegalPlay(card, top, penalty) {
			legal = append(legal, "["+strconv.Itoa(i)+"]"+cuiCardStr(card))
		}
	}
	if len(legal) == 0 {
		return color.Yellow(i18n.T("prsi.noLegalPlay")) + "\n"
	}
	return color.Yellow(i18n.Tf("prsi.legalPlays", "cards", strings.Join(legal, ", "))) + "\n"
}

// PrsiCuiPresenter renders the Prší CUI view.
type PrsiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PrsiCuiPresenter) Output(g interfaces.PrsiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("prsi.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("prsi.header",
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("prsi.discardLine", "card", cuiCardStr(top)))
			if g.GetPenaltyDrawCount() > 0 {
				b.WriteString(i18n.Tf("prsi.penaltyDraw",
					"count", strconv.Itoa(g.GetPenaltyDrawCount())))
			}
			// **スキップも重ねられる (#4772)。**7 の累積ペナルティは出して
			// いたのに、エース/ジャックの累積は CUI にも Web にも出ていな
			// かった。2以上なら複数人が連続で飛ばされる。
			if g.GetPendingSkips() > 0 {
				b.WriteString(i18n.Tf("prsi.pendingSkips",
					"count", strconv.Itoa(g.GetPendingSkips())))
			}
			b.WriteString("\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(prsiPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("prsi.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		if g.GetPhase() == domain.PrsiPhasePlay {
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("prsi.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// On the human's turn, spell out which hand indices are legal (or that
			// they must draw) so they need not re-derive the match rule each turn.
			if human := g.GetPlayer(currentIdx); human != nil && human.GetIsHuman() {
				b.WriteString(prsiLegalCardsLine(human, g.GetDiscardTop(), g.GetPenaltyDrawCount()))
			}
			b.WriteString(i18n.T("prsi.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("prsi.promptDrawHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PrsiCuiPresenter) ActionLogOutput(g interfaces.PrsiGame) string {
	return actionLogOutputText(g)
}
