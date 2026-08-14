//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tarabishPlayerStr returns the display string for a single player.
func tarabishPlayerStr(player *domain.TarabishPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tarabish.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.TarabishTeamOf(idx)),
		"meld", tarabishMeldStr(player),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// tarabishMeldStr メルドの内訳を短く表す
func tarabishMeldStr(player *domain.TarabishPlayer) string {
	if player.GetMeldPoints() == 0 {
		return i18n.T("tarabish.meldNone")
	}
	parts := make([]string, 0, 2)
	if n := player.GetRunLen(); n > 0 {
		parts = append(parts, i18n.Tf("tarabish.meldRun", "len", strconv.Itoa(n)))
	}
	if player.GetHasBella() {
		parts = append(parts, i18n.T("tarabish.meldBella"))
	}
	return i18n.Tf("tarabish.meldSummary",
		"detail", strings.Join(parts, "+"),
		"points", strconv.Itoa(player.GetMeldPoints()))
}

// TarabishCuiPresenter renders the Tarabish CUI view.
type TarabishCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TarabishCuiPresenter) Output(t interfaces.TarabishGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tarabish.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("tarabish.header",
			"round", strconv.Itoa(t.GetRoundNumber()),
			"trick", strconv.Itoa(t.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.TarabishTricksPerRound),
			"target", strconv.Itoa(t.GetConfig().Target)) + "\n")
		sb.WriteString(i18n.Tf("tarabish.scoreLine",
			"t0", strconv.Itoa(t.GetScore(0)),
			"t1", strconv.Itoa(t.GetScore(1))) + "\n")
		// **切り札の序列はこの系統の肝。** 常時出す。
		sb.WriteString(i18n.T("tarabish.orderLine") + "\n")

		if up := t.GetUpCard(); up != nil && t.GetTrumpTakerIdx() < 0 {
			sb.WriteString(i18n.Tf("tarabish.upCardLine", "card", cuiCardStr(up)) + "\n")
		} else if t.GetTrumpTakerIdx() >= 0 {
			sb.WriteString(i18n.Tf("tarabish.trumpLine",
				"suit", tarabishSuitName(t.GetTrumpSuit()),
				"name", cuiPlayerName(t.GetPlayer(t.GetTrumpTakerIdx()), t.GetTrumpTakerIdx())) + "\n")
		}

		for i := 0; i < t.GetPlayerCnt(); i++ {
			sb.WriteString(tarabishPlayerStr(t.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, t.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(t.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if t.GetGameEndFlag() {
			var banner string
			switch t.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("tarabish.gameEndTeam0", "t0", strconv.Itoa(t.GetScore(0)), "t1", strconv.Itoa(t.GetScore(1)))
			case 1:
				banner = i18n.Tf("tarabish.gameEndTeam1", "t0", strconv.Itoa(t.GetScore(0)), "t1", strconv.Itoa(t.GetScore(1)))
			default:
				banner = i18n.Tf("tarabish.gameEndTie", "t0", strconv.Itoa(t.GetScore(0)), "t1", strconv.Itoa(t.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch t.GetPhase() {
		case domain.TarabishPhaseBid:
			sb.WriteString(i18n.T("tarabish.promptBid") + "\n")
			// **親は見送れない。** 選べない選択肢を案内しない。
			if t.GetDealerIdx() == 0 {
				sb.WriteString(i18n.T("tarabish.promptBidDealer") + "\n")
			} else {
				sb.WriteString(i18n.T("tarabish.promptBidHelp") + "\n")
			}
			return
		case domain.TarabishPhaseRoundEnd:
			sb.WriteString(i18n.T("tarabish.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("tarabish.promptNext") + "\n")
			return
		}

		currentIdx := t.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("tarabish.promptCurrentPlayer",
			"name", cuiPlayerName(t.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("tarabish.promptPlay") + "\n")
	})
}

// tarabishSuitName スート番号を i18n のスート名に変換する
func tarabishSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("tarabish.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("tarabish.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("tarabish.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("tarabish.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *TarabishCuiPresenter) HintOutput(t interfaces.TarabishGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("tarabish.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("tarabish.hintBid",
			"reason", hintReasonStr(hint.Reason, tarabishHintReasonKeys))) + "\n"
	}
	card := t.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("tarabish.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, tarabishHintReasonKeys))) + "\n"
}

// tarabishHintReasonKeys maps hint-reason identifiers to their i18n keys.
var tarabishHintReasonKeys = map[string]string{
	"tarabishTakeTrump":   "tarabish.hintReasonTakeTrump",
	"tarabishPassTrump":   "tarabish.hintReasonPassTrump",
	"tarabishWinTrick":    "tarabish.hintReasonWinTrick",
	"tarabishFeedPartner": "tarabish.hintReasonFeedPartner",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TarabishCuiPresenter) ActionLogOutput(t interfaces.TarabishGame) string {
	return actionLogOutputText(t)
}
