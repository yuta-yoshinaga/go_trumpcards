package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SlapjackCuiPresenter renders the Slapjack CUI view.
type SlapjackCuiPresenter struct{}

// slapjackDifficultyStr は難易度の表示名を返す。Web のセレクトと同じ 3 つ。
//
// 未知の値は番号のまま出す ── 「Easy」に丸めると、設定が壊れていることが
// 画面から消える。
func slapjackDifficultyStr(d domain.SlapjackCpuDifficulty) string {
	switch d {
	case domain.SlapjackCpuEasy:
		return i18n.T("slapjack.difficultyEasy")
	case domain.SlapjackCpuNormal:
		return i18n.T("slapjack.difficultyNormal")
	case domain.SlapjackCpuHard:
		return i18n.T("slapjack.difficultyHard")
	}
	return strconv.Itoa(int(d))
}

// Output renders the current game state for the active locale (#1699).
func (p *SlapjackCuiPresenter) Output(g interfaces.SlapjackGame, lastErr error) string {
	return buildCuiOutput(i18n.T("slapjack.helpTitle"), func(b *strings.Builder) {
		cpu := g.GetPlayer(1)
		human := g.GetPlayer(0)

		// **難易度は変えられるのに、変えた結果を確かめる手段が無かった** (#5579)。
		// `sd` コマンドは前からあり、Web は選択中の値を常に出している。
		b.WriteString(i18n.Tf("slapjack.difficultyLine",
			"difficulty", slapjackDifficultyStr(g.GetConfig().CpuDifficulty)) + "\n")

		b.WriteString(i18n.Tf("slapjack.cpuStockLine",
			"count", strconv.Itoa(cpu.GetStockSize())) + "\n")
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("slapjack.boardLabel")) + " ")
		centerSize := strconv.Itoa(g.GetCenterPileSize())
		if c := g.GetTopCard(); c != nil {
			b.WriteString(i18n.Tf("slapjack.boardCard",
				"card", cuiCardStr(c),
				"count", centerSize) + "\n")
		} else {
			b.WriteString(i18n.Tf("slapjack.boardEmpty", "count", centerSize) + "\n")
		}
		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("slapjack.humanStockLine",
			"count", strconv.Itoa(human.GetStockSize())) + "\n")

		if g.IsTopJack() {
			b.WriteString(color.Yellow(i18n.T("slapjack.promptJackOnTop")) + "\n")
		} else if g.GetCurrentTurnIdx() == 0 {
			b.WriteString(i18n.T("slapjack.promptHumanTurn") + "\n")
		} else {
			b.WriteString(i18n.T("slapjack.promptCpuTurn") + "\n")
		}

		switch g.GetLastEvent().Kind {
		case domain.SlapjackEventSlapCorrect:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Green(i18n.T("slapjack.slapCorrectHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("slapjack.slapCorrectCpu")) + "\n")
			}
		case domain.SlapjackEventSlapWrong:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Red(i18n.T("slapjack.slapWrongHuman")) + "\n")
			} else {
				b.WriteString(color.Green(i18n.T("slapjack.slapWrongCpu")) + "\n")
			}
		}

		if g.GetGameEndFlag() {
			if g.GetWinnerIdx() == 0 {
				b.WriteString(color.Green(i18n.T("slapjack.winHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("slapjack.winCpu")) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SlapjackCuiPresenter) ActionLogOutput(g interfaces.SlapjackGame) string {
	return actionLogOutputTextForSeats[*domain.SlapjackPlayer](g)
}
