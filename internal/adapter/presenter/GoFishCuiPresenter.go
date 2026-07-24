package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// goFishPlayerStr returns the display string for a single GoFish player.
func goFishPlayerStr(player *domain.GoFishPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("gofish.playerLine",
		"name", cuiPlayerName(player, i),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"books", strconv.Itoa(player.GetBookCount())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// writeGoFishKnownRanks lists, per opponent, the ranks they are known to hold
// from past asks; a rank the human also holds is starred as a strong ask
// target. Emitted only on the human's turn.
func writeGoFishKnownRanks(b *strings.Builder, gf interfaces.GoFishGame) {
	known := gf.GetKnownRanks()
	var human *domain.GoFishPlayer
	for i := 0; i < gf.GetPlayerCnt(); i++ {
		if p := gf.GetPlayer(i); p != nil && p.GetIsHuman() {
			human = p
			break
		}
	}
	wrote := false
	for i := 0; i < gf.GetPlayerCnt(); i++ {
		p := gf.GetPlayer(i)
		if p == nil || p.GetIsHuman() {
			continue
		}
		ranks := known[i]
		if len(ranks) == 0 {
			continue
		}
		parts := make([]string, len(ranks))
		for k, r := range ranks {
			s := cuiRankLabel(r)
			if human != nil && human.HasRank(r) {
				s += "*" // you hold this rank too — a strong ask target
			}
			parts[k] = s
		}
		b.WriteString(i18n.Tf("gofish.knownRanks",
			"name", cuiPlayerName(p, i),
			"ranks", strings.Join(parts, " ")) + "\n")
		wrote = true
	}
	if wrote {
		b.WriteString(i18n.T("gofish.knownRanksLegend") + "\n")
	}
}

// GoFishCuiPresenter renders the Go Fish CUI view.
type GoFishCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *GoFishCuiPresenter) Output(gf interfaces.GoFishGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gofish.outputTitle"), func(b *strings.Builder) {
		for i := 0; i < gf.GetPlayerCnt(); i++ {
			b.WriteString(goFishPlayerStr(gf.GetPlayer(i), i))
		}

		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("gofish.deckLine",
			"count", strconv.Itoa(gf.GetDeckRemaining())) + "\n")

		// Last ask result
		if gf.GetLastAskPlayerIdx() >= 0 {
			askerName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskPlayerIdx()), gf.GetLastAskPlayerIdx())
			targetName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskTargetIdx()), gf.GetLastAskTargetIdx())
			rankStr := cuiRankLabel(gf.GetLastAskRank())
			if gf.GetLastAskSuccess() {
				b.WriteString(i18n.Tf("gofish.askSuccess",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr,
					"count", strconv.Itoa(len(gf.GetLastCardsReceived()))) + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.askFail",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr) + "\n")
			}
			if gf.GetLastBookFormed() {
				b.WriteString(i18n.Tf("gofish.bookFormed",
					"rank", cuiRankLabel(gf.GetLastBookRank())) + "\n")
			}
		}

		// CPU action history
		for _, action := range gf.GetCpuActions() {
			askerName := cuiPlayerName(gf.GetPlayer(action.AskPlayerIdx), action.AskPlayerIdx)
			targetName := cuiPlayerName(gf.GetPlayer(action.AskTargetIdx), action.AskTargetIdx)
			rankStr := cuiRankLabel(action.AskRank)
			if action.Success {
				b.WriteString(i18n.Tf("gofish.cpuAskSuccess",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr,
					"count", strconv.Itoa(action.CardsReceived)) + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.cpuAskFail",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr) + "\n")
			}
			if action.BookFormed {
				b.WriteString(i18n.Tf("gofish.cpuBookFormed",
					"asker", askerName,
					"rank", cuiRankLabel(action.BookRank)) + "\n")
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if gf.GetGameEndFlag() {
			winnerIdx := gf.GetWinnerIdx()
			winner := gf.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString(i18n.T("gofish.winHuman") + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.winCpu",
					"idx", strconv.Itoa(winnerIdx)) + "\n")
			}
			for i := 0; i < gf.GetPlayerCnt(); i++ {
				b.WriteString(i18n.Tf("gofish.scoreEntry",
					"name", cuiPlayerName(gf.GetPlayer(i), i),
					"count", strconv.Itoa(gf.GetPlayer(i).GetBookCount())) + "\n")
			}
			return
		}
		turnName := cuiPlayerName(gf.GetPlayer(gf.GetCurrentTurn()), gf.GetCurrentTurn())
		b.WriteString(i18n.Tf("gofish.promptCurrentTurn", "name", turnName) + "\n")
		if gf.IsHumanTurn() {
			b.WriteString(i18n.T("gofish.promptHumanHelp") + "\n")
			writeGoFishKnownRanks(b, gf)
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GoFishCuiPresenter) ActionLogOutput(gf interfaces.GoFishGame) string {
	return actionLogOutputText(gf)
}
