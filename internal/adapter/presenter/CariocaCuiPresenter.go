//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// cariocaPlayerStr returns the display string for a single Carioca player.
func cariocaPlayerStr(player *domain.CariocaPlayer, i int) string {
	var b strings.Builder
	contractStatus := i18n.T("carioca.notMet")
	if player.IsContractMet() {
		contractStatus = i18n.T("carioca.met")
	}
	b.WriteString(i18n.Tf("carioca.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", contractStatus) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	if player.GetMeldCount() > 0 {
		b.WriteString(i18n.T("carioca.meldsHeader"))
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			if mi > 0 {
				b.WriteString(" | ")
			}
			cardStrs := make([]string, len(meld))
			for ci, c := range meld {
				cardStrs[ci] = cuiCardStr(c)
			}
			b.WriteString(strings.Join(cardStrs, " "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// cariocaContractDescription returns a human-readable description of the round's contract.
func cariocaContractDescription(c domain.Contract) string {
	parts := make([]string, 0, len(c.Slots))
	for _, slot := range c.Slots {
		if slot.Kind == domain.ContractSlotSet {
			parts = append(parts, i18n.Tf("carioca.slotSet", "n", strconv.Itoa(slot.Size)))
		} else {
			parts = append(parts, i18n.Tf("carioca.slotRun", "n", strconv.Itoa(slot.Size)))
		}
	}
	return strings.Join(parts, " + ")
}

// CariocaCuiPresenter renders the Carioca CUI view.
type CariocaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CariocaCuiPresenter) Output(g interfaces.CariocaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("carioca.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("carioca.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(domain.CariocaTotalRounds),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		contract := g.GetCurrentContract()
		b.WriteString(i18n.Tf("carioca.contractLine",
			"contract", cariocaContractDescription(contract)) + "\n")
		// Expand each slot's detailed requirement (a set is same-rank, a run is a
		// same-suit sequence) with its index, plus the meld slot-index syntax, so
		// the human knows exactly what each slot needs and how to target it.
		for i, slot := range contract.Slots {
			detail := i18n.Tf("carioca.slotDetailRun", "n", strconv.Itoa(slot.Size))
			if slot.Kind == domain.ContractSlotSet {
				detail = i18n.Tf("carioca.slotDetailSet", "n", strconv.Itoa(slot.Size))
			}
			b.WriteString(i18n.Tf("carioca.slotLine",
				"idx", strconv.Itoa(i),
				"detail", detail) + "\n")
		}
		if len(contract.Slots) > 0 {
			b.WriteString(i18n.T("carioca.meldSyntaxHint") + "\n")
		}

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("carioca.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(cariocaPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("carioca.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.CariocaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("carioca.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("carioca.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("carioca.promptDrawHelpDiscard") + "\n")
		case domain.CariocaPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("carioca.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur != nil && cur.IsContractMet() {
				// Contract already down: extra melds / layoffs are optional, discard ends the turn.
				b.WriteString(i18n.T("carioca.promptPlayHelpAfterContract") + "\n")
				b.WriteString(i18n.T("carioca.promptPlayHelpMeldExtra") + "\n")
			} else {
				b.WriteString(i18n.T("carioca.promptPlayHelpContractRequired") + "\n")
				b.WriteString(i18n.T("carioca.promptPlayHelpMeldContract") + "\n")
			}
			b.WriteString(i18n.T("carioca.promptPlayHelpLayoff") + "\n")
			b.WriteString(i18n.T("carioca.promptPlayHelpDiscard") + "\n")
		case domain.CariocaPhaseRoundEnd:
			b.WriteString(i18n.T("carioca.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("carioca.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CariocaCuiPresenter) ActionLogOutput(g interfaces.CariocaGame) string {
	return actionLogOutputTextForSeats[*domain.CariocaPlayer](g)
}
