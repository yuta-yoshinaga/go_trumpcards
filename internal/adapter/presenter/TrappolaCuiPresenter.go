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

// trappolaPlayerStr returns the display string for a single Trappola player.
func trappolaPlayerStr(player *domain.TrappolaPlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("trappola.playerLine",
		"name", cuiPlayerName(player, i),
		"team", trappolaTeamLabel(domain.TrappolaTeamOf(i)),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **マストフォローで何が出せるかを示す。**Web はリング表示しているのに、
		// CUI は番号を打ってエラーを踏むまで分からなかった (#5633)。playable が
		// 空なら無印 (制限が決まっていない状態と区別する)。
		b.WriteString(cuiPlayableMarkedCardListStr(player, playable) + "\n")
	}
	return b.String()
}

// trappolaTeamLabel チーム表示名 (0=A, 1=B)
func trappolaTeamLabel(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// TrappolaCuiPresenter renders the Trappola CUI view.
type TrappolaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TrappolaCuiPresenter) Output(g interfaces.TrappolaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("trappola.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("trappola.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		thirds := g.GetTeamRoundThirds()
		b.WriteString(i18n.Tf("trappola.teamScores",
			"a", strconv.Itoa(scores[0]),
			"athird", strconv.Itoa(thirds[0]),
			"b", strconv.Itoa(scores[1]),
			"bthird", strconv.Itoa(thirds[1])) + "\n")
		b.WriteString(i18n.T("trappola.thirdsRule") + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			if player := g.GetPlayer(i); player != nil {
				// 目印はプレイフェーズで本人の手番のときだけ。
				var playable []int
				if g.GetPhase() == domain.TrappolaPhasePlay && g.GetCurrentPlayerIdx() == i {
					playable = g.GetPlayableIndices(i)
				}
				b.WriteString(trappolaPlayerStr(player, i, playable))
			}
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("trappola.gameEnd", "team", trappolaTeamLabel(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TrappolaPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("trappola.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("trappola.promptPlayHelp") + "\n")
		case domain.TrappolaPhaseTrickEnd:
			b.WriteString(i18n.T("trappola.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("trappola.promptTrickEndHelp") + "\n")
		case domain.TrappolaPhaseRoundEnd:
			// Break down each team's thirds this round and name the team that took
			// the last trick (the ultima-presa +1/3 bonus goes to that trick's
			// winner, who becomes the next lead).
			lastTeam := domain.TrappolaTeamOf(g.GetLeadPlayerIdx())
			b.WriteString(i18n.Tf("trappola.roundBreakdown",
				"a", trappolaTeamLabel(0),
				"athird", strconv.Itoa(thirds[0]),
				"b", trappolaTeamLabel(1),
				"bthird", strconv.Itoa(thirds[1]),
				"lastteam", trappolaTeamLabel(lastTeam)) + "\n")
			b.WriteString(i18n.T("trappola.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("trappola.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Trappola hint.
func (p *TrappolaCuiPresenter) HintOutput(g interfaces.TrappolaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("trappola.hintNone") + "\n"
	}
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
	}
	return color.Yellow(i18n.Tf("trappola.hintCard",
		"cards", strings.Join(cards, ", "),
		"reason", hintReasonStr(hint.Reason, trappolaHintReasonKeys))) + "\n"
}

// trappolaHintReasonKeys maps Trappola-specific hint-reason identifiers to i18n keys.
var trappolaHintReasonKeys = map[string]string{
	"lead_low":     "trappola.hintReasonLeadLow",
	"follow_win":   "trappola.hintReasonFollowWin",
	"follow_duck":  "trappola.hintReasonFollowDuck",
	"give_partner": "trappola.hintReasonGivePartner",
	"discard_low":  "trappola.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrappolaCuiPresenter) ActionLogOutput(g interfaces.TrappolaGame) string {
	return actionLogOutputTextForSeats[*domain.TrappolaPlayer](g)
}
