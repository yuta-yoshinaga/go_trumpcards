//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ultiContractLabel maps a contract value to its i18n label key.
func ultiContractLabel(contract domain.UltiContract) string {
	switch contract {
	case domain.UltiContractParty:
		return i18n.T("ulti.contractParty")
	case domain.UltiContractBetli:
		return i18n.T("ulti.contractBetli")
	case domain.UltiContractDurchmarsch:
		return i18n.T("ulti.contractDurchmarsch")
	case domain.UltiContractUlti:
		return i18n.T("ulti.contractUlti")
	default:
		return i18n.T("ulti.contractNone")
	}
}

// ultiTrumpLabel maps a trump suit value to its i18n label key.
func ultiTrumpLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("ulti.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("ulti.suitClub")
	case domain.CardDesignHeart:
		return i18n.T("ulti.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("ulti.suitDiamond")
	default:
		return i18n.T("ulti.suitNone")
	}
}

func ultiPlayerStr(g interfaces.UltiGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	coins := g.GetPlayerCoins()
	role := i18n.T("ulti.roleCoalition")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("ulti.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("ulti.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"coins", strconv.Itoa(coins[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(g.GetCardPoints(idx)),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// UltiCuiPresenter renders the Ulti CUI view.
type UltiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *UltiCuiPresenter) Output(g interfaces.UltiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ulti.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ulti.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", ultiContractLabel(g.GetContract()),
			"trump", ultiTrumpLabel(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ultiPlayerStr(g, i))
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
			banner := i18n.Tf("ulti.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			// マッチを決めたディールは RoundEnd を飛ばして GameEnd に入るので、
			// ここで出さないと最終ディールの精算だけ見えないまま終わる。
			ultiCoinSettlementLine(b, g)
			return
		}
		switch g.GetPhase() {
		case domain.UltiPhaseBid:
			b.WriteString(i18n.Tf("ulti.promptBid",
				"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
			b.WriteString(i18n.T("ulti.promptBidHelp") + "\n")
		case domain.UltiPhaseDiscard:
			b.WriteString(i18n.Tf("ulti.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
			b.WriteString(i18n.T("ulti.promptDiscardHelp") + "\n")
		case domain.UltiPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ulti.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
				"trump", ultiTrumpLabel(g.GetTrumpSuit())) + "\n")
			b.WriteString(i18n.T("ulti.promptPlayHelp") + "\n")
		case domain.UltiPhaseTrickEnd:
			b.WriteString(i18n.T("ulti.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("ulti.promptTrickEndHelp") + "\n")
		case domain.UltiPhaseRoundEnd:
			b.WriteString(i18n.Tf("ulti.promptRoundEnd",
				"declarer", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx()),
				"outcome", ultiOutcomeLabel(g.GetOutcome())) + "\n")
			ultiCoinSettlementLine(b, g)
			b.WriteString(i18n.T("ulti.promptRoundEndHelp") + "\n")
		}
	})
}

// ultiCoinSettlementLine writes the signed coin change this deal applied.
// 累積コインだけでは「今回いくら動いたか」が読めないので、精算のあった
// ディールでのみ 1 行足す (未精算のディール中は全員 0 なので出さない)。
func ultiCoinSettlementLine(b *strings.Builder, g interfaces.UltiGame) {
	deltas := g.GetLastDealCoins()
	moved := false
	parts := make([]string, 0, len(deltas))
	for i, d := range deltas {
		if d != 0 {
			moved = true
		}
		parts = append(parts, fmt.Sprintf("%s %+d", cuiPlayerName(g.GetPlayer(i), i), d))
	}
	if !moved {
		return
	}
	b.WriteString(i18n.Tf("ulti.coinSettlement", "deltas", strings.Join(parts, " / ")) + "\n")
}

// ultiOutcomeLabel maps a deal outcome to its i18n label key.
func ultiOutcomeLabel(o domain.UltiOutcome) string {
	switch o {
	case domain.UltiOutcomeWin:
		return i18n.T("ulti.outcomeWin")
	case domain.UltiOutcomeLoss:
		return i18n.T("ulti.outcomeLoss")
	default:
		return i18n.T("ulti.outcomeNone")
	}
}

// HintOutput emits the current Ulti hint.
func (p *UltiCuiPresenter) HintOutput(g interfaces.UltiGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("ulti.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, ultiHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.UltiPhaseDiscard {
			playerIdx = g.GetDeclarerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("ulti.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("ulti.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// ultiHintReasonKeys maps Ulti-specific hint-reason identifiers to i18n keys.
var ultiHintReasonKeys = map[string]string{
	"lead_high":       "ulti.hintReasonLeadHigh",
	"lead_low":        "ulti.hintReasonLeadLow",
	"follow_win":      "ulti.hintReasonFollowWin",
	"follow_duck":     "ulti.hintReasonFollowDuck",
	"discard_low":     "ulti.hintReasonDiscardLow",
	"discard_weak":    "ulti.hintReasonDiscardWeak",
	"bid_party":       "ulti.hintReasonBidParty",
	"bid_betli":       "ulti.hintReasonBidBetli",
	"bid_durchmarsch": "ulti.hintReasonBidDurchmarsch",
	"bid_ulti":        "ulti.hintReasonBidUlti",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *UltiCuiPresenter) ActionLogOutput(g interfaces.UltiGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
