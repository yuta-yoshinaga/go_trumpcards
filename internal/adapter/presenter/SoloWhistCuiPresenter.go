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

// soloWhistBidName maps a bid constant (0-3) to its localized contract name.
func soloWhistBidName(bid int) string {
	switch domain.SoloWhistBid(bid) {
	case domain.SoloWhistBidSolo:
		return i18n.T("solowhist.bid.solo")
	case domain.SoloWhistBidMisere:
		return i18n.T("solowhist.bid.misere")
	case domain.SoloWhistBidAbundance:
		return i18n.T("solowhist.bid.abundance")
	default:
		return i18n.T("solowhist.bid.pass")
	}
}

// soloWhistTrumpStr renders the trump glyph, or a "no trump" label when none.
func soloWhistTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("solowhist.noTrump")
	}
	return cuiSuitName(suit)
}

// soloWhistPlayerStr returns the display string for a single player.
func soloWhistPlayerStr(g interfaces.SoloWhistGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("solowhist.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("solowhist.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("solowhist.playerLine",
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

// SoloWhistCuiPresenter renders the Solo Whist CUI view.
type SoloWhistCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SoloWhistCuiPresenter) Output(g interfaces.SoloWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("solowhist.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("solowhist.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", soloWhistTrumpStr(g.GetTrumpSuit())) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("solowhist.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"contract", soloWhistBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(soloWhistPlayerStr(g, i))
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
			banner := i18n.Tf("solowhist.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *SoloWhistCuiPresenter) writePrompt(b *strings.Builder, g interfaces.SoloWhistGame) {
	switch g.GetPhase() {
	case domain.SoloWhistPhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("solowhist.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"contract", soloWhistBidName(int(g.GetContract()))) + "\n")
		// List every player's declared bid (pass included) so later bidders can see
		// the auction; players who have not yet bid show "-".
		bids := g.GetBids()
		bidDone := g.GetBidDone()
		entries := make([]string, g.GetPlayerCnt())
		for i := 0; i < g.GetPlayerCnt(); i++ {
			state := "-"
			if bidDone[i] {
				state = soloWhistBidName(int(bids[i]))
			}
			entries[i] = cuiPlayerName(g.GetPlayer(i), i) + "=" + state
		}
		b.WriteString(i18n.Tf("solowhist.bidHistory",
			"bids", strings.Join(entries, ", ")) + "\n")
		b.WriteString(i18n.T("solowhist.promptBidHelp") + "\n")
	case domain.SoloWhistPhasePlay:
		writeSoloWhistDeclarerProgress(b, g)
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("solowhist.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("solowhist.promptPlayHelp") + "\n")
	case domain.SoloWhistPhaseTrickEnd:
		writeSoloWhistDeclarerProgress(b, g)
		b.WriteString(i18n.T("solowhist.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("solowhist.promptTrickEndHelp") + "\n")
	case domain.SoloWhistPhaseRoundEnd:
		b.WriteString(i18n.T("solowhist.promptRoundEnd") + "\n")
		b.WriteString(i18n.T("solowhist.promptRoundEndHelp") + "\n")
	}
}

// writeSoloWhistDeclarerProgress は宣言者の契約達成状況を1行で書く。
// 宣言者が決まっていない、プレイ中以外のフェーズでは何も書かない。
//
// **ミゼールは1トリック取った瞬間に失敗が確定する** (#5649)。Web は
// solowhist-contract-progress で常時出しているのに、CUI はラウンドが終わるまで
// 何も言わず、決着済みのラウンドを最後まで打たせていた。
func writeSoloWhistDeclarerProgress(b *strings.Builder, g interfaces.SoloWhistGame) {
	pr := g.GetDeclarerProgress()
	if pr == nil {
		return
	}
	// ミゼールの必要トリックは 0 なので「0/0」と出しても意味を成さない。
	// 「1つでも取ると失敗」という規則そのものを書く。
	line := i18n.Tf("solowhist.declarerProgress",
		"won", strconv.Itoa(pr.Won),
		"needed", strconv.Itoa(pr.Needed),
		"remaining", strconv.Itoa(pr.Remaining))
	if pr.IsMisere {
		line = i18n.Tf("solowhist.misereProgress",
			"won", strconv.Itoa(pr.Won),
			"remaining", strconv.Itoa(pr.Remaining))
	}
	switch {
	case pr.Unreachable:
		b.WriteString(color.BoldYellow(line+i18n.T("solowhist.contractUnreachable")) + "\n")
	case pr.Made:
		b.WriteString(color.Green(line+i18n.T("solowhist.contractMade")) + "\n")
	default:
		b.WriteString(line + "\n")
	}
}

// HintOutput emits the current Solo Whist hint.
func (p *SoloWhistCuiPresenter) HintOutput(g interfaces.SoloWhistGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("solowhist.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, soloWhistHintReasonKeys)
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
		return color.Yellow(i18n.Tf("solowhist.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("solowhist.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// soloWhistHintReasonKeys maps Solo-Whist-specific hint-reason identifiers to i18n keys.
var soloWhistHintReasonKeys = map[string]string{
	"lead_low":    "solowhist.hintReasonLeadLow",
	"follow_win":  "solowhist.hintReasonFollowWin",
	"follow_duck": "solowhist.hintReasonFollowDuck",
	"discard_low": "solowhist.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SoloWhistCuiPresenter) ActionLogOutput(g interfaces.SoloWhistGame) string {
	return actionLogOutputText(g)
}
