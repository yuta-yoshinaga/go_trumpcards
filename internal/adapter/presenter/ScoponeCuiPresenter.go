//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ScoponeCuiPresenter renders the Scopone CUI view.
type ScoponeCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ScoponeCuiPresenter) Output(sg interfaces.ScoponeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("scopone.helpTitle"), func(b *strings.Builder) {
		// Team scores header.
		b.WriteString(i18n.Tf("scopone.teamScoreLine",
			"t0", strconv.Itoa(sg.GetTeamScore(0)),
			"t1", strconv.Itoa(sg.GetTeamScore(1))) + "\n")
		b.WriteString(i18n.Tf("scopone.roundLine",
			"round", strconv.Itoa(sg.GetRoundNumber())) + "\n")

		for i := 0; i < sg.GetPlayerCnt(); i++ {
			b.WriteString(scoponePlayerStr(sg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := sg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("scopone.tableLine", "cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("scopone.tableEmpty") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if sg.GetGameEndFlag() {
			b.WriteString(i18n.T("scopone.gameEnd") + "\n")
			b.WriteString(i18n.Tf("scopone.winnerLine",
				"team", strconv.Itoa(sg.GetWinnerTeam())) + "\n")
			scoponeScoreDetailStr(b, sg.GetLastRoundDetail())
			return
		}

		if sg.GetPhase() == domain.ScoponePhaseRoundEnd {
			b.WriteString(i18n.T("scopone.roundEnd") + "\n")
			scoponeScoreDetailStr(b, sg.GetLastRoundDetail())
			b.WriteString(i18n.T("scopone.promptNext") + "\n")
			return
		}

		currentTurn := sg.GetCurrentTurn()
		b.WriteString(i18n.Tf("scopone.promptCurrentTurn",
			"name", cuiPlayerName(sg.GetPlayer(currentTurn), currentTurn),
			"team", strconv.Itoa(domain.ScoponeTeamOf(currentTurn))) + "\n")
		b.WriteString(i18n.T("scopone.promptHelp") + "\n")
	})
}

// scoponePlayerStr returns the display string for a single Scopone player.
func scoponePlayerStr(player *domain.ScopaPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("scopone.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.ScoponeTeamOf(i)),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"scopa", strconv.Itoa(player.GetScopaCount())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// scoponeScoreDetailStr renders the per-team score breakdown.
func scoponeScoreDetailStr(b *strings.Builder, det *domain.ScoponeScoreDetail) {
	if det == nil {
		return
	}
	for t := 0; t < domain.ScoponeTeamCnt; t++ {
		b.WriteString(i18n.Tf("scopone.scoreDetailLine",
			"team", strconv.Itoa(t),
			"cards", strconv.Itoa(det.Cards[t]),
			"diamonds", strconv.Itoa(det.Diamonds[t]),
			"sevens", strconv.Itoa(det.Sevens[t]),
			"scopas", strconv.Itoa(det.Scopas[t]),
			"gained", strconv.Itoa(det.Gained[t])) + "\n")
	}
}

// HintOutput emits a capture recommendation for the human's turn: the hand
// card and the table cards it captures (flagging a scopa when it clears the
// table), reusing the domain's GetValidCaptures.
func (p *ScoponeCuiPresenter) HintOutput(sg interfaces.ScoponeGame) string {
	if sg.GetPhase() != domain.ScoponePhasePlayerTurn {
		return i18n.T("scopone.hintNone") + "\n"
	}
	turn := sg.GetCurrentTurn()
	player := sg.GetPlayer(turn)
	if player == nil || !player.GetIsHuman() {
		return i18n.T("scopone.hintNone") + "\n"
	}
	table := sg.GetTableCards()
	bestHand := -1
	var bestCap []int
	bestScopa := false
	for i := 0; i < player.GetCardsSize(); i++ {
		for _, cap := range sg.GetValidCaptures(i) {
			isScopa := len(table) > 0 && len(cap) == len(table)
			switch {
			case bestHand == -1:
			case isScopa && !bestScopa:
			case isScopa == bestScopa && len(cap) > len(bestCap):
			default:
				continue
			}
			bestHand, bestCap, bestScopa = i, cap, isScopa
		}
	}
	if bestHand == -1 {
		return color.Yellow(i18n.T("scopone.hintNoCapture")) + "\n"
	}
	capCards := make([]*domain.Card, 0, len(bestCap))
	for _, idx := range bestCap {
		capCards = append(capCards, table[idx])
	}
	key := "scopone.hintCapture"
	if bestScopa {
		key = "scopone.hintScopa"
	}
	return color.Yellow(i18n.Tf(key,
		"played", cuiCardSliceStr([]*domain.Card{player.GetCard(bestHand)}),
		"captured", cuiCardSliceStr(capCards))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScoponeCuiPresenter) ActionLogOutput(sg interfaces.ScoponeGame) string {
	return actionLogOutputTextForSeats[*domain.ScopaPlayer](sg)
}
