//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tutePlayerStr returns the display string for a single Tute player.
// tuteSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func tuteSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return []string{"", "♠", "♣", "♥", "♦"}[suit]
}

func tutePlayerStr(g interfaces.TuteGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	teamScores := g.GetTeamScores()
	team := domain.TuteTeamOf(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("tute.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"teamScore", strconv.Itoa(teamScores[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// TuteCuiPresenter renders the Tute CUI view.
type TuteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TuteCuiPresenter) Output(g interfaces.TuteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tute.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tute.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", tuteSuitSymbol(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tutePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerTeam := g.GetWinnerTeam()
			var winnerStr string
			if winnerTeam >= 0 {
				winnerStr = strconv.Itoa(winnerTeam)
			}
			banner := i18n.Tf("tute.gameEnd", "team", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TutePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tute.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tute.promptPlayHelp") + "\n")
			if g.CanHumanDeclareMarriage() {
				b.WriteString(i18n.T("tute.promptMarriage") + "\n")
				// List which suits' K+Q can be declared (trump marked), so the
				// human need not scan the hand and recall what is already claimed.
				if suits := g.GetHumanDeclarableMarriageSuits(); len(suits) > 0 {
					trump := g.GetTrumpSuit()
					labels := make([]string, len(suits))
					for i, suit := range suits {
						label := tuteSuitSymbol(suit)
						if suit == trump {
							label += i18n.T("tute.marriageTrumpMark")
						}
						labels[i] = label
					}
					b.WriteString(i18n.Tf("tute.promptMarriageSuits",
						"suits", strings.Join(labels, ", ")) + "\n")
				}
			}
			if g.CanHumanDeclareTute() {
				b.WriteString(i18n.T("tute.promptTute") + "\n")
			}
		case domain.TutePhaseTrickEnd:
			b.WriteString(i18n.T("tute.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("tute.promptTrickEndHelp") + "\n")
		case domain.TutePhaseRoundEnd:
			pts := g.GetRoundTeamPoints()
			b.WriteString(i18n.Tf("tute.promptRoundEnd",
				"ptsA", strconv.Itoa(pts[0]),
				"ptsB", strconv.Itoa(pts[1])) + "\n")
			b.WriteString(i18n.T("tute.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Tute hint.
func (p *TuteCuiPresenter) HintOutput(g interfaces.TuteGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tute.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tuteHintReasonKeys)
	if hint.Marriage > 0 {
		return color.Yellow(i18n.Tf("tute.hintMarriage",
			"suit", strconv.Itoa(hint.Marriage),
			"reason", reason)) + "\n"
	}
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("tute.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("tute.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// tuteHintReasonKeys maps Tute-specific hint-reason identifiers to i18n keys.
var tuteHintReasonKeys = map[string]string{
	"lead_low":         "tute.hintReasonLeadLow",
	"follow_win":       "tute.hintReasonFollowWin",
	"follow_duck":      "tute.hintReasonFollowDuck",
	"discard_low":      "tute.hintReasonDiscardLow",
	"declare_marriage": "tute.hintReasonDeclareMarriage",
	"declare_tute":     "tute.hintReasonDeclareTute",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TuteCuiPresenter) ActionLogOutput(g interfaces.TuteGame) string {
	return actionLogOutputText(g)
}
