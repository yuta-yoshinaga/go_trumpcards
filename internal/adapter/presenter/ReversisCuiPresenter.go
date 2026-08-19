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

// reversisPlayerStr returns the display string for a single player.
func reversisPlayerStr(player *domain.ReversisPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("reversis.playerLine",
		"name", cuiPlayerName(player, idx),
		"chips", strconv.Itoa(player.GetChips()),
		"penalty", strconv.Itoa(player.GetRoundPenalty()),
		"marks", reversisMarkStr(player),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(reversisHandStr(player) + "\n")
	}
	return b.String()
}

// reversisHandStr は手札を、札ごとの失点付きで並べる。
//
// **点を取り合うのが核なのに、どの札が何点かは画面に出ていなかった** (#5747)。
// A=4 / K=3 / Q=2 / J=1 を暗算し続けることになる。値は
// domain.ReversisCardPenalty から引き、表を写さない。
func reversisHandStr(player *domain.ReversisPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := range player.GetCardsSize() {
		card := player.GetCard(i)
		parts = append(parts, i18n.Tf("reversis.handCard",
			"idx", strconv.Itoa(i),
			"card", cuiCardStr(card),
			"points", strconv.Itoa(domain.ReversisCardPenalty(card))))
	}
	return strings.Join(parts, "  ")
}

// reversisMarkStr 取ってしまった印付きの札を短く表す
func reversisMarkStr(player *domain.ReversisPlayer) string {
	marks := make([]string, 0, 2)
	if player.GetTookQuinola() {
		marks = append(marks, i18n.T("reversis.markQuinola"))
	}
	if player.GetTookDiamondAce() {
		marks = append(marks, i18n.T("reversis.markDiamondAce"))
	}
	if len(marks) == 0 {
		return i18n.T("reversis.markNone")
	}
	return strings.Join(marks, ",")
}

// ReversisCuiPresenter renders the Reversis CUI view.
type ReversisCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ReversisCuiPresenter) Output(r interfaces.ReversisGame, lastErr error) string {
	return buildCuiOutput(i18n.T("reversis.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("reversis.header",
			"round", strconv.Itoa(r.GetRoundNumber()),
			"rounds", strconv.Itoa(r.GetConfig().Rounds),
			"trick", strconv.Itoa(r.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.ReversisTricksPerRound)) + "\n")
		// **プールは常に見せる。** 何を取り合っているのかが盤面から読めない。
		sb.WriteString(color.Yellow(i18n.Tf("reversis.poolLine", "pool", strconv.Itoa(r.GetPool()))) + "\n")
		sb.WriteString(i18n.T("reversis.penaltyLine") + "\n")

		for i := 0; i < r.GetPlayerCnt(); i++ {
			sb.WriteString(reversisPlayerStr(r.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, r.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(r.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if r.GetGameEndFlag() {
			var banner string
			if r.GetWinnerIdx() < 0 {
				banner = i18n.T("reversis.gameEndTie")
			} else {
				banner = i18n.Tf("reversis.gameEndWinner",
					"name", cuiPlayerName(r.GetPlayer(r.GetWinnerIdx()), r.GetWinnerIdx()),
					"chips", strconv.Itoa(r.GetPlayer(r.GetWinnerIdx()).GetChips()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if r.GetPhase() == domain.ReversisPhaseRoundEnd {
			sb.WriteString(i18n.T("reversis.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("reversis.promptNext") + "\n")
			return
		}

		currentIdx := r.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("reversis.promptCurrentPlayer",
			"name", cuiPlayerName(r.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("reversis.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *ReversisCuiPresenter) HintOutput(r interfaces.ReversisGame) string {
	hint := r.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("reversis.hintNone") + "\n"
	}
	card := r.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("reversis.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, reversisHintReasonKeys))) + "\n"
}

// reversisHintReasonKeys maps hint-reason identifiers to their i18n keys.
var reversisHintReasonKeys = map[string]string{
	"reversisAvoidMarked": "reversis.hintReasonAvoidMarked",
	"reversisAvoidPoints": "reversis.hintReasonAvoidPoints",
	"reversisLeadSafe":    "reversis.hintReasonLeadSafe",
	"reversisDumpHigh":    "reversis.hintReasonDumpHigh",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ReversisCuiPresenter) ActionLogOutput(r interfaces.ReversisGame) string {
	return actionLogOutputText(r)
}
