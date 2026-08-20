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

// machiavelliPlayerStr returns the display string for a single Machiavelli player.
func machiavelliPlayerStr(g interfaces.MachiavelliGame, i int) string {
	player := g.GetPlayer(i)
	var b strings.Builder
	b.WriteString(i18n.Tf("machiavelli.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// machiavelliTableStr renders the shared table of melds.
func machiavelliTableStr(g interfaces.MachiavelliGame) string {
	var b strings.Builder
	table := g.GetTable()
	if len(table) == 0 {
		b.WriteString(i18n.T("machiavelli.tableEmpty") + "\n")
		return b.String()
	}
	for mi, meld := range table {
		parts := make([]string, 0, len(meld))
		for _, c := range meld {
			parts = append(parts, cuiCardStr(c))
		}
		b.WriteString(i18n.Tf("machiavelli.meldLine",
			"idx", strconv.Itoa(mi),
			"cards", strings.Join(parts, " ")) + "\n")
	}
	return b.String()
}

// MachiavelliCuiPresenter renders the Machiavelli CUI view.
type MachiavelliCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MachiavelliCuiPresenter) Output(g interfaces.MachiavelliGame, lastErr error) string {
	return buildCuiOutput(i18n.T("machiavelli.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("machiavelli.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetTargetRounds()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		b.WriteString(i18n.T("machiavelli.tableTitle") + "\n")
		b.WriteString(machiavelliTableStr(g))

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(machiavelliPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("machiavelli.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MachiavelliPhaseTurn:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("machiavelli.promptTurn",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("machiavelli.promptDrawHelp") + "\n")
			b.WriteString(i18n.T("machiavelli.promptNewMeldHelp") + "\n")
			// Spell out the layoff argument format and how many table melds it can
			// target; when the table is empty, layoff is impossible.
			if meldCount := len(g.GetTable()); meldCount > 0 {
				b.WriteString(i18n.Tf("machiavelli.promptLayoffHelp",
					"count", strconv.Itoa(meldCount)) + "\n")
				// 場の組み替えはこのゲームの核心だが、場が空なら組み替える対象が
				// 無いので layoff と同じ条件で出す。
				b.WriteString(i18n.T("machiavelli.promptRearrangeHelp") + "\n")
			} else {
				b.WriteString(i18n.T("machiavelli.promptLayoffNone") + "\n")
			}
		case domain.MachiavelliPhaseRoundEnd:
			b.WriteString(i18n.T("machiavelli.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("machiavelli.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MachiavelliCuiPresenter) ActionLogOutput(g interfaces.MachiavelliGame) string {
	return actionLogOutputTextForSeats[*domain.MachiavelliPlayer](g)
}
