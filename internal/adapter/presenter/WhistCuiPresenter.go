package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// whistPlayerStr returns the display string for a single Whist player.
func whistPlayerStr(player *domain.WhistPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("whist.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// WhistCuiPresenter renders the Whist CUI view.
type WhistCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *WhistCuiPresenter) Output(w interfaces.WhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("whist.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("whist.header",
			"round", strconv.Itoa(w.GetRoundNumber()),
			"trick", strconv.Itoa(w.GetTrickNumber())) + "\n")
		b.WriteString(i18n.Tf("whist.trumpLine",
			"suit", suitDisplayName(w.GetTrumpSuit())) + "\n")
		b.WriteString(i18n.Tf("whist.teamScoreLine",
			"t0", strconv.Itoa(w.GetTeamScore(0)),
			"t1", strconv.Itoa(w.GetTeamScore(1))) + "\n")

		for i := 0; i < w.GetPlayerCnt(); i++ {
			b.WriteString(whistPlayerStr(w.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := w.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(w.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if w.GetGameEndFlag() {
			banner := i18n.Tf("whist.gameEnd", "team", strconv.Itoa(w.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch w.GetPhase() {
		case domain.WhistPhasePlay:
			currentIdx := w.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("whist.promptCurrentPlayer",
				"name", cuiPlayerName(w.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("whist.promptPlay") + "\n")
		case domain.WhistPhaseTrickEnd:
			b.WriteString(i18n.T("whist.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("whist.promptTrickEndHelp") + "\n")
		case domain.WhistPhaseRoundEnd:
			b.WriteString(i18n.T("whist.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("whist.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Whist hint.
func (p *WhistCuiPresenter) HintOutput(w interfaces.WhistGame) string {
	hint := w.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("whist.hintNone") + "\n"
	}
	player := w.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("whist.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, whistHintReasonKeys))) + "\n"
}

// whistHintReasonKeys maps Whist-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var whistHintReasonKeys = map[string]string{
	"trump_cut": "whist.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WhistCuiPresenter) ActionLogOutput(w interfaces.WhistGame) string {
	return actionLogOutputText(w)
}
