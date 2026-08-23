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

// brusquembillePlayerStr returns the display string for a single Brusquembille player.
func brusquembillePlayerStr(player *domain.BrusquembillePlayer, idx, points int, legal []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("brusquembille.playerLine",
		"name", cuiPlayerName(player, idx),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(points),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **後半は出せない札がある。** 山札が尽きると追従義務が生まれるので、
		// 素の一覧だと毎ターン手札とリードスートを暗算することになる。
		// 合法手が手札全部なら印は付けない (前半は全部合法なので無意味)。
		if len(legal) > 0 && len(legal) < player.GetCardsSize() {
			b.WriteString(cuiIndexMarkedCardListStr(player, legal, CuiLegalMark) + "\n")
			b.WriteString(i18n.T("brusquembille.followLegend") + "\n")
		} else {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// BrusquembilleCuiPresenter renders the Brusquembille CUI view.
type BrusquembilleCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BrusquembilleCuiPresenter) Output(b interfaces.BrusquembilleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("brusquembille.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("brusquembille.header",
			"trick", strconv.Itoa(b.GetTrickNumber()),
			"stock", strconv.Itoa(b.GetStockRemaining())) + "\n")

		if tc := b.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("brusquembille.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.T("brusquembille.trumpLineNone") + "\n")
		}
		// **全席ぶん並べる。** 席 0 と席 1 だけだと、3〜5 人卓で残りが消える。
		sb.WriteString(i18n.Tf("brusquembille.pointsLine",
			"scores", brusquembilleCuiScoreSummary(b)) + "\n")

		for i := 0; i < b.GetPlayerCnt(); i++ {
			legal := []int(nil)
			if i == 0 && b.IsFollowRequired() {
				legal = b.GetValidPlayIndices(0)
			}
			sb.WriteString(brusquembillePlayerStr(b.GetPlayer(i), i, b.GetPlayerPoints(i), legal))
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
			scores := brusquembilleCuiScoreSummary(b)
			var banner string
			switch w := b.GetWinnerIdx(); {
			case w == 0:
				banner = i18n.Tf("brusquembille.gameEndP0", "scores", scores)
			case w > 0:
				banner = i18n.Tf("brusquembille.gameEndCpu", "seat", strconv.Itoa(w), "scores", scores)
			default:
				banner = i18n.Tf("brusquembille.gameEndTie", "scores", scores)
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}
		switch b.GetPhase() {
		case domain.BrusquembillePhasePlay:
			currentIdx := b.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("brusquembille.promptCurrentPlayer",
				"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("brusquembille.promptPlay") + "\n")
		case domain.BrusquembillePhaseTrickEnd:
			sb.WriteString(i18n.T("brusquembille.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("brusquembille.promptTrickEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Brusquembille hint.
func (p *BrusquembilleCuiPresenter) HintOutput(b interfaces.BrusquembilleGame) string {
	hint := b.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("brusquembille.hintNone") + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("brusquembille.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, brusquembilleHintReasonKeys))) + "\n"
}

// brusquembilleHintReasonKeys maps Brusquembille-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var brusquembilleHintReasonKeys = map[string]string{
	"lead_trump":  "brusquembille.hintReasonLeadTrump",
	"lead_low":    "brusquembille.hintReasonLeadLow",
	"lead_value":  "brusquembille.hintReasonLeadValue",
	"follow_cut":  "brusquembille.hintReasonFollowCut",
	"follow_win":  "brusquembille.hintReasonFollowWin",
	"follow_dump": "brusquembille.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BrusquembilleCuiPresenter) ActionLogOutput(b interfaces.BrusquembilleGame) string {
	return actionLogOutputTextForSeats[*domain.BrusquembillePlayer](b)
}

// brusquembilleCuiScoreSummary は全席の得点を "12-34-56" の形で返す。
func brusquembilleCuiScoreSummary(b interfaces.BrusquembilleGame) string {
	parts := make([]string, 0, b.GetPlayerCnt())
	for i := 0; i < b.GetPlayerCnt(); i++ {
		parts = append(parts, strconv.Itoa(b.GetPlayerPoints(i)))
	}
	return strings.Join(parts, "-")
}
