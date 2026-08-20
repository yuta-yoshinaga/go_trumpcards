//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// shelemPlayerStr returns the display string for a single player.
func shelemPlayerStr(player *domain.ShelemPlayer, idx int, declarer bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("shelem.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.ShelemTeamOf(idx)),
		"bid", shelemBidStr(player, declarer),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// shelemBidStr 競りでの立場を短く表す
func shelemBidStr(player *domain.ShelemPlayer, declarer bool) string {
	switch {
	case declarer && player.GetDeclaredShelem():
		return i18n.T("shelem.roleShelem")
	case declarer:
		return i18n.Tf("shelem.roleDeclarer", "n", strconv.Itoa(player.GetBid()))
	case player.GetPassed():
		return i18n.T("shelem.rolePassed")
	case player.GetBid() >= 0:
		return i18n.Tf("shelem.roleBid", "n", strconv.Itoa(player.GetBid()))
	default:
		return i18n.T("shelem.roleActive")
	}
}

// ShelemCuiPresenter renders the Shelem CUI view.
type ShelemCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ShelemCuiPresenter) Output(s interfaces.ShelemGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shelem.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("shelem.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.ShelemTricksPerRound),
			"target", strconv.Itoa(s.GetConfig().Target)) + "\n")
		// **点になるのは A/10/5 だけ。** 盤面から読めないので常時出す。
		sb.WriteString(i18n.T("shelem.pointTable") + "\n")
		sb.WriteString(i18n.Tf("shelem.scoreLine",
			"t0", strconv.Itoa(s.GetScore(0)),
			"t1", strconv.Itoa(s.GetScore(1))) + "\n")

		if s.GetDeclarerIdx() >= 0 {
			if s.GetShelemBid() {
				sb.WriteString(i18n.Tf("shelem.contractShelem",
					"name", cuiPlayerName(s.GetPlayer(s.GetDeclarerIdx()), s.GetDeclarerIdx())) + "\n")
			} else {
				sb.WriteString(i18n.Tf("shelem.contractLine",
					"n", strconv.Itoa(s.GetContract()),
					"name", cuiPlayerName(s.GetPlayer(s.GetDeclarerIdx()), s.GetDeclarerIdx()),
					"got", strconv.Itoa(s.GetRoundPoints(domain.ShelemTeamOf(s.GetDeclarerIdx())))) + "\n")
			}
			// **守備側の点も出す。**契約を阻止できているかは、宣言側の点だけ
			// 見ていても分からない (#5754)。合計は必ず 100 点。
			//
			// **Shelem 宣言のときは出さない。**成否は全トリック取れたかどうか
			// だけで決まり、カード点は一切見ないので、点を出すと「点が要る」と
			// 誤解させる (レビュー指摘 #6098)。
			if !s.GetShelemBid() {
				defenders := 1 - domain.ShelemTeamOf(s.GetDeclarerIdx())
				sb.WriteString(i18n.Tf("shelem.defenderLine",
					"got", strconv.Itoa(s.GetRoundPoints(defenders)),
					"total", strconv.Itoa(domain.ShelemHandPoints)) + "\n")
			}
		} else {
			sb.WriteString(i18n.T("shelem.contractUndecided") + "\n")
		}

		if s.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("shelem.trumpLine", "suit", shelemSuitName(s.GetTrumpSuit())) + "\n")
		}
		if s.GetWidowSize() > 0 {
			sb.WriteString(i18n.Tf("shelem.widowLine", "n", strconv.Itoa(s.GetWidowSize())) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(shelemPlayerStr(s.GetPlayer(i), i, i == s.GetDeclarerIdx()))
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
			switch s.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("shelem.gameEndTeam0", "t0", strconv.Itoa(s.GetScore(0)), "t1", strconv.Itoa(s.GetScore(1)))
			case 1:
				banner = i18n.Tf("shelem.gameEndTeam1", "t0", strconv.Itoa(s.GetScore(0)), "t1", strconv.Itoa(s.GetScore(1)))
			default:
				banner = i18n.Tf("shelem.gameEndTie", "t0", strconv.Itoa(s.GetScore(0)), "t1", strconv.Itoa(s.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.ShelemPhaseBid:
			if s.IsHumanBidTurn() {
				sb.WriteString(i18n.Tf("shelem.promptBid",
					"min", strconv.Itoa(shelemNextBid(s.GetContract()))) + "\n")
			} else {
				sb.WriteString(i18n.T("shelem.promptBidWait") + "\n")
			}
			return
		case domain.ShelemPhaseDiscard:
			if s.IsHumanDiscardTurn() {
				sb.WriteString(i18n.Tf("shelem.promptDiscard", "n", strconv.Itoa(domain.ShelemWidowSize)) + "\n")
			} else {
				sb.WriteString(i18n.T("shelem.promptDiscardWait") + "\n")
			}
			return
		case domain.ShelemPhaseRoundEnd:
			sb.WriteString(i18n.T("shelem.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("shelem.promptNext") + "\n")
			return
		}

		currentIdx := s.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("shelem.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("shelem.promptPlay") + "\n")
	})
}

// shelemSuitName スート番号を i18n のスート名に変換する
func shelemSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("shelem.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("shelem.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("shelem.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("shelem.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *ShelemCuiPresenter) HintOutput(s interfaces.ShelemGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("shelem.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		reason := hintReasonStr(hint.Reason, shelemHintReasonKeys)
		switch hint.Reason {
		case "shelemBid":
			reason = i18n.Tf("shelem.hintReasonBidValue", "n", strconv.Itoa(hint.Value))
		case "shelemDiscard":
			reason = i18n.Tf("shelem.hintReasonDiscardSuit", "suit", shelemSuitName(hint.Suit))
		}
		return color.Yellow(i18n.Tf("shelem.hintCall", "reason", reason)) + "\n"
	}
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("shelem.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, shelemHintReasonKeys))) + "\n"
}

// shelemHintReasonKeys maps hint-reason identifiers to their i18n keys.
var shelemHintReasonKeys = map[string]string{
	"shelemBid":         "shelem.hintReasonBid",
	"shelemPass":        "shelem.hintReasonPass",
	"shelemDiscard":     "shelem.hintReasonDiscard",
	"shelemWinTrick":    "shelem.hintReasonWinTrick",
	"shelemFeedPartner": "shelem.hintReasonFeedPartner",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ShelemCuiPresenter) ActionLogOutput(s interfaces.ShelemGame) string {
	return actionLogOutputText(s)
}
