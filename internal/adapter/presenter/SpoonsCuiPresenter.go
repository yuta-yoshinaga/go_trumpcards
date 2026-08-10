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
				// **同ランクの組が最大の判断材料。**Web は色付きリングで囲み、
				// 3 枚以上は脈ありとして強調するのに、CUI は素の一覧しか出さず、
				// パスする札を暗算させていた (#4889)。
				b.WriteString("  " + spoonsGroupedHandStr(player) + "\n")
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

// spoonsGroupedHandStr は手札をインデックス付きで並べ、同ランクが 2 枚以上ある
// 札に枚数を添える。3 枚以上（フォーカード目前）は強調する。
func spoonsGroupedHandStr(player cuiCardList) string {
	counts := map[int]int{}
	for i := range player.GetCardsSize() {
		if c := player.GetCard(i); c != nil {
			counts[c.GetValue()]++
		}
	}
	// 索引と区切りは他ゲームと同じ formatCardList に任せ、注記だけを足す
	// (OmbreCuiPresenter と同じ形)。
	return formatCardList(player, func(c *domain.Card) string {
		out := cuiCardStr(c)
		if c == nil {
			return out
		}
		n := counts[c.GetValue()]
		if n < 2 {
			return out
		}
		tag := i18n.Tf("spoons.rankGroup", "count", strconv.Itoa(n))
		// **3 枚は 1 枚違い。**2 枚と同じ見た目だと、脈があることが伝わらない。
		if n >= 3 {
			tag = color.Yellow(tag)
		}
		return out + tag
	}, "  ", true)
}
