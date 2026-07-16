//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// macauPlayerStr returns the display string for a single Macau player.
func macauPlayerStr(player *domain.MacauPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("macau.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// macauDirectionName returns the human-readable play-direction marker.
func macauDirectionName(direction int) string {
	if direction < 0 {
		return "←"
	}
	return "→"
}

// MacauCuiPresenter renders the Macau CUI view.
type MacauCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MacauCuiPresenter) Output(g interfaces.MacauGame, lastErr error) string {
	return buildCuiOutput(i18n.T("macau.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("macau.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"dir", macauDirectionName(g.GetDirection())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("macau.discardLine", "card", cuiCardStr(top)))
			if g.GetChosenSuit() > 0 {
				b.WriteString(i18n.Tf("macau.chosenSuit",
					"suit", suitDisplayName(g.GetChosenSuit())))
			}
			b.WriteString("\n")
		}

		if g.GetPenaltyDrawCount() > 0 {
			b.WriteString(i18n.Tf("macau.penaltyLine",
				"count", strconv.Itoa(g.GetPenaltyDrawCount())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(macauPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("macau.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MacauPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("macau.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("macau.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("macau.promptDrawHelp") + "\n")
		case domain.MacauPhaseChooseSuit:
			b.WriteString(i18n.T("macau.promptChooseSuit") + "\n")
			b.WriteString(i18n.T("macau.promptChooseSuitHelp") + "\n")
		case domain.MacauPhaseMustDeclare:
			b.WriteString(i18n.T("macau.promptMustDeclare") + "\n")
			b.WriteString(i18n.T("macau.promptMustDeclareHelp") + "\n")
		case domain.MacauPhaseRoundEnd:
			b.WriteString(i18n.T("macau.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("macau.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MacauCuiPresenter) ActionLogOutput(g interfaces.MacauGame) string {
	return actionLogOutputText(g)
}

// HintOutput lists the human's currently playable card indices. During a
// penalty chain IsValidPlay already restricts to counter cards, so the same
// scan works; when nothing is playable it advises drawing (or, mid-penalty,
// accepting the accumulated penalty).
func (p *MacauCuiPresenter) HintOutput(g interfaces.MacauGame) string {
	if g.GetPhase() != domain.MacauPhasePlay || !g.IsHumanTurn() {
		return i18n.T("macau.hintNone") + "\n"
	}
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	if player == nil {
		return i18n.T("macau.hintNone") + "\n"
	}
	var parts []string
	for i := 0; i < player.GetCardsSize(); i++ {
		if g.IsValidPlay(player.GetCard(i)) {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(player.GetCard(i)))
		}
	}
	if len(parts) == 0 {
		if g.GetPenaltyDrawCount() > 0 {
			return color.Yellow(i18n.T("macau.hintReceivePenalty")) + "\n"
		}
		return color.Yellow(i18n.T("macau.hintDraw")) + "\n"
	}
	return color.Yellow(i18n.Tf("macau.hintPlayable", "cards", strings.Join(parts, " "))) + "\n"
}
