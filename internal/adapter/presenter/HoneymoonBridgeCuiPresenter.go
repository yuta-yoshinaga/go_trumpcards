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

// honeymoonBridgePlayerStr returns the display string for a single player.
func honeymoonBridgePlayerStr(player *domain.HoneymoonBridgePlayer, idx int, isDeclarer, current bool) string {
	var b strings.Builder
	role := ""
	if isDeclarer {
		role = i18n.T("honeymoonbridge.roleDeclarer")
	}
	marker := " "
	if current {
		marker = ">"
	}
	b.WriteString(marker + i18n.Tf("honeymoonbridge.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"took", strconv.Itoa(player.GetTrickCount()),
		"score", strconv.Itoa(player.GetScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// HoneymoonBridgeCuiPresenter renders the Honeymoon Bridge CUI view.
type HoneymoonBridgeCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *HoneymoonBridgeCuiPresenter) Output(s interfaces.HoneymoonBridgeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("honeymoonbridge.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("honeymoonbridge.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"target", strconv.Itoa(s.GetConfig().Target),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.HoneymoonBridgeTricksPerPhase)) + "\n")
		// **前半と後半で意味が変わる。** 規則そのものを毎回書く。
		sb.WriteString(i18n.T("honeymoonbridge.rule") + "\n")

		if s.GetPhase() == domain.HoneymoonBridgePhaseDraw {
			sb.WriteString(i18n.Tf("honeymoonbridge.stockLine",
				"stock", strconv.Itoa(s.GetStockSize())) + "\n")
		}

		if s.GetContractLevel() > 0 {
			sb.WriteString(i18n.Tf("honeymoonbridge.contractLine",
				"level", strconv.Itoa(s.GetContractLevel()),
				"suit", honeymoonBridgeSuitName(s.GetTrumpSuit()),
				"name", cuiPlayerName(s.GetPlayer(s.GetDeclarerIdx()), s.GetDeclarerIdx()),
				"need", strconv.Itoa(s.RequiredTricks())) + "\n")
		} else {
			sb.WriteString(i18n.T("honeymoonbridge.contractUndecided") + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(honeymoonBridgePlayerStr(s.GetPlayer(i), i,
				i == s.GetDeclarerIdx(),
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
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
				banner = i18n.T("honeymoonbridge.gameEndYou")
			case -1:
				banner = i18n.T("honeymoonbridge.gameEndTie")
			default:
				banner = i18n.Tf("honeymoonbridge.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.HoneymoonBridgePhaseBid:
			if s.IsHumanBidTurn() {
				// **通る最小の宣言を出す。** これを出さないと、拒否される値を
				// 何度も打ち込むはめになる。
				level, suit := s.NextBid()
				if level == 0 {
					sb.WriteString(i18n.T("honeymoonbridge.promptBidCapped") + "\n")
				} else {
					sb.WriteString(i18n.Tf("honeymoonbridge.promptBid",
						"level", strconv.Itoa(level),
						"suit", honeymoonBridgeSuitName(suit)) + "\n")
				}
			} else {
				sb.WriteString(i18n.T("honeymoonbridge.promptBidWait") + "\n")
			}
		case domain.HoneymoonBridgePhaseRoundEnd:
			// **得点式は細かい (契約レベル×10 + オーバートリック×5 / 失敗は
			// 不足×10)。**トリックの過不足だけでは何点動いたか読めない (#5760)。
			if s.GetContractLevel() > 0 {
				key := "honeymoonbridge.roundDown"
				if s.GetLastMade() {
					key = "honeymoonbridge.roundMade"
				}
				sb.WriteString(i18n.Tf(key,
					"need", strconv.Itoa(s.RequiredTricks()),
					"took", strconv.Itoa(s.GetLastTricks()),
					"points", strconv.Itoa(s.GetLastPoints())) + "\n")
			}
			sb.WriteString(i18n.T("honeymoonbridge.promptNext") + "\n")
		default:
			currentIdx := s.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("honeymoonbridge.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("honeymoonbridge.promptPlay") + "\n")
		}
	})
}

// honeymoonBridgeSuitName スート番号を i18n のスート名に変換する。
//
// **0 はノートランプ。** 「未決定」ではなく、競りで選べる 5 つ目の選択肢です。
func honeymoonBridgeSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("honeymoonbridge.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("honeymoonbridge.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("honeymoonbridge.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("honeymoonbridge.suitDiamond")
	default:
		return i18n.T("honeymoonbridge.suitNoTrump")
	}
}

// HintOutput emits the current hint.
func (p *HoneymoonBridgeCuiPresenter) HintOutput(s interfaces.HoneymoonBridgeGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("honeymoonbridge.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, honeymoonBridgeHintReasonKeys)
	if hint.CardIndex == nil {
		// **競りの助言は札ではなく契約を指す。**
		if hint.Level == 0 {
			return color.Yellow(i18n.Tf("honeymoonbridge.hintPass", "reason", reason)) + "\n"
		}
		return color.Yellow(i18n.Tf("honeymoonbridge.hintBid",
			"level", strconv.Itoa(hint.Level),
			"suit", honeymoonBridgeSuitName(hint.Suit),
			"reason", reason)) + "\n"
	}
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("honeymoonbridge.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// honeymoonBridgeHintReasonKeys maps hint-reason identifiers to their i18n keys.
var honeymoonBridgeHintReasonKeys = map[string]string{
	"honeymoonbridgeDraw":     "honeymoonbridge.hintReasonDraw",
	"honeymoonbridgeBid":      "honeymoonbridge.hintReasonBid",
	"honeymoonbridgePass":     "honeymoonbridge.hintReasonPass",
	"honeymoonbridgeWinTrick": "honeymoonbridge.hintReasonWinTrick",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HoneymoonBridgeCuiPresenter) ActionLogOutput(s interfaces.HoneymoonBridgeGame) string {
	return actionLogOutputTextForSeats[*domain.HoneymoonBridgePlayer](s)
}
