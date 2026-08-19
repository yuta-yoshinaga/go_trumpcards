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

// balootPlayerStr returns the display string for a single player.
func balootPlayerStr(player *domain.BalootPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("baloot.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.BalootTeamOf(idx)),
		"baloot", balootBonusStr(player),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// balootBonusStr Baloot（切り札の K+Q）の有無を短く表す
//
// **開示前の席は「不明」。**配られた瞬間に相手の手の内が割れるのは体験を
// 壊すので、切り札の K か Q が実際に出るまで伏せる (#5750)。
func balootBonusStr(player *domain.BalootPlayer) string {
	if !player.GetBalootRevealed() {
		return i18n.T("baloot.balootHidden")
	}
	if !player.GetHasBaloot() {
		return i18n.T("baloot.balootNone")
	}
	return i18n.Tf("baloot.balootHeld", "points", strconv.Itoa(domain.BalootBonus))
}

// BalootCuiPresenter renders the Baloot CUI view.
type BalootCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BalootCuiPresenter) Output(b interfaces.BalootGame, lastErr error) string {
	return buildCuiOutput(i18n.T("baloot.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("baloot.header",
			"round", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.BalootTricksPerRound),
			"target", strconv.Itoa(b.GetConfig().Target)) + "\n")
		sb.WriteString(i18n.Tf("baloot.scoreLine",
			"t0", strconv.Itoa(b.GetScore(0)),
			"t1", strconv.Itoa(b.GetScore(1))) + "\n")
		sb.WriteString(balootModeLine(b))

		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(balootPlayerStr(b.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, b.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if b.GetGameEndFlag() {
			var banner string
			switch b.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("baloot.gameEndTeam0", "t0", strconv.Itoa(b.GetScore(0)), "t1", strconv.Itoa(b.GetScore(1)))
			case 1:
				banner = i18n.Tf("baloot.gameEndTeam1", "t0", strconv.Itoa(b.GetScore(0)), "t1", strconv.Itoa(b.GetScore(1)))
			default:
				banner = i18n.Tf("baloot.gameEndTie", "t0", strconv.Itoa(b.GetScore(0)), "t1", strconv.Itoa(b.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch b.GetPhase() {
		case domain.BalootPhaseDeclare:
			sb.WriteString(i18n.T("baloot.promptDeclare") + "\n")
			// **親は見送れない。** 選べない選択肢を案内しない。
			if b.GetDealerIdx() == 0 {
				sb.WriteString(i18n.T("baloot.promptDeclareDealer") + "\n")
			} else {
				sb.WriteString(i18n.T("baloot.promptDeclareHelp") + "\n")
			}
			return
		case domain.BalootPhaseRoundEnd:
			sb.WriteString(i18n.T("baloot.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("baloot.promptNext") + "\n")
			return
		}

		currentIdx := b.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("baloot.promptCurrentPlayer",
			"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("baloot.promptPlay") + "\n")
	})
}

// balootModeLine モードと序列の案内。
//
// **序列はモードで入れ替わるので、どちらが有効かを毎回出す。** これを出さないと
// 同じ札の強さが前のラウンドと違う理由が画面から読み取れない。
func balootModeLine(b interfaces.BalootGame) string {
	// **宣言者が居ないのにモードだけ立っている状態を描かない。** 復元した
	// スナップショットでは -1 が入りうる。
	if b.GetDeclarerIdx() < 0 {
		return i18n.T("baloot.modeUndecided") + "\n"
	}
	switch b.GetMode() {
	case domain.BalootModeSun:
		return i18n.Tf("baloot.modeSunLine",
			"name", cuiPlayerName(b.GetPlayer(b.GetDeclarerIdx()), b.GetDeclarerIdx())) + "\n" +
			i18n.T("baloot.orderSun") + "\n"
	case domain.BalootModeHokom:
		return i18n.Tf("baloot.modeHokomLine",
			"suit", balootSuitName(b.GetTrumpSuit()),
			"name", cuiPlayerName(b.GetPlayer(b.GetDeclarerIdx()), b.GetDeclarerIdx())) + "\n" +
			i18n.T("baloot.orderHokom") + "\n"
	default:
		return i18n.T("baloot.modeUndecided") + "\n"
	}
}

// balootSuitName スート番号を i18n のスート名に変換する
func balootSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("baloot.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("baloot.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("baloot.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("baloot.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *BalootCuiPresenter) HintOutput(b interfaces.BalootGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("baloot.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		reason := hintReasonStr(hint.Reason, balootHintReasonKeys)
		if hint.Reason == "balootDeclareHokom" {
			reason = i18n.Tf("baloot.hintReasonDeclareHokomSuit", "suit", balootSuitName(hint.Suit))
		}
		return color.Yellow(i18n.Tf("baloot.hintDeclare", "reason", reason)) + "\n"
	}
	card := b.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("baloot.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, balootHintReasonKeys))) + "\n"
}

// balootHintReasonKeys maps hint-reason identifiers to their i18n keys.
var balootHintReasonKeys = map[string]string{
	"balootDeclareSun":   "baloot.hintReasonDeclareSun",
	"balootDeclareHokom": "baloot.hintReasonDeclareHokom",
	"balootPassDeclare":  "baloot.hintReasonPassDeclare",
	"balootWinTrick":     "baloot.hintReasonWinTrick",
	"balootFeedPartner":  "baloot.hintReasonFeedPartner",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BalootCuiPresenter) ActionLogOutput(b interfaces.BalootGame) string {
	return actionLogOutputText(b)
}
