package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// batakPlayerStr returns the display string for a single Batak player.
func batakPlayerStr(player *domain.BatakPlayer, i int, playable []int, isDeclarer bool) string {
	var b strings.Builder
	bidStr := i18n.T("batak.bidPending")
	if player.GetBid() == domain.BatakPassBid {
		bidStr = i18n.T("batak.bidPass")
	} else if player.GetBid() > 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	roleStr := ""
	if isDeclarer {
		roleStr = " " + i18n.T("batak.roleDeclarer")
	}
	b.WriteString(i18n.Tf("batak.playerLine",
		"name", cuiPlayerName(player, i),
		"role", roleStr,
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **Web は validPlayIndices で出せない札を無効化しているのに、CUI は素の
		// 一覧だけだった。**マストフォロー/マストトランプで何が出せるかは、番号を
		// 打ってエラーを踏むまで分からなかった (#5605)。playable が空のときは
		// 目印を付けない (制限が決まっていない状態と区別する)。
		b.WriteString(cuiPlayableMarkedCardListStr(player, playable) + "\n")
	}
	return b.String()
}

// BatakCuiPresenter renders the Batak CUI view.
type BatakCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BatakCuiPresenter) Output(cb interfaces.BatakGame, lastErr error) string {
	return buildCuiOutput(i18n.T("batak.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("batak.round",
			"round", strconv.Itoa(cb.GetRoundNumber()),
			"trick", strconv.Itoa(cb.GetTrickNumber()),
			"max", strconv.Itoa(cb.GetConfig().MaxRounds)) + "\n")

		declarerIdx := cb.GetDeclarerIdx()
		declarerName := i18n.T("batak.declarerUndecided")
		highBidStr := "-"
		if declarerIdx >= 0 {
			declarerName = cuiPlayerName(cb.GetPlayer(declarerIdx), declarerIdx)
			highBidStr = strconv.Itoa(cb.GetHighBid())
		} else if cb.GetHighBid() > 0 {
			highBidStr = strconv.Itoa(cb.GetHighBid())
		}
		b.WriteString(i18n.Tf("batak.declarer", "declarer", declarerName, "bid", highBidStr) + "\n")
		b.WriteString(i18n.T("batak.scoringRule") + "\n")

		if cb.GetSpadesBroken() {
			b.WriteString(i18n.T("batak.spadesBrokenYes") + "\n")
		} else {
			b.WriteString(i18n.T("batak.spadesBrokenNo") + "\n")
		}

		for i := 0; i < cb.GetPlayerCnt(); i++ {
			// 目印はプレイフェーズで本人の手番のときだけ。
			var playable []int
			if cb.GetPhase() == domain.BatakPhasePlay && cb.GetCurrentPlayerIdx() == i {
				playable = cb.GetValidPlayIndices(i)
			}
			isDeclarer := declarerIdx == i
			b.WriteString(batakPlayerStr(cb.GetPlayer(i), i, playable, isDeclarer))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := cb.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(cb.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if cb.GetGameEndFlag() {
			winnerIdx := cb.GetWinnerIdx()
			player := cb.GetPlayer(winnerIdx)
			banner := i18n.Tf("batak.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch cb.GetPhase() {
		case domain.BatakPhaseBid:
			bidIdx := cb.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("batak.promptBid",
				"name", cuiPlayerName(cb.GetPlayer(bidIdx), bidIdx)) + "\n")
			minLegal := cb.MinLegalBid()
			if minLegal == domain.BatakPassBid {
				b.WriteString(i18n.T("batak.promptBidPassOnly") + "\n")
			} else {
				b.WriteString(i18n.Tf("batak.promptBidLegal", "min", strconv.Itoa(minLegal)) + "\n")
			}
		case domain.BatakPhasePlay:
			currentIdx := cb.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("batak.promptPlay",
				"name", cuiPlayerName(cb.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("batak.promptPlayHelp") + "\n")
		case domain.BatakPhaseTrickEnd:
			b.WriteString(i18n.T("batak.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("batak.promptTrickEndHelp") + "\n")
		case domain.BatakPhaseRoundEnd:
			b.WriteString(i18n.T("batak.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("batak.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Batak hint.
func (p *BatakCuiPresenter) HintOutput(cb interfaces.BatakGame) string {
	hint := cb.GetHint()
	if hint == nil {
		return i18n.T("batak.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, batakHintReasonKeys)
	if hint.Bid != nil {
		if *hint.Bid == domain.BatakPassBid {
			return color.Yellow(i18n.Tf("batak.hintPass",
				"reason", reason)) + "\n"
		}
		return color.Yellow(i18n.Tf("batak.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("batak.hintNone") + "\n"
	}
	card := cb.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("batak.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// batakHintReasonKeys maps Batak-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to hintReasonStr.
var batakHintReasonKeys = map[string]string{
	"pass_weak_hand": "batak.hintReasonPassWeakHand",
	"trump_cut":      "batak.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BatakCuiPresenter) ActionLogOutput(cb interfaces.BatakGame) string {
	return actionLogOutputTextForSeats[*domain.BatakPlayer](cb)
}
