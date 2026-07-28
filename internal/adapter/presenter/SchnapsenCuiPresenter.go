//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// schnapsenPlayerStr returns the display string for a single Schnapsen player.
func schnapsenPlayerStr(player *domain.SchnapsenPlayer, idx, points int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("schnapsen.playerLine",
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

// SchnapsenCuiPresenter renders the Schnapsen CUI view.
type SchnapsenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SchnapsenCuiPresenter) Output(s interfaces.SchnapsenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("schnapsen.helpTitle"), func(sb *strings.Builder) {
		phaseKey := "schnapsen.phaseFirst"
		if s.IsEndgame() {
			phaseKey = "schnapsen.phaseSecond"
		}
		sb.WriteString(i18n.Tf("schnapsen.header",
			"trick", strconv.Itoa(s.GetTrickNumber()),
			"stock", strconv.Itoa(s.GetStockRemaining()),
			"phase", i18n.T(phaseKey)) + "\n")

		if tc := s.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("schnapsen.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.Tf("schnapsen.trumpLineNone", "suit", schnapsenSuitName(s.GetTrumpSuit())) + "\n")
		}
		sb.WriteString(i18n.Tf("schnapsen.pointsLine",
			"p0", strconv.Itoa(s.GetPlayerPoints(0)),
			"p1", strconv.Itoa(s.GetPlayerPoints(1))) + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(schnapsenPlayerStr(s.GetPlayer(i), i, s.GetPlayerPoints(i)))
		}

		// マリアージュ宣言が可能なら案内を表示する
		if marriages := s.GetMarriageIndices(0); len(marriages) > 0 && s.GetCurrentPlayerIdx() == 0 {
			sb.WriteString(i18n.Tf("schnapsen.marriageAvailable", "indices", schnapsenJoinInts(marriages)) + "\n")
		}

		sb.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(sb, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			p0 := s.GetPlayerPoints(0)
			p1 := s.GetPlayerPoints(1)
			var banner string
			switch s.GetWinnerIdx() {
			case 0:
				banner = i18n.Tf("schnapsen.gameEndP0", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			case 1:
				banner = i18n.Tf("schnapsen.gameEndP1", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			default:
				banner = i18n.Tf("schnapsen.gameEndTie", "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}
		switch s.GetPhase() {
		case domain.SchnapsenPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("schnapsen.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("schnapsen.promptPlay") + "\n")
		case domain.SchnapsenPhaseTrickEnd:
			sb.WriteString(i18n.T("schnapsen.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("schnapsen.promptTrickEndHelp") + "\n")
		}
	})
}

// joinInts ints をカンマ区切り文字列に整形する。
func schnapsenJoinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, ", ")
}

// schnapsenSuitName スート番号を i18n のスート名に変換する。
func schnapsenSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("schnapsen.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("schnapsen.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("schnapsen.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("schnapsen.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current Schnapsen hint.
func (p *SchnapsenCuiPresenter) HintOutput(s interfaces.SchnapsenGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("schnapsen.hintNone") + "\n"
	}
	player := s.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	if hint.IsMarriage {
		return color.Yellow(i18n.Tf("schnapsen.hintMarriage",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card))) + "\n"
	}
	return color.Yellow(i18n.Tf("schnapsen.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, schnapsenHintReasonKeys))) + "\n"
}

// schnapsenHintReasonKeys maps Schnapsen-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var schnapsenHintReasonKeys = map[string]string{
	"lead_trump":  "schnapsen.hintReasonLeadTrump",
	"lead_low":    "schnapsen.hintReasonLeadLow",
	"lead_value":  "schnapsen.hintReasonLeadValue",
	"follow_cut":  "schnapsen.hintReasonFollowCut",
	"follow_win":  "schnapsen.hintReasonFollowWin",
	"follow_dump": "schnapsen.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SchnapsenCuiPresenter) ActionLogOutput(s interfaces.SchnapsenGame) string {
	return actionLogOutputText(s)
}
