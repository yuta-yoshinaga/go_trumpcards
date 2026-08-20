//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// thirtyOneSuitSymbol maps a card design to its (locale-independent) suit symbol.
func thirtyOneSuitSymbol(suit int) string {
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

// thirtyOnePlayerStr returns the display string for a single ThirtyOne player.
func thirtyOnePlayerStr(g interfaces.ThirtyOneGame, player *domain.ThirtyOnePlayer, i int) string {
	var b strings.Builder
	reveal := thirtyOneReveal(g)
	score := "?"
	if player.GetIsHuman() || reveal {
		score = strconv.Itoa(player.BestSuitScore())
	}
	if player.IsEliminated() {
		score = "OUT"
	}
	b.WriteString(i18n.Tf("thirtyone.playerLine",
		"name", cuiPlayerName(player, i),
		"lives", strconv.Itoa(player.GetLives()),
		"score", score,
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
		// Per-suit totals mirror the web suit-score badges so the human can see
		// which suit to build toward 31 and which cards to discard.
		scores := player.SuitScores()
		b.WriteString(i18n.Tf("thirtyone.suitBreakdown",
			"spade", strconv.Itoa(scores[domain.CardDesignSpade]),
			"clover", strconv.Itoa(scores[domain.CardDesignClover]),
			"heart", strconv.Itoa(scores[domain.CardDesignHeart]),
			"diamond", strconv.Itoa(scores[domain.CardDesignDiamond]),
			"best", thirtyOneSuitSymbol(player.BestSuit())) + "\n")
	}
	return b.String()
}

// ThirtyOneCuiPresenter renders the ThirtyOne CUI view.
type ThirtyOneCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ThirtyOneCuiPresenter) Output(g interfaces.ThirtyOneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("thirtyone.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("thirtyone.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		// **難易度の違いはこの数字がすべて。**Easy/Normal/Hard という名前だけでは、
		// 何が変わるのか体験からしか分からなかった (#5623)。
		b.WriteString(i18n.Tf("thirtyone.knockThresholdLine",
			"score", strconv.Itoa(g.GetCpuKnockThreshold())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("thirtyone.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(thirtyOnePlayerStr(g, g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("thirtyone.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		// The best-suit total is the win condition (31), so surface the human's
		// running total and announce a knock (which makes the round's last turn).
		for i := 0; i < g.GetPlayerCnt(); i++ {
			if hp := g.GetPlayer(i); hp.GetIsHuman() {
				b.WriteString(i18n.Tf("thirtyone.yourBestSuit", "score", strconv.Itoa(hp.BestSuitScore())) + "\n")
				break
			}
		}
		if knocker := g.GetKnockerIdx(); knocker >= 0 {
			b.WriteString(color.Yellow(i18n.Tf("thirtyone.knockNotice",
				"name", cuiPlayerName(g.GetPlayer(knocker), knocker))) + "\n")
		}

		switch g.GetPhase() {
		case domain.ThirtyOnePhaseDraw:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("thirtyone.promptDraw", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("thirtyone.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("thirtyone.promptDrawHelpDiscard") + "\n")
			if g.GetKnockerIdx() < 0 {
				b.WriteString(i18n.T("thirtyone.promptKnockHelp") + "\n")
			}
		case domain.ThirtyOnePhaseDiscard:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("thirtyone.promptDiscard", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("thirtyone.promptDiscardHelp") + "\n")
		case domain.ThirtyOnePhaseRoundEnd:
			if g.GetThirtyOneIdx() >= 0 {
				b.WriteString(i18n.Tf("thirtyone.promptThirtyOne",
					"name", cuiPlayerName(g.GetPlayer(g.GetThirtyOneIdx()), g.GetThirtyOneIdx())) + "\n")
			}
			b.WriteString(i18n.T("thirtyone.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("thirtyone.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ThirtyOneCuiPresenter) ActionLogOutput(g interfaces.ThirtyOneGame) string {
	return actionLogOutputTextForSeats[*domain.ThirtyOnePlayer](g)
}

// HintOutput emits the recommended move for the human player.
//
// **CPU と同じ材料で判断する。**ドロー先も捨て札も、ドメインの GetHint が
// cpuWantsDiscard / bestDropIndex / ノック閾値をそのまま通した結果を出す。
// Web はクライアント側ヒントを持っていたが、ネイティブ CUI には何も無かった (#4806)。
func (p *ThirtyOneCuiPresenter) HintOutput(g interfaces.ThirtyOneGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("thirtyone.hintNone") + "\n"
	}
	if hint.Action == "discard" && hint.CardIndex >= 0 {
		idx := g.GetCurrentPlayerIdx()
		card := g.GetPlayer(idx).GetCard(hint.CardIndex)
		return color.Yellow(i18n.Tf("thirtyone.hintDiscard",
			"idx", strconv.Itoa(hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", i18n.T("thirtyone.hintReason."+hint.Reason))) + "\n"
	}
	return color.Yellow(i18n.Tf("thirtyone.hintAction",
		"action", i18n.T("thirtyone.hintAction."+hint.Action),
		"reason", i18n.T("thirtyone.hintReason."+hint.Reason))) + "\n"
}
