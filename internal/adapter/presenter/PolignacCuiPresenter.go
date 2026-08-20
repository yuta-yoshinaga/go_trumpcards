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

// polignacPlayerStr returns the display string for a single player.
func polignacPlayerStr(player *domain.PolignacPlayer, idx int, isCapot bool) string {
	var b strings.Builder
	name := cuiPlayerName(player, idx)
	if isCapot {
		name += i18n.T("polignac.capotMark")
	}
	b.WriteString(i18n.Tf("polignac.playerLine",
		"name", name,
		"score", strconv.Itoa(player.GetScore()),
		"round", strconv.Itoa(player.GetRoundPenalty()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	// **合計失点だけでは、♠J を踏んだのか他を 2 枚拾ったのかが分からない** (#5746)。
	if suits := player.GetTakenJackSuits(); len(suits) > 0 {
		marks := make([]string, 0, len(suits))
		for _, suit := range suits {
			if suit == domain.CardDesignSpade {
				marks = append(marks, i18n.T("polignac.jackSpade"))
				continue
			}
			marks = append(marks, i18n.Tf("polignac.jackOther",
				"suit", cuiCardStr(domain.NewCard(suit, domain.PolignacJackValue, true))))
		}
		b.WriteString(i18n.Tf("polignac.jackMarks", "jacks", strings.Join(marks, " ")))
	}
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// PolignacCuiPresenter renders the Polignac CUI view.
type PolignacCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PolignacCuiPresenter) Output(g interfaces.PolignacGame, lastErr error) string {
	return buildCuiOutput(i18n.T("polignac.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("polignac.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"rounds", strconv.Itoa(g.GetConfig().Rounds),
			"trick", strconv.Itoa(g.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.PolignacTricksPerRound)) + "\n")
		// **失点するのはジャック4枚だけ、♠J はその2倍。** 常時出す。
		sb.WriteString(i18n.T("polignac.penaltyLine") + "\n")

		if idx := g.GetCapotIdx(); idx >= 0 {
			sb.WriteString(color.Yellow(i18n.Tf("polignac.capotActive",
				"name", cuiPlayerName(g.GetPlayer(idx), idx),
				"tricks", strconv.Itoa(g.GetCapotTricks()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			sb.WriteString(polignacPlayerStr(g.GetPlayer(i), i, i == g.GetCapotIdx()))
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
			if g.GetWinnerIdx() < 0 {
				banner = i18n.T("polignac.gameEndTie")
			} else {
				banner = i18n.Tf("polignac.gameEndWinner",
					"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()),
					"score", strconv.Itoa(g.GetPlayer(g.GetWinnerIdx()).GetScore()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.PolignacPhaseDeclare:
			sb.WriteString(i18n.T("polignac.promptDeclare") + "\n")
			sb.WriteString(i18n.T("polignac.promptDeclareHelp") + "\n")
			return
		case domain.PolignacPhaseRoundEnd:
			sb.WriteString(i18n.T("polignac.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("polignac.promptNext") + "\n")
			return
		}

		currentIdx := g.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("polignac.promptCurrentPlayer",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("polignac.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *PolignacCuiPresenter) HintOutput(g interfaces.PolignacGame) string {
	hint := g.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("polignac.hintNone") + "\n"
	}
	card := g.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("polignac.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, polignacHintReasonKeys))) + "\n"
}

// polignacHintReasonKeys maps hint-reason identifiers to their i18n keys.
var polignacHintReasonKeys = map[string]string{
	"polignacAvoidJack":  "polignac.hintReasonAvoidJack",
	"polignacDumpJack":   "polignac.hintReasonDumpJack",
	"polignacLeadSafe":   "polignac.hintReasonLeadSafe",
	"polignacBlockCapot": "polignac.hintReasonBlockCapot",
	"polignacWinCapot":   "polignac.hintReasonWinCapot",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PolignacCuiPresenter) ActionLogOutput(g interfaces.PolignacGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
