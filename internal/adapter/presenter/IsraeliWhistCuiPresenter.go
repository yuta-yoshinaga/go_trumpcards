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

// israeliWhistPlayerStr returns the display string for a single player.
func israeliWhistPlayerStr(player *domain.IsraeliWhistPlayer, idx int, declarer bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("israeliwhist.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", israeliWhistRoleStr(player, declarer),
		"bid", israeliWhistBidStr(player),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"total", strconv.Itoa(player.GetTotalScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// israeliWhistRoleStr 1 段階目の立場を短く表す
func israeliWhistRoleStr(player *domain.IsraeliWhistPlayer, declarer bool) string {
	switch {
	case declarer:
		return i18n.Tf("israeliwhist.roleDeclarer", "n", strconv.Itoa(player.GetAuctionBid()))
	case player.GetPassed():
		return i18n.T("israeliwhist.rolePassed")
	default:
		return i18n.T("israeliwhist.roleActive")
	}
}

// israeliWhistBidStr 2 段階目の宣言を短く表す
func israeliWhistBidStr(player *domain.IsraeliWhistPlayer) string {
	if player.GetBid() < 0 {
		return i18n.T("israeliwhist.bidNone")
	}
	return i18n.Tf("israeliwhist.bidValue", "n", strconv.Itoa(player.GetBid()))
}

// IsraeliWhistCuiPresenter renders the Israeli Whist CUI view.
type IsraeliWhistCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *IsraeliWhistCuiPresenter) Output(w interfaces.IsraeliWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("israeliwhist.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("israeliwhist.header",
			"round", strconv.Itoa(w.GetRoundNumber()),
			"rounds", strconv.Itoa(w.GetConfig().Rounds),
			"trick", strconv.Itoa(w.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.IsraeliWhistTricksPerRound)) + "\n")
		// **得点表は盤面から読めない。** 2 乗で跳ねることと全員一致の倍率を出す。
		sb.WriteString(i18n.T("israeliwhist.scoreTable") + "\n")

		if w.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("israeliwhist.trumpLine",
				"suit", israeliWhistSuitName(w.GetTrumpSuit()),
				"n", strconv.Itoa(w.GetHighBid())) + "\n")
		} else {
			sb.WriteString(i18n.Tf("israeliwhist.auctionLine",
				"n", strconv.Itoa(w.GetHighBid()),
				"suit", israeliWhistSuitName(w.GetHighSuit())) + "\n")
		}

		for i := 0; i < w.GetPlayerCnt(); i++ {
			sb.WriteString(israeliWhistPlayerStr(w.GetPlayer(i), i, i == w.GetDeclarerIdx()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, w.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(w.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if w.GetGameEndFlag() {
			var banner string
			switch {
			case w.GetWinnerIdx() == 0:
				banner = i18n.Tf("israeliwhist.gameEndWin", "score", strconv.Itoa(w.GetPlayer(0).GetTotalScore()))
			case w.GetWinnerIdx() < 0:
				banner = i18n.T("israeliwhist.gameEndTie")
			default:
				banner = i18n.Tf("israeliwhist.gameEndLose",
					"name", cuiPlayerName(w.GetPlayer(w.GetWinnerIdx()), w.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch w.GetPhase() {
		case domain.IsraeliWhistPhaseAuction:
			if w.IsHumanAuctionTurn() {
				sb.WriteString(i18n.T("israeliwhist.promptAuction") + "\n")
			} else {
				sb.WriteString(i18n.T("israeliwhist.promptAuctionWait") + "\n")
			}
			return
		case domain.IsraeliWhistPhaseBid:
			sb.WriteString(i18n.T("israeliwhist.promptBid") + "\n")
			// **ノルマと禁止値は先に伝える。** 押してから断られるのでは遅い。
			if m := w.MinimumBidFor(0); m > 0 && w.IsHumanBidTurn() {
				sb.WriteString(i18n.Tf("israeliwhist.promptBidQuota", "n", strconv.Itoa(m)) + "\n")
			}
			if r := w.GetRestrictedBid(); r >= 0 && w.IsHumanBidTurn() {
				sb.WriteString(i18n.Tf("israeliwhist.promptBidRestricted", "n", strconv.Itoa(r)) + "\n")
			}
			return
		case domain.IsraeliWhistPhaseRoundEnd:
			// **2 倍はこのゲームの起伏そのもの。**これまで畳まれたアクション
			// ログにしか残っておらず、点が普段の倍動いた理由が読めなかった (#5752)。
			if w.GetRoundDoubled() {
				key := "israeliwhist.doubledAllMissed"
				if w.GetRoundAllExact() {
					key = "israeliwhist.doubledAllExact"
				}
				sb.WriteString(color.Yellow(i18n.T(key)) + "\n")
			}
			sb.WriteString(i18n.T("israeliwhist.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("israeliwhist.promptNext") + "\n")
			return
		}

		currentIdx := w.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("israeliwhist.promptCurrentPlayer",
			"name", cuiPlayerName(w.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("israeliwhist.promptPlay") + "\n")
	})
}

// israeliWhistSuitName スート番号を i18n のスート名に変換する
func israeliWhistSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("israeliwhist.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("israeliwhist.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("israeliwhist.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("israeliwhist.suitDiamond")
	default:
		return i18n.T("israeliwhist.suitNone")
	}
}

// HintOutput emits the current hint.
func (p *IsraeliWhistCuiPresenter) HintOutput(w interfaces.IsraeliWhistGame) string {
	hint := w.GetHint()
	if hint == nil {
		return i18n.T("israeliwhist.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		reason := hintReasonStr(hint.Reason, israeliWhistHintReasonKeys)
		switch hint.Reason {
		case "israeliwhistAuctionBid":
			reason = i18n.Tf("israeliwhist.hintReasonAuctionBidValue",
				"n", strconv.Itoa(hint.Value), "suit", israeliWhistSuitName(hint.Suit))
		case "israeliwhistBid", "israeliwhistAvoidRestricted", "israeliwhistMeetQuota":
			reason = i18n.Tf("israeliwhist.hintReasonBidValue",
				"n", strconv.Itoa(hint.Value),
				"why", hintReasonStr(hint.Reason, israeliWhistHintReasonKeys))
		}
		return color.Yellow(i18n.Tf("israeliwhist.hintCall", "reason", reason)) + "\n"
	}
	card := w.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("israeliwhist.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, israeliWhistHintReasonKeys))) + "\n"
}

// israeliWhistHintReasonKeys maps hint-reason identifiers to their i18n keys.
var israeliWhistHintReasonKeys = map[string]string{
	"israeliwhistAuctionBid":      "israeliwhist.hintReasonAuctionBid",
	"israeliwhistAuctionPass":     "israeliwhist.hintReasonAuctionPass",
	"israeliwhistBid":             "israeliwhist.hintReasonBid",
	"israeliwhistMeetQuota":       "israeliwhist.hintReasonMeetQuota",
	"israeliwhistAvoidRestricted": "israeliwhist.hintReasonAvoidRestricted",
	"israeliwhistWinTrick":        "israeliwhist.hintReasonWinTrick",
	"israeliwhistDuck":            "israeliwhist.hintReasonDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *IsraeliWhistCuiPresenter) ActionLogOutput(w interfaces.IsraeliWhistGame) string {
	return actionLogOutputTextForSeats[*domain.IsraeliWhistPlayer](w)
}
