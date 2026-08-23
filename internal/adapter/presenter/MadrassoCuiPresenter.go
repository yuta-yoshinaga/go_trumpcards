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

// madrassoPlayerStr returns the display string for a single Madrasso player.
func madrassoPlayerStr(player *domain.MadrassoPlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("madrasso.playerLine",
		"name", cuiPlayerName(player, i),
		"team", madrassoTeamLabel(domain.MadrassoTeamOf(i)),
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

// madrassoTeamLabel チーム表示名 (0=A, 1=B)
func madrassoTeamLabel(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// MadrassoCuiPresenter renders the Madrasso CUI view.
type MadrassoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MadrassoCuiPresenter) Output(g interfaces.MadrassoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("madrasso.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("madrasso.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		thirds := g.GetTeamRoundPoints()
		b.WriteString(i18n.Tf("madrasso.teamScores",
			"a", strconv.Itoa(scores[0]),
			"athird", strconv.Itoa(thirds[0]),
			"b", strconv.Itoa(scores[1]),
			"bthird", strconv.Itoa(thirds[1])) + "\n")
		b.WriteString(i18n.T("madrasso.thirdsRule") + "\n")
		// **切り札は配りで決まる。** クローン元 (トレセッテ) には無い概念なので、
		// 出さないと盤面から切り札が読めない。
		if s := g.GetTrumpSuit(); s >= domain.CardDesignSpade && s <= domain.CardDesignMax {
			b.WriteString(i18n.Tf("madrasso.trumpLine", "suit", cuiSuitName(s)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			if player := g.GetPlayer(i); player != nil {
				// 目印はプレイフェーズで本人の手番のときだけ。
				var playable []int
				if g.GetPhase() == domain.MadrassoPhasePlay && g.GetCurrentPlayerIdx() == i {
					playable = g.GetPlayableIndices(i)
				}
				b.WriteString(madrassoPlayerStr(player, i, playable))
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
			banner := i18n.Tf("madrasso.gameEnd", "team", madrassoTeamLabel(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MadrassoPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("madrasso.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("madrasso.promptPlayHelp") + "\n")
		case domain.MadrassoPhaseTrickEnd:
			b.WriteString(i18n.T("madrasso.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("madrasso.promptTrickEndHelp") + "\n")
		case domain.MadrassoPhaseRoundEnd:
			// Break down each team's thirds this round and name the team that took
			// the last trick (the ultima-presa +1/3 bonus goes to that trick's
			// winner, who becomes the next lead).
			lastTeam := domain.MadrassoTeamOf(g.GetLeadPlayerIdx())
			b.WriteString(i18n.Tf("madrasso.roundBreakdown",
				"a", madrassoTeamLabel(0),
				"athird", strconv.Itoa(thirds[0]),
				"b", madrassoTeamLabel(1),
				"bthird", strconv.Itoa(thirds[1]),
				"lastteam", madrassoTeamLabel(lastTeam)) + "\n")
			b.WriteString(i18n.T("madrasso.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("madrasso.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Madrasso hint.
func (p *MadrassoCuiPresenter) HintOutput(g interfaces.MadrassoGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("madrasso.hintNone") + "\n"
	}
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
	}
	return color.Yellow(i18n.Tf("madrasso.hintCard",
		"cards", strings.Join(cards, ", "),
		"reason", hintReasonStr(hint.Reason, madrassoHintReasonKeys))) + "\n"
}

// madrassoHintReasonKeys maps Madrasso-specific hint-reason identifiers to i18n keys.
var madrassoHintReasonKeys = map[string]string{
	"lead_low":     "madrasso.hintReasonLeadLow",
	"follow_win":   "madrasso.hintReasonFollowWin",
	"follow_duck":  "madrasso.hintReasonFollowDuck",
	"give_partner": "madrasso.hintReasonGivePartner",
	"discard_low":  "madrasso.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MadrassoCuiPresenter) ActionLogOutput(g interfaces.MadrassoGame) string {
	return actionLogOutputTextForSeats[*domain.MadrassoPlayer](g)
}
