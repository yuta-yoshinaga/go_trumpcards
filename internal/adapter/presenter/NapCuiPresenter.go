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

// napBidName maps a bid constant (0/2/3/4/5) to its localized contract name.
func napBidName(bid int) string {
	switch domain.NapBid(bid) {
	case domain.NapBidTwo:
		return i18n.T("nap.bid.two")
	case domain.NapBidThree:
		return i18n.T("nap.bid.three")
	case domain.NapBidFour:
		return i18n.T("nap.bid.four")
	case domain.NapBidNap:
		return i18n.T("nap.bid.nap")
	default:
		return i18n.T("nap.bid.pass")
	}
}

// napTrumpStr renders the trump glyph, or a "no trump" label when none.
func napTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("nap.noTrump")
	}
	return cuiSuitName(suit)
}

// napPlayerStr returns the display string for a single player.
func napPlayerStr(g interfaces.NapGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("nap.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("nap.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("nap.playerLine",
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

// NapCuiPresenter renders the Nap CUI view.
type NapCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *NapCuiPresenter) Output(g interfaces.NapGame, lastErr error) string {
	return buildCuiOutput(i18n.T("nap.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("nap.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", napTrumpStr(g.GetTrumpSuit())) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("nap.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"contract", napBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(napPlayerStr(g, i))
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
			banner := i18n.Tf("nap.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writeNapDeclarerProgress は宣言者の契約達成状況を1行で書く。
//
// **CUI は宣言者が何トリック取ったかを一切知らせていなかった (#4763)。**
// Web は nap-declarer-progress で常時出しているのに、CLI プレイヤーは自分で
// トリック数を数えるしかなかった。
func writeNapDeclarerProgress(b *strings.Builder, g interfaces.NapGame) {
	pr := g.GetDeclarerProgress()
	if pr == nil {
		return
	}
	line := i18n.Tf("nap.declarerProgress",
		"won", strconv.Itoa(pr.Won),
		"needed", strconv.Itoa(pr.Needed),
		"remaining", strconv.Itoa(pr.Remaining))
	// **もう届かないなら押す意味が変わる。**同じ文言だと区別が付かない。
	if pr.Unreachable {
		b.WriteString(color.BoldYellow(line+i18n.T("nap.contractUnreachable")) + "\n")
		return
	}
	b.WriteString(line + "\n")
}

// writePrompt// writePrompt renders the phase-specific prompt block.
func (p *NapCuiPresenter) writePrompt(b *strings.Builder, g interfaces.NapGame) {
	switch g.GetPhase() {
	case domain.NapPhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("nap.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"contract", napBidName(int(g.GetContract()))) + "\n")
		// Name who holds the current high bid (matching the web UI's bidHighest
		// line); before anyone bids there is no holder, so the line is omitted.
		if declIdx := g.GetDeclarerIdx(); declIdx >= 0 {
			b.WriteString(i18n.Tf("nap.promptBidHolder",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx)) + "\n")
		}
		b.WriteString(i18n.T("nap.promptBidHelp") + "\n")
	case domain.NapPhasePlay:
		writeNapDeclarerProgress(b, g)
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("nap.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("nap.promptPlayHelp") + "\n")
	case domain.NapPhaseTrickEnd:
		writeNapDeclarerProgress(b, g)
		b.WriteString(i18n.T("nap.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("nap.promptTrickEndHelp") + "\n")
	case domain.NapPhaseRoundEnd:
		b.WriteString(i18n.T("nap.promptRoundEnd") + "\n")
		b.WriteString(i18n.T("nap.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current Nap hint.
func (p *NapCuiPresenter) HintOutput(g interfaces.NapGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("nap.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, napHintReasonKeys)
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
		return color.Yellow(i18n.Tf("nap.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("nap.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// napHintReasonKeys maps Nap-specific hint-reason identifiers to i18n keys.
var napHintReasonKeys = map[string]string{
	"lead_high":   "nap.hintReasonLeadHigh",
	"follow_win":  "nap.hintReasonFollowWin",
	"follow_duck": "nap.hintReasonFollowDuck",
	"discard_low": "nap.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *NapCuiPresenter) ActionLogOutput(g interfaces.NapGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
