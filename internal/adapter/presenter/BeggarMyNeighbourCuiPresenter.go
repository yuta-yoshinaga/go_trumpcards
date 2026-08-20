package presenter

import (
	"math"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BeggarMyNeighbourCuiPresenter renders the Beggar-My-Neighbour CUI view.
type BeggarMyNeighbourCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BeggarMyNeighbourCuiPresenter) Output(g interfaces.BeggarMyNeighbourGame, lastErr error) string {
	return buildCuiOutput(i18n.T("beggarmyneighbour.helpTitle"), func(b *strings.Builder) {
		cpu := g.GetPlayer(1)
		human := g.GetPlayer(0)

		// **引き分け打ち切りがいつ来るかを出す。**上限は設定で大きく変えられ
		// (BeggarMyNeighbourMinMaxRounds〜MaxMaxRounds)、しかも自動進行する
		// ゲームなので、あとどれだけかが分からないと待つほかない (#4896)。
		// Web は進捗バーを常時出している。
		rounds, maxRounds := g.GetRoundsPlayed(), g.GetConfig().MaxRounds
		line := i18n.Tf("beggarmyneighbour.roundProgress",
			"played", strconv.Itoa(rounds), "max", strconv.Itoa(maxRounds))
		// 9 割を超えたら強調する。打ち切りが目前だと分かる必要がある。
		if maxRounds > 0 && rounds*10 >= maxRounds*9 {
			line = color.Yellow(line)
		}
		b.WriteString(line + "\n")

		b.WriteString(i18n.Tf("beggarmyneighbour.cpuStats",
			"draw", strconv.Itoa(cpu.GetDrawPileSize()),
			"discard", strconv.Itoa(cpu.GetDiscardPileSize()),
			"total", strconv.Itoa(cpu.TotalCards())) + "\n")

		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("beggarmyneighbour.boardLabel")) + "\n")
		b.WriteString(i18n.Tf("beggarmyneighbour.boardPile",
			"count", strconv.Itoa(g.GetCentralPileSize())) + "\n")
		if c := g.GetLastCardPlayed(); c != nil {
			b.WriteString(i18n.Tf("beggarmyneighbour.boardLastCard", "card", cuiCardStr(c)) + "\n")
		} else {
			b.WriteString(i18n.T("beggarmyneighbour.boardLastCardEmpty") + "\n")
		}
		b.WriteString("----------\n")

		b.WriteString(i18n.Tf("beggarmyneighbour.humanStats",
			"draw", strconv.Itoa(human.GetDrawPileSize()),
			"discard", strconv.Itoa(human.GetDiscardPileSize()),
			"total", strconv.Itoa(human.TotalCards())) + "\n")

		// **autoplay で何百ラウンドも回せるゲーム** (#5682)。生数値だけだと、どちらが
		// 優勢かを毎回自分で割ることになる。Web は色分けバーで常時出している。
		// 計算式は Web の youPct と同じ (両者 0 枚なら 50/50 — 0 除算しない)。
		youPct := 50
		if held := human.TotalCards() + cpu.TotalCards(); held > 0 {
			youPct = int(math.Round(float64(human.TotalCards()) / float64(held) * 100))
		}
		b.WriteString(i18n.Tf("beggarmyneighbour.share",
			"you", strconv.Itoa(youPct),
			"cpu", strconv.Itoa(100-youPct)) + "\n")

		switch g.GetPhase() {
		case domain.BeggarMyNeighbourPhasePlay:
			b.WriteString(i18n.T("beggarmyneighbour.promptPlay") + "\n")
		case domain.BeggarMyNeighbourPhasePayPenalty:
			b.WriteString(color.Yellow(i18n.Tf("beggarmyneighbour.promptPayPenalty",
				"remaining", strconv.Itoa(g.GetPenaltyRemaining()))) + "\n")
		case domain.BeggarMyNeighbourPhaseCollect:
			b.WriteString(color.Green(i18n.T("beggarmyneighbour.promptCollect")) + "\n")
		}

		if g.GetGameEndFlag() {
			switch g.GetWinnerIdx() {
			case 0:
				b.WriteString(color.Green(i18n.T("beggarmyneighbour.gameWinHuman")) + "\n")
			case 1:
				b.WriteString(color.Red(i18n.T("beggarmyneighbour.gameWinCpu")) + "\n")
			default:
				b.WriteString(i18n.T("beggarmyneighbour.gameDraw") + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BeggarMyNeighbourCuiPresenter) ActionLogOutput(g interfaces.BeggarMyNeighbourGame) string {
	return actionLogOutputText(g)
}
