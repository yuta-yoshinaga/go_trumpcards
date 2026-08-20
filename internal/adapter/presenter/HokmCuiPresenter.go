//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// hokmPlayerStr returns the display string for a single player.
func hokmPlayerStr(player *domain.HokmPlayer, idx int, hakem bool) string {
	var b strings.Builder
	role := ""
	if hakem {
		role = i18n.T("hokm.roleHakem")
	}
	b.WriteString(i18n.Tf("hokm.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.HokmTeamOf(idx)),
		"role", role,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// HokmCuiPresenter renders the Hokm CUI view.
type HokmCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *HokmCuiPresenter) Output(h interfaces.HokmGame, lastErr error) string {
	return buildCuiOutput(i18n.T("hokm.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("hokm.header",
			"hand", strconv.Itoa(h.GetHandNumber()),
			"target", strconv.Itoa(h.GetConfig().Target)) + "\n")
		// **7 トリック先取が肝。** 何トリックで終わるかは盤面から読めない。
		sb.WriteString(i18n.Tf("hokm.raceLine",
			"t0", strconv.Itoa(h.TeamTricks(0)),
			"t1", strconv.Itoa(h.TeamTricks(1)),
			"need", strconv.Itoa(domain.HokmTricksToWin)) + "\n")
		sb.WriteString(i18n.Tf("hokm.scoreLine",
			"t0", strconv.Itoa(h.GetScore(0)),
			"t1", strconv.Itoa(h.GetScore(1))) + "\n")

		if h.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("hokm.trumpLine", "suit", hokmSuitName(h.GetTrumpSuit())) + "\n")
		} else {
			sb.WriteString(i18n.T("hokm.trumpUndecided") + "\n")
		}

		for i := 0; i < h.GetPlayerCnt(); i++ {
			sb.WriteString(hokmPlayerStr(h.GetPlayer(i), i, i == h.GetHakemIdx()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, h.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(h.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if h.GetGameEndFlag() {
			var banner string
			switch h.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("hokm.gameEndTeam0", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			case 1:
				banner = i18n.Tf("hokm.gameEndTeam1", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			default:
				banner = i18n.Tf("hokm.gameEndTie", "t0", strconv.Itoa(h.GetScore(0)), "t1", strconv.Itoa(h.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch h.GetPhase() {
		case domain.HokmPhaseTrump:
			if h.IsHumanTrumpTurn() {
				sb.WriteString(i18n.T("hokm.promptTrump") + "\n")
			} else {
				sb.WriteString(i18n.T("hokm.promptTrumpWait") + "\n")
			}
			return
		case domain.HokmPhaseHandEnd:
			// **Kot は 2 点。** 何が起きたのかを言わないと得点が飛んで見える。
			if h.GetLastHandKot() {
				sb.WriteString(i18n.T("hokm.promptHandEndKot") + "\n")
			} else {
				sb.WriteString(i18n.T("hokm.promptHandEnd") + "\n")
			}
			// **親は負けたときだけ交代する。**次に自分が切り札を選べるかを
			// 左右するのに、次ハンドが始まるまで分からなかった (#5753)。
			if h.GetLastHandHakemChanged() {
				sb.WriteString(i18n.T("hokm.hakemMoves") + "\n")
			} else {
				sb.WriteString(i18n.T("hokm.hakemStays") + "\n")
			}
			sb.WriteString(i18n.T("hokm.promptNext") + "\n")
			return
		}

		currentIdx := h.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("hokm.promptCurrentPlayer",
			"name", cuiPlayerName(h.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("hokm.promptPlay") + "\n")
	})
}

// hokmSuitName スート番号を i18n のスート名に変換する
func hokmSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("hokm.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("hokm.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("hokm.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("hokm.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *HokmCuiPresenter) HintOutput(h interfaces.HokmGame) string {
	hint := h.GetHint()
	if hint == nil {
		return i18n.T("hokm.hintNone") + "\n"
	}
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("hokm.hintTrump",
			"suit", hokmSuitName(hint.Suit))) + "\n"
	}
	card := h.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("hokm.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, hokmHintReasonKeys))) + "\n"
}

// hokmHintReasonKeys maps hint-reason identifiers to their i18n keys.
var hokmHintReasonKeys = map[string]string{
	"hokmDeclareTrump": "hokm.hintReasonDeclareTrump",
	"hokmWinTrick":     "hokm.hintReasonWinTrick",
	"hokmSaveCards":    "hokm.hintReasonSaveCards",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HokmCuiPresenter) ActionLogOutput(h interfaces.HokmGame) string {
	return actionLogOutputTextForSeats[*domain.HokmPlayer](h)
}
