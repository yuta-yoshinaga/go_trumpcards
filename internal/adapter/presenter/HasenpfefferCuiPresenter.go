//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// hasenpfefferBidStr は宣言の表示文字列を返す。
func hasenpfefferBidStr(bid int) string {
	switch {
	case bid < 0:
		return i18n.T("hasenpfeffer.bidNone")
	case bid == 0:
		return i18n.T("hasenpfeffer.bidPassed")
	default:
		return i18n.Tf("hasenpfeffer.bidValue", "n", strconv.Itoa(bid))
	}
}

// hasenpfefferPlayerStr returns the display string for a single player.
func hasenpfefferPlayerStr(player *domain.HasenpfefferPlayer, idx int, isDeclarer, isDealer, current bool) string {
	var b strings.Builder
	role := ""
	if isDeclarer {
		role = i18n.T("hasenpfeffer.roleDeclarer")
	} else if isDealer {
		role = i18n.T("hasenpfeffer.roleDealer")
	}
	marker := " "
	if current {
		marker = ">"
	}
	b.WriteString(marker + i18n.Tf("hasenpfeffer.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.HasenpfefferTeamOf(idx)),
		"role", role,
		"bid", hasenpfefferBidStr(player.GetBid()),
		"took", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// HasenpfefferCuiPresenter renders the Hasenpfeffer CUI view.
type HasenpfefferCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *HasenpfefferCuiPresenter) Output(h interfaces.HasenpfefferGame, lastErr error) string {
	return buildCuiOutput(i18n.T("hasenpfeffer.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("hasenpfeffer.header",
			"hand", strconv.Itoa(h.GetHandNumber()),
			"trick", strconv.Itoa(h.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.HasenpfefferTricksPerRound),
			"target", strconv.Itoa(h.GetConfig().Target)) + "\n")
		// **ジョーカーが全カード中最強。** 序列を知らないと打ち方が変わる。
		sb.WriteString(i18n.T("hasenpfeffer.rule") + "\n")
		sb.WriteString(i18n.Tf("hasenpfeffer.scoreLine",
			"t0", strconv.Itoa(h.GetScore(0)),
			"t1", strconv.Itoa(h.GetScore(1))) + "\n")

		if h.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("hasenpfeffer.trumpLine",
				"suit", hasenpfefferSuitName(h.GetTrumpSuit()),
				"contract", strconv.Itoa(h.GetContract()),
				"name", cuiPlayerName(h.GetPlayer(h.GetDeclarerIdx()), h.GetDeclarerIdx())) + "\n")
		} else if h.GetBlindSize() > 0 {
			sb.WriteString(i18n.Tf("hasenpfeffer.blindLine",
				"n", strconv.Itoa(h.GetBlindSize())) + "\n")
		} else {
			sb.WriteString(i18n.T("hasenpfeffer.trumpUndecided") + "\n")
		}

		for i := 0; i < h.GetPlayerCnt(); i++ {
			sb.WriteString(hasenpfefferPlayerStr(h.GetPlayer(i), i,
				i == h.GetDeclarerIdx(), i == h.GetDealerIdx(),
				i == h.GetCurrentPlayerIdx() && !h.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, h.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(h.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if h.GetGameEndFlag() {
			var banner string
			switch h.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("hasenpfeffer.gameEndTeam0", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			case 1:
				banner = i18n.Tf("hasenpfeffer.gameEndTeam1", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			default:
				banner = i18n.Tf("hasenpfeffer.gameEndTie", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch h.GetPhase() {
		case domain.HasenpfefferPhaseBid:
			p.writeBidPrompt(sb, h)
		case domain.HasenpfefferPhaseDiscard:
			if h.IsHumanDiscardTurn() {
				sb.WriteString(i18n.Tf("hasenpfeffer.promptDiscard",
					"contract", strconv.Itoa(h.GetContract())) + "\n")
			} else {
				sb.WriteString(i18n.T("hasenpfeffer.promptDiscardWait") + "\n")
			}
		case domain.HasenpfefferPhaseHandEnd:
			// **落としたのか達成したのかは盤面から読めない。**
			key := "hasenpfeffer.handEndMade"
			if h.GetLastHandEuchred() {
				key = "hasenpfeffer.handEndEuchred"
			}
			sb.WriteString(i18n.Tf(key,
				"contract", strconv.Itoa(h.GetContract()),
				"took", strconv.Itoa(h.GetLastHandTricks())) + "\n")
			sb.WriteString(i18n.T("hasenpfeffer.promptNext") + "\n")
		default:
			currentIdx := h.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("hasenpfeffer.promptCurrentPlayer",
				"name", cuiPlayerName(h.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("hasenpfeffer.promptPlay") + "\n")
		}
	})
}

// writeBidPrompt は競りの案内を書く。
//
// **降りられるかどうかが場面で変わる。** 親は 3 人が降りたら降りられず、
// 上限が立っていたら逆に降りるしかない。
func (p *HasenpfefferCuiPresenter) writeBidPrompt(sb *strings.Builder, h interfaces.HasenpfefferGame) {
	if !h.IsHumanBidTurn() {
		sb.WriteString(i18n.T("hasenpfeffer.promptBidWait") + "\n")
		return
	}
	switch {
	case h.MustBid(0):
		sb.WriteString(i18n.Tf("hasenpfeffer.promptMustBid", "min", strconv.Itoa(h.NextBid())) + "\n")
	case h.NextBid() == 0:
		sb.WriteString(i18n.Tf("hasenpfeffer.promptBidCapped",
			"max", strconv.Itoa(domain.HasenpfefferMaxBid)) + "\n")
	default:
		sb.WriteString(i18n.Tf("hasenpfeffer.promptBid",
			"min", strconv.Itoa(h.NextBid()),
			"max", strconv.Itoa(domain.HasenpfefferMaxBid)) + "\n")
	}
}

// hasenpfefferSuitName スート番号を i18n のスート名に変換する
func hasenpfefferSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("hasenpfeffer.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("hasenpfeffer.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("hasenpfeffer.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("hasenpfeffer.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *HasenpfefferCuiPresenter) HintOutput(h interfaces.HasenpfefferGame) string {
	hint := h.GetHint()
	if hint == nil {
		return i18n.T("hasenpfeffer.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, hasenpfefferHintReasonKeys)
	if hint.CardIndex == nil {
		// **競りの助言は札ではなく額を指す。**
		return color.Yellow(i18n.Tf("hasenpfeffer.hintBid",
			"n", strconv.Itoa(hint.Value), "reason", reason)) + "\n"
	}
	card := h.GetPlayer(0).GetCard(*hint.CardIndex)
	if hint.Suit > 0 {
		// 捨て札の助言は札とスートの両方を指す。
		return color.Yellow(i18n.Tf("hasenpfeffer.hintDiscard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"suit", hasenpfefferSuitName(hint.Suit),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("hasenpfeffer.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// hasenpfefferHintReasonKeys maps hint-reason identifiers to their i18n keys.
var hasenpfefferHintReasonKeys = map[string]string{
	"hasenpfefferBid":         "hasenpfeffer.hintReasonBid",
	"hasenpfefferPass":        "hasenpfeffer.hintReasonPass",
	"hasenpfefferMustBid":     "hasenpfeffer.hintReasonMustBid",
	"hasenpfefferDiscard":     "hasenpfeffer.hintReasonDiscard",
	"hasenpfefferWinTrick":    "hasenpfeffer.hintReasonWinTrick",
	"hasenpfefferFeedPartner": "hasenpfeffer.hintReasonFeedPartner",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HasenpfefferCuiPresenter) ActionLogOutput(h interfaces.HasenpfefferGame) string {
	return actionLogOutputTextWithNames(h, func(idx int) string { return cuiPlayerName(h.GetPlayer(idx), idx) })
}
