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

// preferenceBidName maps a bid constant (0-4) to its localized contract name.
func preferenceBidName(bid int) string {
	switch domain.PreferenceBid(bid) {
	case domain.PreferenceBidSix:
		return i18n.T("preference.bid.six")
	case domain.PreferenceBidMisere:
		return i18n.T("preference.bid.misere")
	case domain.PreferenceBidSeven:
		return i18n.T("preference.bid.seven")
	case domain.PreferenceBidEight:
		return i18n.T("preference.bid.eight")
	default:
		return i18n.T("preference.bid.pass")
	}
}

// preferenceTrumpStr renders the trump glyph, or a "no trump" label when none.
func preferenceTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("preference.noTrump")
	}
	return cuiSuitName(suit)
}

// preferencePlayerStr returns the display string for a single player.
func preferencePlayerStr(g interfaces.PreferenceGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("preference.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("preference.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("preference.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// PreferenceCuiPresenter renders the Préférence CUI view.
type PreferenceCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PreferenceCuiPresenter) Output(g interfaces.PreferenceGame, lastErr error) string {
	return buildCuiOutput(i18n.T("preference.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("preference.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", preferenceTrumpStr(g.GetTrumpSuit())) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("preference.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"contract", preferenceBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(preferencePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("preference.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *PreferenceCuiPresenter) writePrompt(b *strings.Builder, g interfaces.PreferenceGame) {
	switch g.GetPhase() {
	case domain.PreferencePhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("preference.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"contract", preferenceBidName(int(g.GetContract()))) + "\n")
		b.WriteString(i18n.T("preference.promptBidHelp") + "\n")
	case domain.PreferencePhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("preference.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("preference.promptPlayHelp") + "\n")
	case domain.PreferencePhaseTrickEnd:
		b.WriteString(i18n.T("preference.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("preference.promptTrickEndHelp") + "\n")
	case domain.PreferencePhaseRoundEnd:
		b.WriteString(i18n.T("preference.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("preference.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends the declarer's contract outcome and a one-line
// trick tally for every player, matching the information the Web view already
// shows in its round-result block.
func (p *PreferenceCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.PreferenceGame) {
	declIdx := g.GetDeclarerIdx()
	if declIdx < 0 {
		return
	}
	decl := g.GetPlayer(declIdx)
	if decl == nil {
		return
	}
	contract := g.GetContract()
	declTricks := decl.GetTrickCount()
	// Six/Seven/Eight need at least the target tricks; Misère needs exactly zero.
	achieved := declTricks >= preferenceContractTarget(contract)
	if contract == domain.PreferenceBidMisere {
		achieved = declTricks == 0
	}
	outcome := i18n.T("preference.contractFailed")
	if achieved {
		outcome = i18n.T("preference.contractAchieved")
	}
	b.WriteString(i18n.Tf("preference.promptRoundEndResult",
		"name", cuiPlayerName(decl, declIdx),
		"contract", preferenceBidName(int(contract)),
		"tricks", strconv.Itoa(declTricks),
		"outcome", outcome) + "\n")

	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("preference.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("preference.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")
}

// preferenceContractTarget returns the number of tricks a Six/Seven/Eight
// contract requires (0 for Misère / Pass, which are handled separately).
func preferenceContractTarget(bid domain.PreferenceBid) int {
	switch bid {
	case domain.PreferenceBidSix:
		return 6
	case domain.PreferenceBidSeven:
		return 7
	case domain.PreferenceBidEight:
		return 8
	default:
		return 0
	}
}

// HintOutput emits the current Préférence hint.
func (p *PreferenceCuiPresenter) HintOutput(g interfaces.PreferenceGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("preference.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, preferenceHintReasonKeys)
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
		return color.Yellow(i18n.Tf("preference.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("preference.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// preferenceHintReasonKeys maps Préférence-specific hint-reason identifiers to i18n keys.
var preferenceHintReasonKeys = map[string]string{
	"lead_low":    "preference.hintReasonLeadLow",
	"lead_high":   "preference.hintReasonLeadHigh",
	"follow_win":  "preference.hintReasonFollowWin",
	"follow_duck": "preference.hintReasonFollowDuck",
	"discard_low": "preference.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PreferenceCuiPresenter) ActionLogOutput(g interfaces.PreferenceGame) string {
	return actionLogOutputText(g)
}
