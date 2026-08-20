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

// estimationPlayerStr returns the display string for a single player.
func estimationPlayerStr(player *domain.EstimationPlayer, idx int, roundEnd bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("estimation.playerLine",
		"name", cuiPlayerName(player, idx),
		"bid", estimationBidStr(player),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"total", strconv.Itoa(player.GetTotalScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	// **得点式が複雑（10+宣言 / Dash Call ±23 / Risk 2倍）なので、累計の差分を
	// 暗算させない** (#5751)。増減が確定するラウンド終了時にだけ出す。
	if roundEnd {
		b.WriteString(" " + i18n.Tf("estimation.roundDelta",
			"delta", estimationSignedScore(player.GetRoundScore())))
	}
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// estimationSignedScore は増減を符号付きで表す。**+ は自分で付ける。**
// 0 は「動かなかった」ことを示すので ±0 と書く。
func estimationSignedScore(n int) string {
	if n > 0 {
		return "+" + strconv.Itoa(n)
	}
	if n == 0 {
		return "±0"
	}
	return strconv.Itoa(n)
}

// estimationBidStr 宣言を種類つきで短く表す
func estimationBidStr(player *domain.EstimationPlayer) string {
	if player.GetBid() < 0 {
		return i18n.T("estimation.bidNone")
	}
	switch player.GetCallType() {
	case domain.EstimationCallDash:
		return i18n.T("estimation.bidDash")
	case domain.EstimationCallRisk:
		return i18n.Tf("estimation.bidRisk", "n", strconv.Itoa(player.GetBid()))
	default:
		return i18n.Tf("estimation.bidNormal", "n", strconv.Itoa(player.GetBid()))
	}
}

// EstimationCuiPresenter renders the Estimation CUI view.
type EstimationCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *EstimationCuiPresenter) Output(e interfaces.EstimationGame, lastErr error) string {
	return buildCuiOutput(i18n.T("estimation.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("estimation.header",
			"round", strconv.Itoa(e.GetRoundNumber()),
			"rounds", strconv.Itoa(e.GetConfig().Rounds),
			"trick", strconv.Itoa(e.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.EstimationTricksPerRound)) + "\n")
		// **得点表は盤面から読めない。** Dash と Risk の振れ幅を常時出す。
		sb.WriteString(i18n.T("estimation.scoreTable") + "\n")

		if e.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("estimation.trumpLine", "suit", estimationSuitName(e.GetTrumpSuit())) + "\n")
		} else {
			sb.WriteString(i18n.T("estimation.trumpUndecided") + "\n")
		}

		for i := 0; i < e.GetPlayerCnt(); i++ {
			sb.WriteString(estimationPlayerStr(e.GetPlayer(i), i,
				e.GetPhase() == domain.EstimationPhaseRoundEnd))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, e.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(e.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if e.GetGameEndFlag() {
			var banner string
			switch {
			case e.GetWinnerIdx() == 0:
				banner = i18n.Tf("estimation.gameEndWin", "score", strconv.Itoa(e.GetPlayer(0).GetTotalScore()))
			case e.GetWinnerIdx() < 0:
				banner = i18n.T("estimation.gameEndTie")
			default:
				banner = i18n.Tf("estimation.gameEndLose",
					"name", cuiPlayerName(e.GetPlayer(e.GetWinnerIdx()), e.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch e.GetPhase() {
		case domain.EstimationPhaseTrump:
			if e.IsHumanTrumpTurn() {
				sb.WriteString(i18n.T("estimation.promptTrump") + "\n")
			} else {
				sb.WriteString(i18n.T("estimation.promptTrumpWait") + "\n")
			}
			return
		case domain.EstimationPhaseBid:
			sb.WriteString(i18n.T("estimation.promptBid") + "\n")
			// **最後の宣言者だけ選べない数がある。** 先に言う。
			if r := e.GetRestrictedBid(); r >= 0 && e.IsHumanBidTurn() {
				sb.WriteString(i18n.Tf("estimation.promptBidRestricted", "n", strconv.Itoa(r)) + "\n")
			}
			return
		case domain.EstimationPhaseRoundEnd:
			sb.WriteString(i18n.T("estimation.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("estimation.promptNext") + "\n")
			return
		}

		currentIdx := e.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("estimation.promptCurrentPlayer",
			"name", cuiPlayerName(e.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("estimation.promptPlay") + "\n")
	})
}

// estimationSuitName スート番号を i18n のスート名に変換する
func estimationSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("estimation.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("estimation.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("estimation.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("estimation.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *EstimationCuiPresenter) HintOutput(e interfaces.EstimationGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("estimation.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		reason := hintReasonStr(hint.Reason, estimationHintReasonKeys)
		switch hint.Reason {
		case "estimationSelectTrump":
			reason = i18n.Tf("estimation.hintReasonSelectTrumpSuit", "suit", estimationSuitName(hint.Value))
		case "estimationBid", "estimationAvoidRestricted":
			reason = i18n.Tf("estimation.hintReasonBidValue",
				"n", strconv.Itoa(hint.Value),
				"why", hintReasonStr(hint.Reason, estimationHintReasonKeys))
		}
		return color.Yellow(i18n.Tf("estimation.hintCall", "reason", reason)) + "\n"
	}
	card := e.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("estimation.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, estimationHintReasonKeys))) + "\n"
}

// estimationHintReasonKeys maps hint-reason identifiers to their i18n keys.
var estimationHintReasonKeys = map[string]string{
	"estimationSelectTrump":     "estimation.hintReasonSelectTrump",
	"estimationBid":             "estimation.hintReasonBid",
	"estimationDashCall":        "estimation.hintReasonDashCall",
	"estimationAvoidRestricted": "estimation.hintReasonAvoidRestricted",
	"estimationWinTrick":        "estimation.hintReasonWinTrick",
	"estimationDuck":            "estimation.hintReasonDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EstimationCuiPresenter) ActionLogOutput(e interfaces.EstimationGame) string {
	return actionLogOutputTextForSeats[*domain.EstimationPlayer](e)
}
