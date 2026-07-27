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

// beziquePlayerStr returns the display string for a single Bezique player.
func beziquePlayerStr(player *domain.BeziquePlayer, idx, dealPoints, matchScore int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("bezique.playerLine",
		"name", cuiPlayerName(player, idx),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"deal", strconv.Itoa(dealPoints),
		"match", strconv.Itoa(matchScore),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BeziqueCuiPresenter renders the Bezique CUI view.
type BeziqueCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BeziqueCuiPresenter) Output(b interfaces.BeziqueGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bezique.helpTitle"), func(sb *strings.Builder) {
		phaseKey := "bezique.phaseFirst"
		if b.IsEndgame() {
			phaseKey = "bezique.phaseSecond"
		}
		sb.WriteString(i18n.Tf("bezique.header",
			"deal", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber()),
			"stock", strconv.Itoa(b.GetStockRemaining()),
			"phase", i18n.T(phaseKey)) + "\n")

		if tc := b.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("bezique.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.Tf("bezique.trumpLineNone", "suit", beziqueSuitName(b.GetTrumpSuit())) + "\n")
		}
		sb.WriteString(i18n.Tf("bezique.scoreLine",
			"p0", strconv.Itoa(b.GetMatchScore(0)),
			"p1", strconv.Itoa(b.GetMatchScore(1))) + "\n")

		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(beziquePlayerStr(b.GetPlayer(i), i, b.GetDealPoints(i), b.GetMatchScore(i)))
		}

		// Deal-points breakdown so players see trick vs meld contribution.
		for i := 0; i < b.GetPlayerCnt(); i++ {
			deal := b.GetDealPoints(i)
			meld := b.GetDealMeldPoints(i)
			sb.WriteString(i18n.Tf("bezique.dealBreakdown",
				"name", cuiPlayerName(b.GetPlayer(i), i),
				"trick", strconv.Itoa(deal-meld),
				"meld", strconv.Itoa(meld),
				"total", strconv.Itoa(deal)) + "\n")
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
			beziqueWriteGameEnd(sb, b)
			return
		}
		beziqueWritePrompt(sb, b)
	})
}

// beziqueWriteGameEnd renders the match-end banner.
func beziqueWriteGameEnd(sb *strings.Builder, b interfaces.BeziqueGame) {
	m0 := b.GetMatchScore(0)
	m1 := b.GetMatchScore(1)
	var banner string
	switch b.GetWinnerIdx() {
	case 0:
		banner = i18n.Tf("bezique.gameEndP0", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	case 1:
		banner = i18n.Tf("bezique.gameEndP1", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	default:
		banner = i18n.Tf("bezique.gameEndTie", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	}
	sb.WriteString(color.Green(banner) + "\n")
}

// beziqueWritePrompt renders the phase-specific prompt.
func beziqueWritePrompt(sb *strings.Builder, b interfaces.BeziqueGame) {
	switch b.GetPhase() {
	case domain.BeziquePhasePlay:
		currentIdx := b.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("bezique.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("bezique.promptPlay") + "\n")
	case domain.BeziquePhaseMeld:
		melds := b.GetAvailableMelds(b.GetCurrentPlayerIdx())
		if len(melds) > 0 {
			sb.WriteString(i18n.T("bezique.meldAvailable") + "\n")
			for i, m := range melds {
				sb.WriteString(i18n.Tf("bezique.meldLine",
					"idx", strconv.Itoa(i),
					"name", beziqueMeldName(m),
					"points", strconv.Itoa(m.Points)) + "\n")
			}
		} else {
			sb.WriteString(i18n.T("bezique.meldNone") + "\n")
		}
		sb.WriteString(i18n.T("bezique.promptMeld") + "\n")
	case domain.BeziquePhaseRoundEnd:
		sb.WriteString(i18n.T("bezique.promptRoundEnd") + "\n")
		sb.WriteString(i18n.T("bezique.promptRoundEndHelp") + "\n")
	}
}

// beziqueMeldName メルド種別を i18n 表示名に変換する (結婚はスート名を含む)。
func beziqueMeldName(m domain.BeziqueMeld) string {
	switch m.Type {
	case domain.BeziqueMeldMarriage:
		if m.Points == domain.BeziqueRoyalMarriagePoints {
			return i18n.T("bezique.meldRoyalMarriage")
		}
		return i18n.Tf("bezique.meldMarriage", "suit", beziqueSuitName(m.Suit))
	case domain.BeziqueMeldBezique:
		return i18n.T("bezique.meldBezique")
	case domain.BeziqueMeldFourAces:
		return i18n.T("bezique.meldFourAces")
	case domain.BeziqueMeldFourKings:
		return i18n.T("bezique.meldFourKings")
	case domain.BeziqueMeldFourQueens:
		return i18n.T("bezique.meldFourQueens")
	default:
		return i18n.T("bezique.meldFourJacks")
	}
}

// beziqueSuitName スート番号を i18n のスート名に変換する。
func beziqueSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("bezique.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("bezique.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("bezique.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("bezique.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current Bezique hint.
func (p *BeziqueCuiPresenter) HintOutput(b interfaces.BeziqueGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("bezique.hintNone") + "\n"
	}
	if hint.MeldIndex != nil {
		if *hint.MeldIndex < 0 {
			return color.Yellow(i18n.T("bezique.hintMeldSkip")) + "\n"
		}
		return color.Yellow(i18n.Tf("bezique.hintMeld", "idx", strconv.Itoa(*hint.MeldIndex))) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("bezique.hintNone") + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("bezique.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, beziqueHintReasonKeys))) + "\n"
}

// beziqueHintReasonKeys maps Bezique-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var beziqueHintReasonKeys = map[string]string{
	"lead_trump":  "bezique.hintReasonLeadTrump",
	"lead_low":    "bezique.hintReasonLeadLow",
	"lead_value":  "bezique.hintReasonLeadValue",
	"follow_cut":  "bezique.hintReasonFollowCut",
	"follow_win":  "bezique.hintReasonFollowWin",
	"follow_dump": "bezique.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BeziqueCuiPresenter) ActionLogOutput(b interfaces.BeziqueGame) string {
	return actionLogOutputText(b)
}
