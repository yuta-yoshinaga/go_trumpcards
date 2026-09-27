//go:build !js || !wasm || extra7

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
		b.WriteString(i18n.T("scopone.promptScore") + "\n")
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

// HintOutput recommends a capture for the human's turn, then lists every
// capture available for each hand card and whether capture is mandatory.
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
	var b strings.Builder
	anyCapture := false
	for i := 0; i < player.GetCardsSize(); i++ {
		captures := sg.GetValidCaptures(i)
		if len(captures) == 0 {
			b.WriteString(i18n.Tf("scopone.hintHandNoCapture", "hand", strconv.Itoa(i)) + "\n")
			continue
		}
		anyCapture = true
		for _, cap := range captures {
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
		options := make([]string, 0, len(captures))
		for _, capture := range captures {
			cards := make([]*domain.Card, 0, len(capture))
			for _, idx := range capture {
				cards = append(cards, table[idx])
			}
			option := cuiCardSliceStr(cards)
			if len(table) > 0 && len(capture) == len(table) {
				option += i18n.T("scopone.hintScopaSuffix")
			}
			options = append(options, option)
		}
		b.WriteString(i18n.Tf("scopone.hintHandCaptures",
			"hand", strconv.Itoa(i),
			"played", cuiCardSliceStr([]*domain.Card{player.GetCard(i)}),
			"captured", strings.Join(options, i18n.T("scopone.listSeparator"))) + "\n")
	}
	if !anyCapture {
		return color.Yellow(i18n.T("scopone.hintNoCapture")) + "\n" + b.String()
	}
	capCards := make([]*domain.Card, 0, len(bestCap))
	for _, idx := range bestCap {
		capCards = append(capCards, table[idx])
	}
	key := "scopone.hintCapture"
	if bestScopa {
		key = "scopone.hintScopa"
	}
	recommendation := color.Yellow(i18n.Tf(key,
		"played", cuiCardSliceStr([]*domain.Card{player.GetCard(bestHand)}),
		"captured", cuiCardSliceStr(capCards))) + "\n"
	return recommendation + color.Yellow(i18n.T("scopone.hintCaptureRequired")) + "\n" + b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScoponeCuiPresenter) ActionLogOutput(sg interfaces.ScoponeGame) string {
	return actionLogOutputTextForSeats[*domain.ScopaPlayer](sg)
}
