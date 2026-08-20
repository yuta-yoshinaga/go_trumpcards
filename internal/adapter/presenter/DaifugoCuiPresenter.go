package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// daifugoPlayerStr returns the display string for a single Daifugo player.
//
// playable が非 nil のとき、そのインデックスのカードに "*" を付ける。大富豪は
// 革命・11バック・スートロック・階段縛りで場ごとに強弱と枚数の条件が変わるのに、
// CUI はどれが出せるかを自力で計算させていた (#4733)。CrazyEights の
// crazyEightsHandStr と同じ見せ方に揃える。
func daifugoPlayerStr(player *domain.DaifugoPlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("daifugo.playerFinished", "rank", daifugoRankName(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("daifugo.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(daifugoHandStr(player, playable) + "\n")
		}
	}
	return b.String()
}

// daifugoHandStr renders the hand as an indexed list, starring the cards that
// take part in at least one legal play right now.
func daifugoHandStr(player *domain.DaifugoPlayer, playable []int) string {
	if len(playable) == 0 {
		return cuiIndexedCardListStr(player)
	}
	mark := make(map[int]bool, len(playable))
	for _, idx := range playable {
		mark[idx] = true
	}
	parts := make([]string, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + cuiCardStr(player.GetCard(i))
		if mark[i] {
			parts[i] += "*"
		}
	}
	return strings.Join(parts, "  ")
}

// DaifugoCuiPresenter renders the Daifugo CUI view.
type DaifugoCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *DaifugoCuiPresenter) Output(dg interfaces.DaifugoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("daifugo.helpTitle"), func(b *strings.Builder) {
		playable := dg.GetPlayableCardIndices()
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			b.WriteString(daifugoPlayerStr(dg.GetPlayer(i), i, playable))
		}

		b.WriteString("----------\n")

		// Local rule status
		if dg.GetRevolutionActive() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleRevolution")) + "\n")
		}
		if dg.GetElevenBackActive() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleElevenBack")) + "\n")
		}
		if dg.GetSuitLocked() {
			fmt.Fprintf(b, "%s%s\n",
				color.BoldYellow(i18n.T("daifugo.ruleSuitLockedPrefix")),
				cuiSuitName(dg.GetLockedSuit()))
		}
		if dg.GetTableIsSequence() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleSequence")) + "\n")
		}
		if dg.GetReverseDirection() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleNineReverse")) + "\n")
		}
		if dg.GetNumberLocked() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleNumberLocked")) + "\n")
		}
		if dg.GetSequenceLocked() {
			b.WriteString(color.BoldYellow(i18n.T("daifugo.ruleSequenceLocked")) + "\n")
		}

		// Card exchanges
		exchangeActions := dg.GetExchangeActions()
		if len(exchangeActions) > 0 {
			b.WriteString(color.Bold(i18n.T("daifugo.exchangesHeader")) + "\n")
			for _, ex := range exchangeActions {
				fmt.Fprintf(b, "%s → %s: %s\n",
					cuiPlayerName(dg.GetPlayer(ex.FromPlayerIdx), ex.FromPlayerIdx),
					cuiPlayerName(dg.GetPlayer(ex.ToPlayerIdx), ex.ToPlayerIdx),
					cuiCardSliceStr(ex.Cards))
			}
		}

		// Table cards
		tableCards := dg.GetTableCards()
		if len(tableCards) > 0 {
			lastPlayIdx := dg.GetLastPlayPlayerIdx()
			b.WriteString(i18n.Tf("daifugo.tableLine",
				"cards", cuiCardSliceStr(tableCards),
				"name", cuiPlayerName(dg.GetPlayer(lastPlayIdx), lastPlayIdx)) + "\n")
		} else {
			b.WriteString(i18n.T("daifugo.tableEmpty") + "\n")
		}

		// Human's previous action
		humanAction := dg.GetHumanAction()
		if humanAction != nil {
			actName := cuiPlayerName(dg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx)
			if len(humanAction.PlayedCards) == 0 {
				b.WriteString(i18n.Tf("daifugo.actionPassed", "name", actName) + "\n")
			} else {
				b.WriteString(i18n.Tf("daifugo.actionPlayed",
					"name", actName,
					"cards", cuiCardSliceStr(humanAction.PlayedCards)) + "\n")
			}
		}

		// CPU action history
		cpuActions := dg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("daifugo.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(dg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
				if len(action.PlayedCards) == 0 {
					b.WriteString(i18n.Tf("daifugo.actionPassed", "name", actPlayerName) + "\n")
				} else {
					b.WriteString(i18n.Tf("daifugo.actionPlayed",
						"name", actPlayerName,
						"cards", cuiCardSliceStr(action.PlayedCards)) + "\n")
				}
			}
		}

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", i18n.MarkErrorLine(color.Red(lastErr.Error())))
		}

		if dg.GetGameEndFlag() {
			b.WriteString(i18n.T("daifugo.gameEnd") + "\n")
			for i := 0; i < dg.GetPlayerCnt(); i++ {
				player := dg.GetPlayer(i)
				penalty := ""
				if player.GetIllegalFinishPenalty() {
					penalty = i18n.T("daifugo.penaltyIllegalFinish")
				}
				b.WriteString(i18n.Tf("daifugo.playerRankLine",
					"name", cuiPlayerName(dg.GetPlayer(i), i),
					"rank", daifugoRankName(player.GetRank()),
					"penalty", penalty) + "\n")
			}
			return
		}
		currentTurn := dg.GetCurrentTurn()
		currentName := cuiPlayerName(dg.GetPlayer(currentTurn), currentTurn)
		b.WriteString(i18n.Tf("daifugo.turnLine", "name", currentName) + "\n")
		switch dg.GetPendingActionType() {
		case domain.DaifugoPendingSevenPass:
			b.WriteString(i18n.T("daifugo.promptSevenPass") + "\n")
		case domain.DaifugoPendingTenDiscard:
			b.WriteString(i18n.T("daifugo.promptTenDiscard") + "\n")
		case domain.DaifugoPendingQueenBomber:
			b.WriteString(i18n.T("daifugo.promptQueenBomber") + "\n")
		default:
			b.WriteString(i18n.T("daifugo.promptPlay") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DaifugoCuiPresenter) ActionLogOutput(dg interfaces.DaifugoGame) string {
	return actionLogOutputTextForSeats[*domain.DaifugoPlayer](dg)
}
