package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// briscolaPlayerStr returns the display string for a single Briscola player.
func briscolaPlayerStr(player *domain.BriscolaPlayer, idx, points int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("briscola.playerLine",
		"name", cuiPlayerName(player, idx),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(points),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BriscolaCuiPresenter renders the Briscola CUI view.
type BriscolaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BriscolaCuiPresenter) Output(b interfaces.BriscolaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("briscola.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("briscola.header",
			"trick", strconv.Itoa(b.GetTrickNumber()),
			"stock", strconv.Itoa(b.GetStockRemaining())) + "\n")

		if tc := b.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("briscola.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.T("briscola.trumpLineNone") + "\n")
		}
		sb.WriteString(i18n.Tf("briscola.pointsLine",
			"p0", strconv.Itoa(b.GetPlayerPoints(0)),
			"p1", strconv.Itoa(b.GetPlayerPoints(1))) + "\n")

		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(briscolaPlayerStr(b.GetPlayer(i), i, b.GetPlayerPoints(i)))
		}

		sb.WriteString("----------\n")

		trick := b.GetCurrentTrick()
		cuiTrickBlock(sb, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if b.GetGameEndFlag() {
			p0 := b.GetPlayerPoints(0)
			p1 := b.GetPlayerPoints(1)
			var banner string
			switch b.GetWinnerIdx() {
			case 0:
				banner = i18n.Tf("briscola.gameEndP0", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			case 1:
				banner = i18n.Tf("briscola.gameEndP1", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			default:
				banner = i18n.Tf("briscola.gameEndTie", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}
		switch b.GetPhase() {
		case domain.BriscolaPhasePlay:
			currentIdx := b.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("briscola.promptCurrentPlayer",
				"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("briscola.promptPlay") + "\n")
		case domain.BriscolaPhaseTrickEnd:
			sb.WriteString(i18n.T("briscola.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("briscola.promptTrickEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Briscola hint.
func (p *BriscolaCuiPresenter) HintOutput(b interfaces.BriscolaGame) string {
	hint := b.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("briscola.hintNone") + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("briscola.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, briscolaHintReasonKeys))) + "\n"
}

// briscolaHintReasonKeys maps Briscola-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var briscolaHintReasonKeys = map[string]string{
	"lead_trump":  "briscola.hintReasonLeadTrump",
	"lead_low":    "briscola.hintReasonLeadLow",
	"lead_value":  "briscola.hintReasonLeadValue",
	"follow_cut":  "briscola.hintReasonFollowCut",
	"follow_win":  "briscola.hintReasonFollowWin",
	"follow_dump": "briscola.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BriscolaCuiPresenter) ActionLogOutput(b interfaces.BriscolaGame) string {
	return actionLogOutputTextWithNames(b, func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) })
}
