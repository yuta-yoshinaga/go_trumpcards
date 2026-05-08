package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// cribbagePlayerStr returns the display string for a single Cribbage player.
func cribbagePlayerStr(player *domain.CribbagePlayer, i int, dealerIdx int) string {
	var b strings.Builder
	dealerMark := ""
	if i == dealerIdx {
		dealerMark = i18n.T("cribbage.dealerMark")
	}
	b.WriteString(i18n.Tf("cribbage.playerLine",
		"name", cuiPlayerName(player, i),
		"dealerMark", dealerMark,
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CribbageCuiPresenter renders the Cribbage CUI view.
type CribbageCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CribbageCuiPresenter) Output(g interfaces.CribbageGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cribbage.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("cribbage.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"dealer", strconv.Itoa(g.GetDealerIdx())) + "\n")

		// Starter card
		if starter := g.GetStarter(); starter != nil {
			b.WriteString(i18n.Tf("cribbage.starterLine",
				"card", cuiCardStr(starter)) + "\n")
		}

		// Pegging info
		phase := g.GetPhase()
		if phase == domain.CribbagePhasePegging {
			b.WriteString(i18n.Tf("cribbage.peggingTotal",
				"count", strconv.Itoa(g.GetPegCount())) + "\n")
			pegCards := g.GetPegPlayedCards()
			if len(pegCards) > 0 {
				cardStrs := make([]string, len(pegCards))
				for i, c := range pegCards {
					cardStrs[i] = cuiCardStr(c)
				}
				b.WriteString(i18n.Tf("cribbage.peggingCards",
					"cards", strings.Join(cardStrs, ", ")) + "\n")
			}
		}

		// Players
		for i := range domain.CribbagePlayerCnt {
			player := g.GetPlayer(i)
			if player != nil {
				b.WriteString(cribbagePlayerStr(player, i, g.GetDealerIdx()))
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("cribbage.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.CribbagePhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cribbage.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("cribbage.promptDiscardHelp") + "\n")
		case domain.CribbagePhasePegging:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cribbage.promptPegging",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("cribbage.promptPeggingHelp") + "\n")
			b.WriteString(i18n.T("cribbage.promptPeggingGo") + "\n")
		case domain.CribbagePhaseShow:
			b.WriteString(i18n.T("cribbage.promptShow") + "\n")
			p.writeShowDetails(b, g)
			b.WriteString(i18n.T("cribbage.promptShowHelp") + "\n")
		case domain.CribbagePhaseRoundEnd:
			b.WriteString(i18n.T("cribbage.promptRoundEnd") + "\n")
			p.writeShowDetails(b, g)
			b.WriteString(i18n.T("cribbage.promptRoundEndHelp") + "\n")
		}
	})
}

// writeShowDetails prints score detail lines for the show phase.
func (p *CribbageCuiPresenter) writeShowDetails(b *strings.Builder, g interfaces.CribbageGame) {
	details := g.GetHandScoreDetails()
	labelKeys := [3]string{
		"cribbage.showLabelPone",
		"cribbage.showLabelDealer",
		"cribbage.showLabelCrib",
	}
	for i, d := range details {
		if d == nil {
			continue
		}
		b.WriteString(i18n.Tf("cribbage.showDetailLine",
			"label", i18n.T(labelKeys[i]),
			"total", strconv.Itoa(d.Total),
			"fifteens", strconv.Itoa(d.Fifteens),
			"pairs", strconv.Itoa(d.Pairs),
			"runs", strconv.Itoa(d.Runs),
			"flush", strconv.Itoa(d.Flush),
			"nobs", strconv.Itoa(d.Nobs)) + "\n")
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CribbageCuiPresenter) ActionLogOutput(g interfaces.CribbageGame) string {
	return actionLogOutputText(g)
}
