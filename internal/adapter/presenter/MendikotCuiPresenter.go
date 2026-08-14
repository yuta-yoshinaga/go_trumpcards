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

// mendikotPlayerStr returns the display string for a single player.
func mendikotPlayerStr(player *domain.MendikotPlayer, idx int, chooser bool) string {
	var b strings.Builder
	role := ""
	if chooser {
		role = i18n.T("mendikot.roleChooser")
	}
	b.WriteString(i18n.Tf("mendikot.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.MendikotTeamOf(idx)),
		"role", role,
		"tens", strconv.Itoa(player.GetTens()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// MendikotCuiPresenter renders the Mendikot CUI view.
type MendikotCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MendikotCuiPresenter) Output(m interfaces.MendikotGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mendikot.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("mendikot.header",
			"hand", strconv.Itoa(m.GetHandNumber()),
			"trick", strconv.Itoa(m.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.MendikotTricksPerRound),
			"target", strconv.Itoa(m.GetConfig().Target)) + "\n")
		// **勝敗を決めるのは 10 の枚数。** 盤面から読めないので常時出す。
		sb.WriteString(i18n.Tf("mendikot.tensLine",
			"t0", strconv.Itoa(m.TeamTens(0)),
			"t1", strconv.Itoa(m.TeamTens(1)),
			"total", strconv.Itoa(domain.MendikotTensInDeck)) + "\n")
		sb.WriteString(i18n.T("mendikot.rule") + "\n")
		sb.WriteString(i18n.Tf("mendikot.trickLine",
			"t0", strconv.Itoa(m.TeamTricks(0)),
			"t1", strconv.Itoa(m.TeamTricks(1))) + "\n")
		sb.WriteString(i18n.Tf("mendikot.scoreLine",
			"t0", strconv.Itoa(m.GetScore(0)),
			"t1", strconv.Itoa(m.GetScore(1))) + "\n")

		if m.GetTrumpSuit() > 0 {
			sb.WriteString(i18n.Tf("mendikot.trumpLine",
				"suit", mendikotSuitName(m.GetTrumpSuit()),
				"name", cuiPlayerName(m.GetPlayer(m.GetTrumpChooserIdx()), m.GetTrumpChooserIdx())) + "\n")
		} else {
			sb.WriteString(i18n.T("mendikot.trumpUndecided") + "\n")
		}

		for i := 0; i < m.GetPlayerCnt(); i++ {
			sb.WriteString(mendikotPlayerStr(m.GetPlayer(i), i, i == m.GetTrumpChooserIdx()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, m.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(m.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if m.GetGameEndFlag() {
			var banner string
			switch m.GetWinnerTeam() {
			case 0:
				banner = i18n.Tf("mendikot.gameEndTeam0", "t0", strconv.Itoa(m.GetScore(0)), "t1", strconv.Itoa(m.GetScore(1)))
			case 1:
				banner = i18n.Tf("mendikot.gameEndTeam1", "t0", strconv.Itoa(m.GetScore(0)), "t1", strconv.Itoa(m.GetScore(1)))
			default:
				banner = i18n.Tf("mendikot.gameEndTie", "t0", strconv.Itoa(m.GetScore(0)), "t1", strconv.Itoa(m.GetScore(1)))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if m.GetPhase() == domain.MendikotPhaseHandEnd {
			// **決まり方で勝ち点が 1/2/3 と変わる。** どれだったかを言う。
			sb.WriteString(i18n.Tf("mendikot.handEnd."+m.GetLastHandKind(),
				"team", strconv.Itoa(m.GetLastHandWinner())) + "\n")
			sb.WriteString(i18n.T("mendikot.promptNext") + "\n")
			return
		}

		currentIdx := m.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("mendikot.promptCurrentPlayer",
			"name", cuiPlayerName(m.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("mendikot.promptPlay") + "\n")
	})
}

// mendikotSuitName スート番号を i18n のスート名に変換する
func mendikotSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("mendikot.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("mendikot.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("mendikot.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("mendikot.suitDiamond")
	default:
		return "?"
	}
}

// HintOutput emits the current hint.
func (p *MendikotCuiPresenter) HintOutput(m interfaces.MendikotGame) string {
	hint := m.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("mendikot.hintNone") + "\n"
	}
	card := m.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("mendikot.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, mendikotHintReasonKeys))) + "\n"
}

// mendikotHintReasonKeys maps hint-reason identifiers to their i18n keys.
var mendikotHintReasonKeys = map[string]string{
	"mendikotChaseTen":    "mendikot.hintReasonChaseTen",
	"mendikotFeedPartner": "mendikot.hintReasonFeedPartner",
	"mendikotDuck":        "mendikot.hintReasonDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MendikotCuiPresenter) ActionLogOutput(m interfaces.MendikotGame) string {
	return actionLogOutputText(m)
}
