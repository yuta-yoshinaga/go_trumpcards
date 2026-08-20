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

// contractRummyPlayerStr returns the display string for a single ContractRummy player.
func contractRummyPlayerStr(player *domain.ContractRummyPlayer, i int) string {
	var b strings.Builder
	contractStatus := i18n.T("contractrummy.notMet")
	if player.IsContractMet() {
		contractStatus = i18n.T("contractrummy.met")
	}
	b.WriteString(i18n.Tf("contractrummy.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", contractStatus) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	if player.GetMeldCount() > 0 {
		b.WriteString(i18n.T("contractrummy.meldsHeader"))
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

// contractDescription returns a human-readable description of the round's contract.
func contractDescription(c domain.Contract) string {
	parts := make([]string, 0, len(c.Slots))
	for _, slot := range c.Slots {
		if slot.Kind == domain.ContractSlotSet {
			parts = append(parts, i18n.Tf("contractrummy.slotSet", "n", strconv.Itoa(slot.Size)))
		} else {
			parts = append(parts, i18n.Tf("contractrummy.slotRun", "n", strconv.Itoa(slot.Size)))
		}
	}
	return strings.Join(parts, " + ")
}

// ContractRummyCuiPresenter renders the Contract Rummy CUI view.
type ContractRummyCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ContractRummyCuiPresenter) Output(g interfaces.ContractRummyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("contractrummy.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("contractrummy.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(domain.ContractRummyTotalRounds),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		b.WriteString(i18n.Tf("contractrummy.contractLine",
			"contract", contractDescription(g.GetCurrentContract())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("contractrummy.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(contractRummyPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("contractrummy.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.ContractRummyPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("contractrummy.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("contractrummy.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("contractrummy.promptDrawHelpDiscard") + "\n")
		case domain.ContractRummyPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("contractrummy.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur != nil && cur.IsContractMet() {
				// Contract already down: extra melds / layoffs are optional, discard ends the turn.
				b.WriteString(i18n.T("contractrummy.promptPlayHelpAfterContract") + "\n")
				b.WriteString(i18n.T("contractrummy.promptPlayHelpMeldExtra") + "\n")
			} else {
				b.WriteString(i18n.T("contractrummy.promptPlayHelpContractRequired") + "\n")
				b.WriteString(i18n.T("contractrummy.promptPlayHelpMeldContract") + "\n")
			}
			b.WriteString(i18n.T("contractrummy.promptPlayHelpLayoff") + "\n")
			b.WriteString(i18n.T("contractrummy.promptPlayHelpDiscard") + "\n")
		case domain.ContractRummyPhaseRoundEnd:
			b.WriteString(i18n.T("contractrummy.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("contractrummy.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ContractRummyCuiPresenter) ActionLogOutput(g interfaces.ContractRummyGame) string {
	return actionLogOutputTextForSeats[*domain.ContractRummyPlayer](g)
}
