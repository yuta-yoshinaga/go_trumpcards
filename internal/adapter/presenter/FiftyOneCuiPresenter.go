package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FiftyOneCuiPresenter renders the Fifty-one CUI view.
type FiftyOneCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FiftyOneCuiPresenter) Output(fo interfaces.FiftyOneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fiftyone.outputTitle"), func(b *strings.Builder) {
		// Per-player info
		for i := 0; i < fo.GetPlayerCnt(); i++ {
			player := fo.GetPlayer(i)
			b.WriteString(cuiPlayerName(player, i))
			score := player.BestSuitScore()
			if player.GetIsHuman() {
				b.WriteString(i18n.Tf("fiftyone.humanScoreLine",
					"score", strconv.Itoa(score)) + "\n")
				b.WriteString(cuiIndexedCardListStr(player) + "\n")
				b.WriteString(fiftyOneSuitScoreLine(player) + "\n")
			} else {
				scoreStr := i18n.T("fiftyone.scoreUnknown")
				if fo.GetGameEndFlag() {
					scoreStr = strconv.Itoa(score)
				}
				b.WriteString(i18n.Tf("fiftyone.cpuScoreLine",
					"count", strconv.Itoa(player.GetCardsSize()),
					"score", scoreStr) + "\n")
			}
		}

		// Table cards
		b.WriteString("----------\n")
		b.WriteString(i18n.T("fiftyone.tableHeader"))
		tableCards := fo.GetTableCards()
		for i, c := range tableCards {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(i18n.Tf("fiftyone.tableCard",
				"idx", strconv.Itoa(i),
				"card", cuiCardStr(c)))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Colored so CUI players notice the final round is in effect (every phase).
		if idx := fo.GetStopCallerIdx(); idx >= 0 {
			if caller := fo.GetPlayer(idx); caller != nil {
				name := cuiPlayerName(caller, idx)
				b.WriteString(color.Red(i18n.Tf("fiftyone.cuiStopCalled", "name", name)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if fo.GetGameEndFlag() {
			b.WriteString(i18n.T("fiftyone.gameEndHeader") + "\n")
			for i := 0; i < fo.GetPlayerCnt(); i++ {
				name := cuiPlayerName(fo.GetPlayer(i), i)
				b.WriteString(i18n.Tf("fiftyone.scoreEntry",
					"name", name,
					"score", strconv.Itoa(fo.GetPlayer(i).BestSuitScore())) + "\n")
			}
			winnerIdx := fo.GetWinnerIdx()
			winner := fo.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString(i18n.T("fiftyone.winHuman") + "\n")
			} else {
				b.WriteString(i18n.Tf("fiftyone.winCpu",
					"idx", strconv.Itoa(winnerIdx)) + "\n")
			}
		} else if fo.IsHumanTurn() {
			b.WriteString(i18n.T("fiftyone.promptHumanTurn") + "\n")
		}
	})
}

// fiftyOneSuitScoreLine renders every suit total for the given player, marking
// the best one. The Web GUI shows the four badges permanently; the CUI only had
// `BestSuitScore()` as a single number, so the other three had to be added up by
// hand from the indexed hand list every turn (#4866).
func fiftyOneSuitScoreLine(player *domain.FiftyOnePlayer) string {
	scores := player.SuitScores()
	// BestSuit と同じ決定論的順序で走査する。同点のときに印が付く先が食い違うと
	// 「(スコア: n)」の n がどの行のものか分からなくなる。
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	best := player.BestSuit()
	parts := make([]string, 0, len(suits))
	for _, design := range suits {
		entry := i18n.Tf("fiftyone.suitScoreEntry",
			"suit", cuiSuitName(design), "score", strconv.Itoa(scores[design]))
		if design == best {
			entry += i18n.T("fiftyone.suitScoreBestMark")
		}
		if isRedSuit(design) {
			entry = color.Red(entry)
		}
		parts = append(parts, entry)
	}
	return i18n.Tf("fiftyone.suitScoresLine", "scores", strings.Join(parts, "  "))
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FiftyOneCuiPresenter) ActionLogOutput(fo interfaces.FiftyOneGame) string {
	return actionLogOutputText(fo)
}
