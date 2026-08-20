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

// sergeantMajorPlayerStr returns the display string for a single player.
func sergeantMajorPlayerStr(
	player *domain.SergeantMajorPlayer, idx int, isDealer, current bool, fromKitty func(*domain.Card) bool,
) string {
	var b strings.Builder
	role := ""
	if isDealer {
		role = i18n.T("sergeantmajor.roleDealer")
	}
	marker := " "
	if current {
		marker = ">"
	}
	// **ノルマと獲得数を並べて出す。** あと何トリック要るかが読めないと打てない。
	b.WriteString(marker + i18n.Tf("sergeantmajor.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"target", strconv.Itoa(player.GetTarget()),
		"took", strconv.Itoa(player.GetTrickCount()),
		"score", strconv.Itoa(player.GetScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(sergeantMajorHandStr(player, fromKitty) + "\n")
	}
	return b.String()
}

// sergeantMajorHandStr は手札を並べ、キティ由来の札に印を付ける。
//
// **取り込むと手札に紛れて見分けが付かなくなる** (#5759)。捨てる 4 枚を選ぶ
// のに「元から持っていた札」と「今入ってきた札」の区別は要る。捨て終われば
// fromKitty が偽になるので、印も自然に消える。
func sergeantMajorHandStr(player *domain.SergeantMajorPlayer, fromKitty func(*domain.Card) bool) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := range player.GetCardsSize() {
		card := player.GetCard(i)
		entry := "[" + strconv.Itoa(i) + "]" + cuiCardStr(card)
		if fromKitty != nil && fromKitty(card) {
			entry = i18n.Tf("sergeantmajor.kittyCard", "card", entry)
		}
		parts = append(parts, entry)
	}
	return strings.Join(parts, "  ")
}

// SergeantMajorCuiPresenter renders the Sergeant Major CUI view.
type SergeantMajorCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SergeantMajorCuiPresenter) Output(s interfaces.SergeantMajorGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sergeantmajor.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("sergeantmajor.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"total", strconv.Itoa(s.GetConfig().Rounds),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.SergeantMajorTricksPerRound)) + "\n")
		// **ノルマは席順で決まる。** 規則そのものを毎回書く。
		sb.WriteString(i18n.T("sergeantmajor.rule") + "\n")

		if s.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("sergeantmajor.trumpLine",
				"suit", sergeantMajorSuitName(s.GetTrumpSuit()),
				"name", cuiPlayerName(s.GetPlayer(s.GetDealerIdx()), s.GetDealerIdx())) + "\n")
		} else {
			sb.WriteString(i18n.Tf("sergeantmajor.trumpUndecided",
				"kitty", strconv.Itoa(s.GetKittySize())) + "\n")
		}

		// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
		if s.GetLastExchange() > 0 {
			sb.WriteString(i18n.Tf("sergeantmajor.exchangeLine",
				"n", strconv.Itoa(s.GetLastExchange())) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(sergeantMajorPlayerStr(s.GetPlayer(i), i,
				i == s.GetDealerIdx(),
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag(),
				s.IsAbsorbedKittyCard))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			switch s.GetWinnerIdx() {
			case 0:
				banner = i18n.T("sergeantmajor.gameEndYou")
			case -1:
				banner = i18n.T("sergeantmajor.gameEndTie")
			default:
				banner = i18n.Tf("sergeantmajor.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.SergeantMajorPhaseTrump:
			if s.IsHumanTrumpTurn() {
				sb.WriteString(i18n.T("sergeantmajor.promptTrump") + "\n")
			} else {
				sb.WriteString(i18n.T("sergeantmajor.promptTrumpWait") + "\n")
			}
		case domain.SergeantMajorPhaseDiscard:
			if s.IsHumanDiscardTurn() {
				sb.WriteString(i18n.T("sergeantmajor.kittyNote") + "\n")
				sb.WriteString(i18n.Tf("sergeantmajor.promptDiscard",
					"kitty", strconv.Itoa(domain.SergeantMajorKittySize),
					"n", strconv.Itoa(s.GetDiscardCount())) + "\n")
			} else {
				sb.WriteString(i18n.T("sergeantmajor.promptDiscardWait") + "\n")
			}
		case domain.SergeantMajorPhaseRoundEnd:
			sb.WriteString(i18n.T("sergeantmajor.promptNext") + "\n")
		default:
			currentIdx := s.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("sergeantmajor.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("sergeantmajor.promptPlay") + "\n")
		}
	})
}

// sergeantMajorSuitName スート番号を i18n のスート名に変換する
func sergeantMajorSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("sergeantmajor.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("sergeantmajor.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("sergeantmajor.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("sergeantmajor.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *SergeantMajorCuiPresenter) HintOutput(s interfaces.SergeantMajorGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("sergeantmajor.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, sergeantMajorHintReasonKeys)
	switch {
	case len(hint.Indices) > 0:
		// 捨て札の助言は複数の札を指す。
		parts := make([]string, 0, len(hint.Indices))
		for _, i := range hint.Indices {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(s.GetPlayer(0).GetCard(i)))
		}
		return color.Yellow(i18n.Tf("sergeantmajor.hintDiscard",
			"cards", strings.Join(parts, " "), "reason", reason)) + "\n"
	case hint.CardIndex == nil:
		// **宣言の助言は札ではなくスートを指す。**
		return color.Yellow(i18n.Tf("sergeantmajor.hintTrump",
			"suit", sergeantMajorSuitName(hint.Suit), "reason", reason)) + "\n"
	default:
		card := s.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("sergeantmajor.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
}

// sergeantMajorHintReasonKeys maps hint-reason identifiers to their i18n keys.
var sergeantMajorHintReasonKeys = map[string]string{
	"sergeantmajorSelectTrump": "sergeantmajor.hintReasonSelectTrump",
	"sergeantmajorDiscard":     "sergeantmajor.hintReasonDiscard",
	"sergeantmajorWinTrick":    "sergeantmajor.hintReasonWinTrick",
	"sergeantmajorPressOn":     "sergeantmajor.hintReasonPressOn",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SergeantMajorCuiPresenter) ActionLogOutput(s interfaces.SergeantMajorGame) string {
	return actionLogOutputTextWithNames(s, func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) })
}
