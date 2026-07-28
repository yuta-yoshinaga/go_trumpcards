//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SpoonsCuiPresenter renders the Spoons CUI view.
type SpoonsCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SpoonsCuiPresenter) Output(g interfaces.SpoonsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spoons.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("spoons.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"spoons", strconv.Itoa(g.GetSpoonsRemaining())) + "\n")
		b.WriteString("----------\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			if player == nil {
				continue
			}
			name := cuiPlayerName(player, i)
			letters := i18n.Tf("spoons.lettersLabel", "letters", spoonsLetters(player.GetLetters()))
			if player.GetEliminated() {
				b.WriteString(color.Red(i18n.Tf("spoons.playerEliminated", "name", name)) + "\n")
				continue
			}
			spoon := ""
			if player.GetHasSpoon() {
				spoon = " " + color.Green(i18n.T("spoons.hasSpoon"))
			}
			if i == 0 {
				// 人間の手札のみ公開表示する。
				b.WriteString(name + " " + letters + spoon + "\n")
				b.WriteString("  " + cuiIndexedCardListStr(player) + "\n")
			} else {
				b.WriteString(i18n.Tf("spoons.cpuHandLine",
					"name", name,
					"count", strconv.Itoa(player.GetCardsSize())) + " " + letters + spoon + "\n")
			}
		}
		b.WriteString("----------\n")

		switch g.GetPhase() {
		case domain.SpoonsPhaseGrab:
			b.WriteString(color.Yellow(i18n.T("spoons.promptGrab")) + "\n")
		case domain.SpoonsPhasePass:
			if g.IsHumanTurn() {
				b.WriteString(i18n.T("spoons.promptPass") + "\n")
			} else {
				b.WriteString(i18n.T("spoons.promptCpuTurn") + "\n")
			}
		case domain.SpoonsPhaseRoundEnd:
			b.WriteString(i18n.T("spoons.promptNextRound") + "\n")
		}

		if loser := g.GetRoundLoserIdx(); loser >= 0 && g.GetPhase() == domain.SpoonsPhaseRoundEnd {
			lp := g.GetPlayer(loser)
			b.WriteString(color.Red(i18n.Tf("spoons.roundLoser",
				"name", cuiPlayerName(lp, loser))) + "\n")
		}

		if g.GetGameEndFlag() {
			if g.GetWinnerIdx() == 0 {
				b.WriteString(color.Green(i18n.T("spoons.winHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.Tf("spoons.winCpu",
					"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpoonsCuiPresenter) ActionLogOutput(g interfaces.SpoonsGame) string {
	return actionLogOutputText(g)
}

// spoonsLetters は文字数を "S-P-O-O-N-S" の取得済み接頭辞として表示する。
func spoonsLetters(n int) string {
	letters := []string{"S", "P", "O", "O", "N", "S"}
	if n <= 0 {
		return "-"
	}
	if n > len(letters) {
		n = len(letters)
	}
	return strings.Join(letters[:n], "")
}
