package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// WarCuiPresenter renders the War CUI view.
type WarCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *WarCuiPresenter) Output(w interfaces.WarGame, lastErr error) string {
	return buildCuiOutput(i18n.T("war.helpTitle"), func(b *strings.Builder) {
		cpu := w.GetPlayer(1)
		human := w.GetPlayer(0)

		// **引き分け打ち切りがいつ来るかを出す。**Web はアリーナに
		// 「ラウンド: n / max」を常時出しているのに、CUI は一度も出しておらず、
		// あと何ラウンドで打ち切り (保有枚数の多い方の勝ち) になるかが分からなかった (#4865)。
		rounds, maxRounds := w.GetRoundsPlayed(), w.GetConfig().MaxRounds
		line := i18n.Tf("war.roundProgress",
			"played", strconv.Itoa(rounds), "max", strconv.Itoa(maxRounds))
		// 9 割を超えたら強調する。打ち切りが目前だと分かる必要がある。
		if maxRounds > 0 && rounds*10 >= maxRounds*9 {
			line = color.Yellow(line)
		}
		b.WriteString(line + "\n")

		b.WriteString(i18n.Tf("war.cpuStats",
			"draw", strconv.Itoa(cpu.GetDrawPileSize()),
			"discard", strconv.Itoa(cpu.GetDiscardPileSize()),
			"total", strconv.Itoa(cpu.TotalCards())) + "\n")

		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("war.boardLabel")) + " ")
		if c := w.GetCpuRevealed(); c != nil {
			b.WriteString(i18n.Tf("war.boardCpu", "card", cuiCardStr(c)) + "  ")
		} else {
			b.WriteString(i18n.T("war.boardCpuEmpty") + "  ")
		}
		if c := w.GetPlayerRevealed(); c != nil {
			b.WriteString(i18n.Tf("war.boardHuman", "card", cuiCardStr(c)))
		} else {
			b.WriteString(i18n.T("war.boardHumanEmpty"))
		}
		b.WriteString(i18n.Tf("war.boardPot",
			"count", strconv.Itoa(w.GetWarPotSize())) + "\n")
		b.WriteString("----------\n")

		b.WriteString(i18n.Tf("war.humanStats",
			"draw", strconv.Itoa(human.GetDrawPileSize()),
			"discard", strconv.Itoa(human.GetDiscardPileSize()),
			"total", strconv.Itoa(human.TotalCards())) + "\n")

		switch w.GetPhase() {
		case domain.WarPhaseReveal:
			b.WriteString(i18n.T("war.promptReveal") + "\n")
		case domain.WarPhaseResolved:
			if w.GetLastWinnerIdx() == 0 {
				b.WriteString(color.Green(i18n.T("war.promptResolvedHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("war.promptResolvedCpu")) + "\n")
			}
		case domain.WarPhaseWarBury:
			b.WriteString(color.Yellow(i18n.T("war.promptWarBury")) + "\n")
		}

		if w.GetGameEndFlag() {
			if w.GetWinnerIdx() == 0 {
				b.WriteString(color.Green(i18n.T("war.gameWinHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("war.gameWinCpu")) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WarCuiPresenter) ActionLogOutput(w interfaces.WarGame) string {
	return actionLogOutputTextForSeats[*domain.WarPlayer](w)
}
