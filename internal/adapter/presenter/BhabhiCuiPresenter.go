//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// bhabhiPlayerStr returns the display string for a single player.
func bhabhiPlayerStr(player *domain.BhabhiPlayer, idx int, current bool) string {
	var b strings.Builder
	state := i18n.Tf("bhabhi.stateCards", "n", strconv.Itoa(player.GetCardsSize()))
	if player.IsOut() {
		// **上がった順位は強さではない。** 最後に残った 1 人だけが負け。
		state = i18n.Tf("bhabhi.stateOut", "rank", strconv.Itoa(player.GetRank()))
	}
	marker := " "
	if current {
		marker = ">"
	}
	b.WriteString(marker + i18n.Tf("bhabhi.playerLine",
		"name", cuiPlayerName(player, idx),
		"state", state,
		"pickups", strconv.Itoa(player.GetPickups()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BhabhiCuiPresenter renders the Bhabhi CUI view.
type BhabhiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BhabhiCuiPresenter) Output(b interfaces.BhabhiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bhabhi.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("bhabhi.header",
			"trick", strconv.Itoa(b.GetTrickNumber()+1),
			"players", strconv.Itoa(b.GetPlayerCnt()),
			"alive", strconv.Itoa(b.GetAliveCount())) + "\n")
		// **勝者ではなく敗者を決めるゲーム。** 目的を毎回書く。
		sb.WriteString(i18n.T("bhabhi.rule") + "\n")

		if b.GetLeadSuit() > 0 {
			sb.WriteString(i18n.Tf("bhabhi.leadLine",
				"suit", bhabhiSuitName(b.GetLeadSuit()),
				"n", strconv.Itoa(len(b.GetPile()))) + "\n")
		} else {
			sb.WriteString(i18n.T("bhabhi.leadNone") + "\n")
		}

		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(bhabhiPlayerStr(b.GetPlayer(i), i, i == b.GetCurrentPlayerIdx() && !b.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, b.GetPile(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		// **直前の引き取りは盤面に痕跡が残らない。** 何枚どこへ行ったか言う。
		if b.GetLastPickupIdx() >= 0 && !b.GetGameEndFlag() {
			sb.WriteString(i18n.Tf("bhabhi.lastPickup",
				"name", cuiPlayerName(b.GetPlayer(b.GetLastPickupIdx()), b.GetLastPickupIdx()),
				"n", strconv.Itoa(b.GetLastPickupSize())) + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		if b.GetGameEndFlag() {
			idx := b.GetBhabhiIdx()
			name := "?"
			if idx >= 0 {
				name = cuiPlayerName(b.GetPlayer(idx), idx)
			}
			var banner string
			switch {
			case b.IsStalemate():
				// **膠着で終わったことは盤面から読めない。**
				banner = i18n.Tf("bhabhi.gameEndStalemate",
					"name", name, "tricks", strconv.Itoa(b.GetTrickNumber()))
			case idx == 0:
				banner = i18n.Tf("bhabhi.gameEndYou", "name", name)
			default:
				banner = i18n.Tf("bhabhi.gameEndCpu", "name", name)
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := b.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("bhabhi.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("bhabhi.promptPlay") + "\n")
	})
}

// bhabhiSuitName スート番号を i18n のスート名に変換する
func bhabhiSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("bhabhi.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("bhabhi.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("bhabhi.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("bhabhi.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *BhabhiCuiPresenter) HintOutput(b interfaces.BhabhiGame) string {
	hint := b.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("bhabhi.hintNone") + "\n"
	}
	card := b.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("bhabhi.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, bhabhiHintReasonKeys))) + "\n"
}

// bhabhiHintReasonKeys maps hint-reason identifiers to their i18n keys.
var bhabhiHintReasonKeys = map[string]string{
	"bhabhiLead":     "bhabhi.hintReasonLead",
	"bhabhiDuck":     "bhabhi.hintReasonDuck",
	"bhabhiDumpHigh": "bhabhi.hintReasonDumpHigh",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BhabhiCuiPresenter) ActionLogOutput(b interfaces.BhabhiGame) string {
	return actionLogOutputTextWithNames(b, func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) })
}
