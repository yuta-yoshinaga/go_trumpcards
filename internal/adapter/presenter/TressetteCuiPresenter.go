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

// tressettePlayerStr returns the display string for a single Tressette player.
func tressettePlayerStr(player *domain.TressettePlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tressette.playerLine",
		"name", cuiPlayerName(player, i),
		"team", tressetteTeamLabel(domain.TressetteTeamOf(i)),
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

// tressetteTeamLabel チーム表示名 (0=A, 1=B)
func tressetteTeamLabel(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// TressetteCuiPresenter renders the Tressette CUI view.
type TressetteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TressetteCuiPresenter) Output(g interfaces.TressetteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tressette.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tressette.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		thirds := g.GetTeamRoundThirds()
		b.WriteString(i18n.Tf("tressette.teamScores",
			"a", strconv.Itoa(scores[0]),
			"athird", strconv.Itoa(thirds[0]),
			"b", strconv.Itoa(scores[1]),
			"bthird", strconv.Itoa(thirds[1])) + "\n")
		b.WriteString(i18n.T("tressette.thirdsRule") + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			if player := g.GetPlayer(i); player != nil {
				// 目印はプレイフェーズで本人の手番のときだけ。
				var playable []int
				if g.GetPhase() == domain.TressettePhasePlay && g.GetCurrentPlayerIdx() == i {
					playable = g.GetPlayableIndices(i)
				}
				b.WriteString(tressettePlayerStr(player, i, playable))
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
			banner := i18n.Tf("tressette.gameEnd", "team", tressetteTeamLabel(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TressettePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tressette.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tressette.promptPlayHelp") + "\n")
		case domain.TressettePhaseTrickEnd:
			b.WriteString(i18n.T("tressette.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("tressette.promptTrickEndHelp") + "\n")
		case domain.TressettePhaseRoundEnd:
			// Break down each team's thirds this round and name the team that took
			// the last trick (the ultima-presa +1/3 bonus goes to that trick's
			// winner, who becomes the next lead).
			lastTeam := domain.TressetteTeamOf(g.GetLeadPlayerIdx())
			b.WriteString(i18n.Tf("tressette.roundBreakdown",
				"a", tressetteTeamLabel(0),
				"athird", strconv.Itoa(thirds[0]),
				"b", tressetteTeamLabel(1),
				"bthird", strconv.Itoa(thirds[1]),
				"lastteam", tressetteTeamLabel(lastTeam)) + "\n")
			b.WriteString(i18n.T("tressette.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tressette.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Tressette hint.
func (p *TressetteCuiPresenter) HintOutput(g interfaces.TressetteGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tressette.hintNone") + "\n"
	}
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
	}
	return color.Yellow(i18n.Tf("tressette.hintCard",
		"cards", strings.Join(cards, ", "),
		"reason", hintReasonStr(hint.Reason, tressetteHintReasonKeys))) + "\n"
}

// tressetteHintReasonKeys maps Tressette-specific hint-reason identifiers to i18n keys.
var tressetteHintReasonKeys = map[string]string{
	"lead_low":     "tressette.hintReasonLeadLow",
	"follow_win":   "tressette.hintReasonFollowWin",
	"follow_duck":  "tressette.hintReasonFollowDuck",
	"give_partner": "tressette.hintReasonGivePartner",
	"discard_low":  "tressette.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TressetteCuiPresenter) ActionLogOutput(g interfaces.TressetteGame) string {
	return actionLogOutputTextForSeats[*domain.TressettePlayer](g)
}
