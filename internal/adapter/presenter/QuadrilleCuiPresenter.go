//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// quadrilleBidLabel maps a bid value to its i18n label key.
func quadrilleBidLabel(bid domain.QuadrilleBid) string {
	switch bid {
	case domain.QuadrilleBidEntrar:
		return i18n.T("quadrille.bidEntrar")
	case domain.QuadrilleBidSolo:
		return i18n.T("quadrille.bidSolo")
	default:
		return i18n.T("quadrille.bidNone")
	}
}

// quadrilleTrumpLabel maps a trump suit value to its i18n label key.
func quadrilleTrumpLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("quadrille.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("quadrille.suitClub")
	case domain.CardDesignHeart:
		return i18n.T("quadrille.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("quadrille.suitDiamond")
	default:
		return i18n.T("quadrille.suitNone")
	}
}

func quadrillePlayerStr(g interfaces.QuadrilleGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("quadrille.roleCoalition")
	switch {
	case idx == g.GetQuadrilleIdx():
		role = i18n.T("quadrille.roleQuadrille")
	// **味方の役柄は伏せている間は出さない。** GetPartnerIdx は呼ばれた王が
	// 場に出るまで -1 を返すので、その間は全員が「連合」に見える。
	case idx == g.GetPartnerIdx():
		role = i18n.T("quadrille.rolePartner")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("quadrille.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(quadrilleIndexedHandStr(player, g.GetTrumpSuit()) + "\n")
	}
	return b.String()
}

// quadrilleMatadorKeys はマタドールの序列と表示名キーの対応。
var quadrilleMatadorKeys = map[int]string{
	1: "quadrille.matadorSpadille",
	2: "quadrille.matadorManille",
	3: "quadrille.matadorBasto",
}

// quadrilleIndexedHandStr は手札をインデックス付きで並べ、マタドールに注記を付ける。
//
// **マタドールは常にトリックに勝つ。**Web はバッジで示すのに、CUI は素の
// 一覧しか出しておらず、序列を覚えていないと気づけなかった (#4919)。
// 切り札が未確定なら注記は付かない。
//
// 索引と区切りは他ゲームと同じ formatCardList に任せる。自前で並べると
// 区切り幅が揃わない。
func quadrilleIndexedHandStr(player cuiCardList, trump int) string {
	return formatCardList(player, func(c *domain.Card) string {
		s := cuiCardStr(c)
		if key := quadrilleMatadorKeys[domain.QuadrilleMatadorRank(c, trump)]; key != "" {
			s += "(" + i18n.T(key) + ")"
		}
		return s
	}, "  ", true)
}

// QuadrilleCuiPresenter renders the Quadrille CUI view.
type QuadrilleCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *QuadrilleCuiPresenter) Output(g interfaces.QuadrilleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("quadrille.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("quadrille.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", quadrilleTrumpLabel(g.GetTrumpSuit())) + "\n")

		b.WriteString(quadrilleKingLine(g))

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(quadrillePlayerStr(g, i))
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
			banner := i18n.Tf("quadrille.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.QuadrillePhaseBid:
			bidderIdx := g.GetCurrentBidderIdx()
			b.WriteString(i18n.Tf("quadrille.promptBid",
				"bid", quadrilleBidLabel(g.GetHighestBid()),
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
			b.WriteString(i18n.T("quadrille.promptBidHelp") + "\n")
		case domain.QuadrillePhaseKingCall:
			b.WriteString(i18n.Tf("quadrille.promptKing",
				"name", cuiPlayerName(g.GetPlayer(g.GetQuadrilleIdx()), g.GetQuadrilleIdx())) + "\n")
			b.WriteString(i18n.T("quadrille.promptKingHelp") + "\n")
		case domain.QuadrillePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("quadrille.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
				"trump", quadrilleTrumpLabel(g.GetTrumpSuit())) + "\n")
			b.WriteString(i18n.T("quadrille.promptPlayHelp") + "\n")
		case domain.QuadrillePhaseTrickEnd:
			b.WriteString(i18n.T("quadrille.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("quadrille.promptTrickEndHelp") + "\n")
		case domain.QuadrillePhaseRoundEnd:
			b.WriteString(i18n.Tf("quadrille.promptRoundEnd",
				"quadrille", cuiPlayerName(g.GetPlayer(g.GetQuadrilleIdx()), g.GetQuadrilleIdx()),
				"outcome", quadrilleOutcomeLabel(g.GetOutcome())) + "\n")
			b.WriteString(i18n.T("quadrille.promptRoundEndHelp") + "\n")
		}
	})
}

// quadrilleOutcomeLabel maps a deal outcome to its i18n label key.
func quadrilleOutcomeLabel(o domain.QuadrilleOutcome) string {
	switch o {
	case domain.QuadrilleOutcomeSacar:
		return i18n.T("quadrille.outcomeSacar")
	case domain.QuadrilleOutcomePuesta:
		return i18n.T("quadrille.outcomePuesta")
	case domain.QuadrilleOutcomeCodille:
		return i18n.T("quadrille.outcomeCodille")
	default:
		return i18n.T("quadrille.outcomeNone")
	}
}

// HintOutput emits the current Quadrille hint.
func (p *QuadrilleCuiPresenter) HintOutput(g interfaces.QuadrilleGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("quadrille.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, quadrilleHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("quadrille.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	// Bid-phase decisions carry no cards; render them as an action recommendation
	// instead of the meaningless "recommended cards: -" line.
	if actionKey, ok := quadrilleBidActionKeys[hint.Reason]; ok {
		return color.Yellow(i18n.Tf("quadrille.hintDecision",
			"action", i18n.T(actionKey),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("quadrille.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// quadrilleBidActionKeys maps bid-phase hint-reason identifiers to the i18n key for
// the recommended action name (Entrar / Solo / Pass).
var quadrilleBidActionKeys = map[string]string{
	"bid_entrar": "quadrille.hintActionEntrar",
	"bid_solo":   "quadrille.hintActionSolo",
	"bid_pass":   "quadrille.hintActionPass",
}

// quadrilleHintReasonKeys maps Quadrille-specific hint-reason identifiers to i18n keys.
var quadrilleHintReasonKeys = map[string]string{
	"lead_high":    "quadrille.hintReasonLeadHigh",
	"lead_low":     "quadrille.hintReasonLeadLow",
	"follow_win":   "quadrille.hintReasonFollowWin",
	"follow_duck":  "quadrille.hintReasonFollowDuck",
	"give_partner": "quadrille.hintReasonGivePartner",
	"discard_low":  "quadrille.hintReasonDiscardLow",
	"bid_entrar":   "quadrille.hintReasonBidEntrar",
	"bid_solo":     "quadrille.hintReasonBidSolo",
	"bid_pass":     "quadrille.hintReasonBidPass",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *QuadrilleCuiPresenter) ActionLogOutput(g interfaces.QuadrilleGame) string {
	return actionLogOutputTextForSeats[*domain.QuadrillePlayer](g)
}

// quadrilleKingLine は王呼びの行を組み立てる。
//
// **呼ばれた王は公開情報、持ち主は伏せる。** 呼び声は卓で聞こえるので王の
// スートは全員が知っているが、誰が持っているかはその王が場に出るまで
// 分からない —— そこがこのゲームの緊張感そのもの。
func quadrilleKingLine(g interfaces.QuadrilleGame) string {
	if g.IsRoiSeul() {
		return i18n.T("quadrille.roiSeul") + "\n"
	}
	suit := g.GetCalledKingSuit()
	if !quadrilleSuitInRange(suit) {
		return i18n.T("quadrille.kingUncalled") + "\n"
	}
	partner := g.GetPartnerIdx()
	if partner < 0 || partner >= g.GetPlayerCnt() {
		return i18n.Tf("quadrille.kingCalledHidden",
			"suit", quadrilleTrumpLabel(suit)) + "\n"
	}
	return i18n.Tf("quadrille.kingCalledRevealed",
		"suit", quadrilleTrumpLabel(suit),
		"name", cuiPlayerName(g.GetPlayer(partner), partner)) + "\n"
}

// quadrilleSuitInRange は suit が本物のスート (1..4) かを返す。
func quadrilleSuitInRange(suit int) bool {
	return suit >= domain.CardDesignSpade && suit <= domain.CardDesignMax
}
