package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// callBreakPlayerStr returns the display string for a single Call Break player.
func callBreakPlayerStr(player *domain.CallBreakPlayer, i int, playable []int) string {
	var b strings.Builder
	bidStr := i18n.T("callbreak.bidPending")
	if player.GetBid() >= 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	b.WriteString(i18n.Tf("callbreak.playerLine",
		"name", cuiPlayerName(player, i),
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		// **Web は cb-bags-counter で常時出しているのに CUI には無かった (#4752)。**
		// バッグの蓄積は長期スコアに直結するので、宣言超過を毎行で見せる。
		"bags", strconv.Itoa(player.GetBags()),
		"cum", domain.FormatCallBreakScore(player.GetCumulativeScore()),
		"round", domain.FormatCallBreakScore(player.GetRoundScore()),
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

// CallBreakCuiPresenter renders the Call Break CUI view.
type CallBreakCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CallBreakCuiPresenter) Output(cb interfaces.CallBreakGame, lastErr error) string {
	return buildCuiOutput(i18n.T("callbreak.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("callbreak.round",
			"round", strconv.Itoa(cb.GetRoundNumber()),
			"trick", strconv.Itoa(cb.GetTrickNumber()),
			"max", strconv.Itoa(cb.GetConfig().MaxRounds)) + "\n")

		if cb.GetSpadesBroken() {
			b.WriteString(i18n.T("callbreak.spadesBrokenYes") + "\n")
		} else {
			b.WriteString(i18n.T("callbreak.spadesBrokenNo") + "\n")
		}

		for i := 0; i < cb.GetPlayerCnt(); i++ {
			// 目印はプレイフェーズで本人の手番のときだけ。
			var playable []int
			if cb.GetPhase() == domain.CallBreakPhasePlay && cb.GetCurrentPlayerIdx() == i {
				playable = cb.GetValidPlayIndices(i)
			}
			b.WriteString(callBreakPlayerStr(cb.GetPlayer(i), i, playable))
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
			banner := i18n.Tf("callbreak.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch cb.GetPhase() {
		case domain.CallBreakPhaseBid:
			bidIdx := cb.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("callbreak.promptBid",
				"name", cuiPlayerName(cb.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("callbreak.promptBidHelp") + "\n")
		case domain.CallBreakPhasePlay:
			currentIdx := cb.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("callbreak.promptPlay",
				"name", cuiPlayerName(cb.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("callbreak.promptPlayHelp") + "\n")
		case domain.CallBreakPhaseTrickEnd:
			b.WriteString(i18n.T("callbreak.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("callbreak.promptTrickEndHelp") + "\n")
		case domain.CallBreakPhaseRoundEnd:
			b.WriteString(i18n.T("callbreak.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("callbreak.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Call Break hint.
func (p *CallBreakCuiPresenter) HintOutput(cb interfaces.CallBreakGame) string {
	hint := cb.GetHint()
	if hint == nil {
		return i18n.T("callbreak.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, callBreakHintReasonKeys)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("callbreak.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("callbreak.hintNone") + "\n"
	}
	card := cb.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("callbreak.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// callBreakHintReasonKeys maps Call Break-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to hintReasonStr.
var callBreakHintReasonKeys = map[string]string{
	"trump_cut": "callbreak.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CallBreakCuiPresenter) ActionLogOutput(cb interfaces.CallBreakGame) string {
	return actionLogOutputTextForSeats[*domain.CallBreakPlayer](cb)
}
