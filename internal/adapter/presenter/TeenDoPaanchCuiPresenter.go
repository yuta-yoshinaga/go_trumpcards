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

// teenDoPaanchPlayerStr returns the display string for a single player.
func teenDoPaanchPlayerStr(player *domain.TeenDoPaanchPlayer, idx int, isFive, current bool) string {
	var b strings.Builder
	role := ""
	if isFive {
		role = i18n.T("teendopaanch.roleFive")
	}
	marker := " "
	if current {
		marker = ">"
	}
	// **ノルマと獲得数を並べて出す。** 多く取ってもうれしくないゲームなので、
	// 「あと何トリック要るか」が読めないと打ち方が決まらない。
	b.WriteString(marker + i18n.Tf("teendopaanch.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"target", strconv.Itoa(player.GetTarget()),
		"took", strconv.Itoa(player.GetTrickCount()),
		"met", strconv.Itoa(player.GetMet()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// TeenDoPaanchCuiPresenter renders the 3-2-5 CUI view.
type TeenDoPaanchCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TeenDoPaanchCuiPresenter) Output(g interfaces.TeenDoPaanchGame, lastErr error) string {
	return buildCuiOutput(i18n.T("teendopaanch.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("teendopaanch.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetConfig().Rounds),
			"trick", strconv.Itoa(g.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.TeenDoPaanchTricksPerRound)) + "\n")
		// **ノルマは宣言ではなく割り当て。** 規則そのものを毎回書く。
		sb.WriteString(i18n.T("teendopaanch.rule") + "\n")

		if g.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("teendopaanch.trumpLine",
				"suit", teenDoPaanchSuitName(g.GetTrumpSuit()),
				"name", cuiPlayerName(g.GetPlayer(g.GetFivePlayerIdx()), g.GetFivePlayerIdx())) + "\n")
		} else {
			sb.WriteString(i18n.Tf("teendopaanch.trumpUndecided",
				"seen", strconv.Itoa(domain.TeenDoPaanchFirstDeal)) + "\n")
		}

		// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
		if g.GetLastExchange() > 0 {
			line := i18n.Tf("teendopaanch.exchangeLine", "n", strconv.Itoa(g.GetLastExchange()))
			// **合計だけでは、自分の手札から何が抜かれたのか分からない** (#5757)。
			// 誰の最強札が誰に渡ったのかがこのゲームの名物。
			if pairs := g.GetLastExchangePairs(); len(pairs) > 0 {
				parts := make([]string, 0, len(pairs))
				for _, ex := range pairs {
					parts = append(parts, i18n.Tf("teendopaanch.exchangePair",
						"giver", cuiPlayerName(g.GetPlayer(ex.Giver), ex.Giver),
						"taker", cuiPlayerName(g.GetPlayer(ex.Taker), ex.Taker),
						"n", strconv.Itoa(ex.Count)))
				}
				line += i18n.Tf("teendopaanch.exchangeDetail", "pairs", strings.Join(parts, ", "))
			}
			sb.WriteString(line + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			sb.WriteString(teenDoPaanchPlayerStr(g.GetPlayer(i), i,
				i == g.GetFivePlayerIdx(),
				i == g.GetCurrentPlayerIdx() && !g.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if g.GetGameEndFlag() {
			var banner string
			switch g.GetWinnerIdx() {
			case 0:
				banner = i18n.T("teendopaanch.gameEndYou")
			case -1:
				banner = i18n.T("teendopaanch.gameEndTie")
			default:
				banner = i18n.Tf("teendopaanch.gameEndCpu",
					"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.TeenDoPaanchPhaseTrump:
			if g.IsHumanTrumpTurn() {
				sb.WriteString(i18n.Tf("teendopaanch.promptTrump",
					"seen", strconv.Itoa(domain.TeenDoPaanchFirstDeal)) + "\n")
			} else {
				sb.WriteString(i18n.T("teendopaanch.promptTrumpWait") + "\n")
			}
		case domain.TeenDoPaanchPhaseRoundEnd:
			sb.WriteString(i18n.T("teendopaanch.promptNext") + "\n")
		default:
			currentIdx := g.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("teendopaanch.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("teendopaanch.promptPlay") + "\n")
		}
	})
}

// teenDoPaanchSuitName スート番号を i18n のスート名に変換する
func teenDoPaanchSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("teendopaanch.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("teendopaanch.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("teendopaanch.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("teendopaanch.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *TeenDoPaanchCuiPresenter) HintOutput(g interfaces.TeenDoPaanchGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("teendopaanch.hintNone") + "\n"
	}
	// **宣言フェーズの助言は札ではなくスートを指す。**
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("teendopaanch.hintTrump",
			"suit", teenDoPaanchSuitName(hint.Suit),
			"reason", hintReasonStr(hint.Reason, teenDoPaanchHintReasonKeys))) + "\n"
	}
	card := g.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("teendopaanch.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, teenDoPaanchHintReasonKeys))) + "\n"
}

// teenDoPaanchHintReasonKeys maps hint-reason identifiers to their i18n keys.
var teenDoPaanchHintReasonKeys = map[string]string{
	"teendopaanchSelectTrump": "teendopaanch.hintReasonSelectTrump",
	"teendopaanchWinTrick":    "teendopaanch.hintReasonWinTrick",
	"teendopaanchDuck":        "teendopaanch.hintReasonDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TeenDoPaanchCuiPresenter) ActionLogOutput(g interfaces.TeenDoPaanchGame) string {
	return actionLogOutputText(g)
}
