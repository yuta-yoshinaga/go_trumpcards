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

// ecartePlayerStr returns the display string for a single Écarté player.
func ecartePlayerStr(player *domain.EcartePlayer, idx, dealPoints, matchScore int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("ecarte.playerLine",
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

// EcarteCuiPresenter renders the Écarté CUI view.
type EcarteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *EcarteCuiPresenter) Output(b interfaces.EcarteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ecarte.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("ecarte.header",
			"deal", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber()),
			"stock", strconv.Itoa(b.GetStockRemaining()),
			"phase", i18n.T(ecartePhaseKey(b.GetPhase()))) + "\n")

		if tc := b.GetTrumpCard(); tc != nil {
			sb.WriteString(i18n.Tf("ecarte.trumpLine", "card", cuiCardStr(tc)) + "\n")
		} else {
			sb.WriteString(i18n.Tf("ecarte.trumpLineNone", "suit", ecarteSuitName(b.GetTrumpSuit())) + "\n")
		}
		sb.WriteString(i18n.Tf("ecarte.scoreLine",
			"p0", strconv.Itoa(b.GetMatchScore(0)),
			"p1", strconv.Itoa(b.GetMatchScore(1))) + "\n")

		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(ecartePlayerStr(b.GetPlayer(i), i, b.GetDealPoints(i), b.GetMatchScore(i)))
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
			ecarteWriteGameEnd(sb, b)
			return
		}
		ecarteWritePrompt(sb, b)
	})
}

// ecartePhaseKey フェーズに対応する i18n キーを返す。
func ecartePhaseKey(phase domain.EcartePhase) string {
	switch phase {
	case domain.EcartePhaseExchange:
		return "ecarte.phaseExchange"
	case domain.EcartePhasePlay:
		return "ecarte.phasePlay"
	case domain.EcartePhaseRoundEnd:
		return "ecarte.phaseRoundEnd"
	default:
		return "ecarte.phaseGameEnd"
	}
}

// ecarteWriteGameEnd renders the match-end banner.
func ecarteWriteGameEnd(sb *strings.Builder, b interfaces.EcarteGame) {
	m0 := b.GetMatchScore(0)
	m1 := b.GetMatchScore(1)
	var banner string
	switch b.GetWinnerIdx() {
	case 0:
		banner = i18n.Tf("ecarte.gameEndP0", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	case 1:
		banner = i18n.Tf("ecarte.gameEndP1", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	default:
		banner = i18n.Tf("ecarte.gameEndTie", "p0", strconv.Itoa(m0), "p1", strconv.Itoa(m1))
	}
	sb.WriteString(color.Green(banner) + "\n")
}

// ecarteWritePrompt renders the phase-specific prompt.
func ecarteWritePrompt(sb *strings.Builder, b interfaces.EcarteGame) {
	currentIdx := b.GetCurrentPlayerIdx()
	switch b.GetPhase() {
	case domain.EcartePhaseExchange:
		sb.WriteString(i18n.Tf("ecarte.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		step := b.GetNegStep()
		sb.WriteString(i18n.T(ecarteNegPromptKey(step)) + "\n")
		// On a discard step, spell out the upper bound (you cannot discard more
		// cards than the stock can replace) so the player need not compute it.
		if step == domain.EcarteNegElderDiscard || step == domain.EcarteNegDealerDiscard {
			stock := b.GetStockRemaining()
			maxDiscard := stock
			if player := b.GetPlayer(currentIdx); player != nil && player.GetCardsSize() < maxDiscard {
				maxDiscard = player.GetCardsSize()
			}
			sb.WriteString(i18n.Tf("ecarte.promptDiscardLimit",
				"max", strconv.Itoa(maxDiscard),
				"stock", strconv.Itoa(stock)) + "\n")
		}
	case domain.EcartePhasePlay:
		sb.WriteString(i18n.Tf("ecarte.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("ecarte.promptPlay") + "\n")
	case domain.EcartePhaseRoundEnd:
		sb.WriteString(i18n.T("ecarte.promptRoundEnd") + "\n")
		sb.WriteString(i18n.T("ecarte.promptRoundEndHelp") + "\n")
	}
}

// ecarteNegPromptKey 交換ステップに対応するプロンプト i18n キーを返す。
func ecarteNegPromptKey(step domain.EcarteNegStep) string {
	switch step {
	case domain.EcarteNegElderDecide:
		return "ecarte.promptElderDecide"
	case domain.EcarteNegDealerRespond:
		return "ecarte.promptDealerRespond"
	case domain.EcarteNegElderDiscard, domain.EcarteNegDealerDiscard:
		return "ecarte.promptDiscard"
	default:
		return "ecarte.promptElderDecide"
	}
}

// ecarteSuitName スート番号を i18n のスート名に変換する。
func ecarteSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("ecarte.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("ecarte.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("ecarte.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("ecarte.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current Écarté hint.
func (p *EcarteCuiPresenter) HintOutput(b interfaces.EcarteGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("ecarte.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		// 交換フェーズのアクションヒント。
		return color.Yellow(i18n.Tf("ecarte.hintAction",
			"action", ecarteActionName(hint.Action))) + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("ecarte.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, ecarteHintReasonKeys))) + "\n"
}

// ecarteActionName 交換アクション名を i18n 表示名に変換する。
func ecarteActionName(action string) string {
	switch action {
	case "propose":
		return i18n.T("ecarte.actionPropose")
	case "stand":
		return i18n.T("ecarte.actionStand")
	case "accept":
		return i18n.T("ecarte.actionAccept")
	case "refuse":
		return i18n.T("ecarte.actionRefuse")
	case "discard":
		return i18n.T("ecarte.actionDiscard")
	default:
		return action
	}
}

// ecarteHintReasonKeys maps Écarté-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var ecarteHintReasonKeys = map[string]string{
	"lead_trump":  "ecarte.hintReasonLeadTrump",
	"lead_high":   "ecarte.hintReasonLeadHigh",
	"follow_win":  "ecarte.hintReasonFollowWin",
	"follow_dump": "ecarte.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EcarteCuiPresenter) ActionLogOutput(b interfaces.EcarteGame) string {
	return actionLogOutputTextWithNames(b, func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) })
}
