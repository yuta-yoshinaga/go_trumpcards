//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// doppelkopfPlayerStr returns the display string for a single Doppelkopf player.
func doppelkopfPlayerStr(g interfaces.DoppelkopfGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("doppelkopf.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"chips", strconv.Itoa(player.GetChips()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **切り札は「♦全札 + 全 Q + 全 J + ♥10」で、並びからは判別できない**
		// (#5639)。Web は該当カードにバッジを付けているのに、CUI は無印で
		// 並べるだけだった。判定はドメインに任せる (dkIsTrump が正)。
		trumps := g.GetTrumpIndices(idx)
		b.WriteString(cuiIndexMarkedCardListStr(player, trumps, CuiTrumpMark) + "\n")
		if len(trumps) > 0 {
			b.WriteString(i18n.T("doppelkopf.trumpLegend") + "  " +
				i18n.T("doppelkopf.trumpOrder") + "\n")
		}
	}
	return b.String()
}

// DoppelkopfCuiPresenter renders the Doppelkopf CUI view.
type DoppelkopfCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DoppelkopfCuiPresenter) Output(g interfaces.DoppelkopfGame, lastErr error) string {
	return buildCuiOutput(i18n.T("doppelkopf.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("doppelkopf.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(doppelkopfPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			var winnerName string
			if winnerIdx >= 0 {
				winnerName = cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx)
			}
			banner := i18n.Tf("doppelkopf.gameEnd", "name", winnerName)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.DoppelkopfPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("doppelkopf.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("doppelkopf.promptPlayHelp") + "\n")
			if g.CanHumanAnnounce() {
				b.WriteString(i18n.T("doppelkopf.promptAnnounce") + "\n")
			}
		case domain.DoppelkopfPhaseTrickEnd:
			b.WriteString(i18n.T("doppelkopf.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("doppelkopf.promptTrickEndHelp") + "\n")
		case domain.DoppelkopfPhaseRoundEnd:
			reWon := g.GetRoundReWon()
			outcome := i18n.T("doppelkopf.outcomeKontraWins")
			if reWon {
				outcome = i18n.T("doppelkopf.outcomeReWins")
			}
			b.WriteString(i18n.Tf("doppelkopf.promptRoundEnd",
				"rePts", strconv.Itoa(g.GetRoundRePoints()),
				"outcome", outcome,
				"gamePts", strconv.Itoa(g.GetRoundGamePoints())) + "\n")
			b.WriteString(i18n.T("doppelkopf.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Doppelkopf hint.
func (p *DoppelkopfCuiPresenter) HintOutput(g interfaces.DoppelkopfGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("doppelkopf.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, doppelkopfHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("doppelkopf.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("doppelkopf.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// doppelkopfHintReasonKeys maps Doppelkopf-specific hint-reason identifiers to i18n keys.
var doppelkopfHintReasonKeys = map[string]string{
	"lead_low":    "doppelkopf.hintReasonLeadLow",
	"follow_win":  "doppelkopf.hintReasonFollowWin",
	"follow_duck": "doppelkopf.hintReasonFollowDuck",
	"discard_low": "doppelkopf.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DoppelkopfCuiPresenter) ActionLogOutput(g interfaces.DoppelkopfGame) string {
	return actionLogOutputTextForSeats[*domain.DoppelkopfPlayer](g)
}
