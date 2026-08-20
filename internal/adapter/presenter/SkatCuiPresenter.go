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

// skatPlayerStr returns the display string for a single Skat player.
func skatPlayerStr(player *domain.SkatPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	role := ""
	if player.GetIsDeclarer() {
		role = i18n.T("skat.roleDeclarer")
	}
	bidStr := "-"
	if player.GetBid() == 0 {
		bidStr = i18n.T("skat.choiceBidPass")
	} else if player.GetBid() > 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	fmt.Fprintf(&b, "%s%s: bid=%s tricks=%d cardPts=%d total=%d round=%d hand=%d\n",
		name, role, bidStr,
		player.GetTrickCount(),
		player.GetCardPoints(),
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// skatBonusStr は乗数に加算された理由を並べる。何も無ければ空。
//
// 敗北の 2 倍とオーバービッドはここには入れない。どちらも乗数ではなく**式そのもの**を
// 変えるので、skatBreakdownLine が別の文で言う。
func skatBonusStr(bd *domain.SkatScoreBreakdown) string {
	parts := make([]string, 0, 3)
	if bd.Hand {
		parts = append(parts, i18n.T("skat.bonusHand"))
	}
	if bd.Schneider {
		parts = append(parts, i18n.T("skat.bonusSchneider"))
	}
	if bd.Schwarz {
		parts = append(parts, i18n.T("skat.bonusSchwarz"))
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

// skatBreakdownLine は得点の式を書く。**式は 3 通りある。**
//
// 勝ちなら 基礎点×乗数、負けならその 2 倍、オーバービッドなら基礎点と無関係な
// 入札×2 に置き換わる。1 つの文に固定すると、負けたラウンド (全体のおよそ半分) で
// 「11 × 3 = 66」という嘘の式が出る。
func skatBreakdownLine(bd *domain.SkatScoreBreakdown) string {
	if bd.Overbid {
		return i18n.Tf("skat.scoreBreakdownLineOverbid",
			"bid", strconv.Itoa(bd.Bid),
			"value", strconv.Itoa(bd.Value),
			"base", strconv.Itoa(bd.Base),
			"multiplier", strconv.Itoa(bd.Multiplier))
	}
	key := "skat.scoreBreakdownLine"
	if bd.Doubled {
		key = "skat.scoreBreakdownLineDoubled"
	}
	return i18n.Tf(key,
		"base", strconv.Itoa(bd.Base),
		"matadors", strconv.Itoa(bd.Matadors),
		"multiplier", strconv.Itoa(bd.Multiplier),
		"value", strconv.Itoa(bd.Value),
		"bonuses", skatBonusStr(bd))
}

// SkatCuiPresenter Skat CUI presenter.
type SkatCuiPresenter struct{}

// Output renders the game state as a CUI string.
func (p *SkatCuiPresenter) Output(s interfaces.SkatGame, lastErr error) string {
	return buildCuiOutput(i18n.T("skat.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("skat.statusLine",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()),
			"dealer", strconv.Itoa(s.GetDealerIdx()),
			"fore", strconv.Itoa(s.GetForehandIdx()),
			"mid", strconv.Itoa(s.GetMiddlehandIdx()),
			"rear", strconv.Itoa(s.GetRearhandIdx())) + "\n")

		if s.GetGameType() != domain.SkatGameNone {
			b.WriteString(i18n.Tf("skat.gameLabel", "type", skatGameTypeLabel(s.GetGameType())))
			if s.GetGameType() == domain.SkatGameSuit {
				b.WriteString(i18n.Tf("skat.trumpLabel", "suit", skatSuitSymbol(s.GetTrumpSuit())))
			}
			b.WriteString("\n")
		}
		if s.GetCurrentBid() > 0 {
			b.WriteString(i18n.Tf("skat.currentBid", "bid", strconv.Itoa(s.GetCurrentBid())) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(skatPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(color.Green(i18n.T("skat.gameOver")) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.SkatPhaseBid:
			actor := s.GetActiveBidActorIdx()
			if actor >= 0 {
				name := cuiPlayerName(s.GetPlayer(actor), actor)
				b.WriteString(i18n.Tf("skat.biddingTurn", "name", name) + "\n")
			}
			// **どこまで受けて安全かの目安を出す。**Web は常時表示しているのに
			// CUI には無く、オーバービッドの危険を測れなかった (#4905)。
			if human := s.GetPlayer(0); human != nil && human.GetCardsSize() > 0 {
				hand := make([]*domain.Card, human.GetCardsSize())
				for i := range hand {
					hand[i] = human.GetCard(i)
				}
				est := domain.SkatBestBidEstimate(hand)
				b.WriteString(i18n.Tf("skat.bidEstimate",
					"value", strconv.Itoa(est.Value),
					"game", skatEstimateGameLabel(est.GameType, est.TrumpSuit),
					"matadors", strconv.Itoa(est.Matadors)) + "\n")
				if cur := s.GetCurrentBid(); cur > est.Value {
					b.WriteString(color.Red(i18n.Tf("skat.bidExceedsHand", "value", strconv.Itoa(est.Value))) + "\n")
				}
			}
			b.WriteString(i18n.T("skat.promptBid") + "\n")
		case domain.SkatPhaseSkatPickup:
			b.WriteString(i18n.T("skat.skatPickup") + "\n")
			b.WriteString(i18n.T("skat.promptPickup") + "\n")
		case domain.SkatPhaseDiscard:
			b.WriteString(i18n.T("skat.discardPrompt") + "\n")
			b.WriteString(i18n.T("skat.promptDiscard") + "\n")
		case domain.SkatPhaseGameDeclaration:
			b.WriteString(i18n.T("skat.gameDeclaration") + "\n")
			b.WriteString(i18n.T("skat.promptDeclare") + "\n")
		case domain.SkatPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			player := s.GetPlayer(currentIdx)
			b.WriteString(i18n.Tf("skat.turnLabel", "name", cuiPlayerName(player, currentIdx)) + "\n")
			b.WriteString(i18n.T("skat.promptPlay") + "\n")
		case domain.SkatPhaseTrickEnd:
			b.WriteString(i18n.T("skat.trickComplete") + "\n")
			b.WriteString(i18n.T("skat.promptNext") + "\n")
		case domain.SkatPhaseRoundEnd:
			b.WriteString(i18n.Tf("skat.roundEndLine",
				"declarer", strconv.Itoa(s.GetDeclarerCardPoints()),
				"defenders", strconv.Itoa(s.GetDefendersCardPoints()),
				"value", strconv.Itoa(s.GetGameValue())) + "\n")
			// **なぜこの点数なのか。**マタドール (切り札の連続所持/不所持) は
			// スカートで最も分かりにくい規則なのに、最終値しか出ていなかった (#5561)。
			if bd := s.GetScoreBreakdown(); bd != nil && !bd.Null {
				b.WriteString(skatBreakdownLine(bd) + "\n")
			}
			b.WriteString(i18n.T("skat.promptNextRound") + "\n")
		}
	})
}

// HintOutput renders the hint output.
func (p *SkatCuiPresenter) HintOutput(s interfaces.SkatGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("skat.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, skatHintReasonKeys)
	if hint.Bid != nil {
		choice := i18n.T("skat.choiceBidPass")
		if *hint.Bid == 1 {
			choice = i18n.T("skat.choiceBidAccept")
		}
		return color.Yellow(i18n.Tf("skat.hintBid", "choice", choice, "reason", reason)) + "\n"
	}
	if hint.PickSkat != nil {
		choice := i18n.T("skat.choiceSkatDecline")
		if *hint.PickSkat {
			choice = i18n.T("skat.choiceSkatPickUp")
		}
		return color.Yellow(i18n.Tf("skat.hintPickSkat", "choice", choice, "reason", reason)) + "\n"
	}
	if hint.DiscardIndex != nil {
		card := s.GetPlayer(0).GetCard(*hint.DiscardIndex)
		return color.Yellow(i18n.Tf("skat.hintDiscard",
			"idx", strconv.Itoa(*hint.DiscardIndex), "card", cuiCardStr(card), "reason", reason)) + "\n"
	}
	if hint.GameType != nil {
		gt := domain.SkatGameType(*hint.GameType)
		s2 := skatGameTypeLabel(gt)
		if gt == domain.SkatGameSuit && hint.TrumpSuit != nil {
			s2 = fmt.Sprintf("%s %s", s2, skatSuitSymbol(*hint.TrumpSuit))
		}
		return color.Yellow(i18n.Tf("skat.hintDeclare", "game", s2, "reason", reason)) + "\n"
	}
	if hint.CardIndex != nil {
		card := s.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("skat.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "card", cuiCardStr(card), "reason", reason)) + "\n"
	}
	return i18n.T("skat.hintNone") + "\n"
}

// ActionLogOutput returns the round's action log as text.
func (p *SkatCuiPresenter) ActionLogOutput(s interfaces.SkatGame) string {
	return actionLogOutputTextForSeats[*domain.SkatPlayer](s)
}

// skatGameTypeLabel returns the localized label for a Skat game type.
func skatGameTypeLabel(gt domain.SkatGameType) string {
	switch gt {
	case domain.SkatGameSuit:
		return i18n.T("skat.gameTypeSuit")
	case domain.SkatGameGrand:
		return i18n.T("skat.gameTypeGrand")
	case domain.SkatGameNull:
		return i18n.T("skat.gameTypeNull")
	}
	return i18n.T("skat.gameTypeNone")
}

// skatSuitSymbol returns the suit symbol.
func skatSuitSymbol(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "?"
}

// skatHintReasonKeys maps Skat-specific hint-reason identifiers to i18n keys.
var skatHintReasonKeys = map[string]string{
	"strategic_bid": "skat.hintReasonStrategicBid",
	"skat_pickup":   "skat.hintReasonSkatPickup",
	"discard_low":   "skat.hintReasonDiscardLow",
	"game_choice":   "skat.hintReasonGameChoice",
	"best_play":     "skat.hintReasonBestPlay",
}

// skatEstimateGameLabel は見積もりの契約名を返す。スート戦は切札名まで出す。
func skatEstimateGameLabel(gameType domain.SkatGameType, trumpSuit int) string {
	if gameType == domain.SkatGameSuit {
		return skatGameTypeLabel(gameType) + " " + cuiSuitName(trumpSuit)
	}
	return skatGameTypeLabel(gameType)
}
