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

// cucumberPlayerStr returns the display string for a single seat.
func cucumberPlayerStr(s interfaces.CucumberGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	role := ""
	if idx == s.GetLastTrickWinnerIdx() && s.GetLastPenalty() > 0 {
		role = i18n.Tf("cucumber.roleLastTrick", "n", strconv.Itoa(s.GetLastPenalty()))
	}
	b.WriteString(marker + i18n.Tf("cucumber.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"penalty", strconv.Itoa(player.GetPenalty()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CucumberCuiPresenter renders the Cucumber CUI view.
type CucumberCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CucumberCuiPresenter) Output(s interfaces.CucumberGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cucumber.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("cucumber.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"total", strconv.Itoa(domain.CucumberHandSize),
			"target", strconv.Itoa(s.GetConfig().TargetScore)) + "\n")
		// **スート無関係・失点は最終トリックだけ、が規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("cucumber.rule") + "\n")

		if high := s.HighestInTrick(); high > 0 {
			sb.WriteString(i18n.Tf("cucumber.highest", "n", strconv.Itoa(high)) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(cucumberPlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && s.GetPhase() == domain.CucumberPhasePlay && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			winner := s.GetWinnerIdx()
			var banner string
			if winner == 0 {
				banner = i18n.Tf("cucumber.gameEndYou", "n", strconv.Itoa(s.GetPlayer(winner).GetPenalty()))
			} else {
				banner = i18n.Tf("cucumber.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(winner), winner),
					"n", strconv.Itoa(s.GetPlayer(winner).GetPenalty()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if s.GetPhase() == domain.CucumberPhaseRoundEnd {
			loser := s.GetLastTrickWinnerIdx()
			sb.WriteString(color.Yellow(i18n.Tf("cucumber.promptRoundEnd",
				"name", cuiPlayerName(s.GetPlayer(loser), loser),
				"n", strconv.Itoa(s.GetLastPenalty()))) + "\n")
			sb.WriteString(i18n.T("cucumber.promptNext") + "\n")
			return
		}

		if !s.IsHumanTurn() {
			sb.WriteString(i18n.Tf("cucumber.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(s.GetCurrentPlayerIdx()), s.GetCurrentPlayerIdx())) + "\n")
			return
		}

		// **出す札が決まっている場面は、選べる場面と言い分けます。**
		//
		// **判定はドメインの 1 か所だけ。** ここで数え直すと、規則が変わったときに
		// 片方だけ直り損ねます（「合法手が 1 つ = 更新できない」は偽です）。
		if s.IsForcedLowest(0) {
			sb.WriteString(color.Yellow(i18n.T("cucumber.promptForced")) + "\n")
		} else if high := s.HighestInTrick(); high > 0 {
			sb.WriteString(i18n.Tf("cucumber.promptBeat", "n", strconv.Itoa(high)) + "\n")
		} else {
			sb.WriteString(i18n.T("cucumber.promptLead") + "\n")
		}
		sb.WriteString(i18n.T("cucumber.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *CucumberCuiPresenter) HintOutput(s interfaces.CucumberGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("cucumber.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, cucumberHintReasonKeys)
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("cucumber.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// cucumberHintReasonKeys maps hint-reason identifiers to their i18n keys.
var cucumberHintReasonKeys = map[string]string{
	"cucumberLead":   "cucumber.hintReasonLead",
	"cucumberBeat":   "cucumber.hintReasonBeat",
	"cucumberForced": "cucumber.hintReasonForced",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CucumberCuiPresenter) ActionLogOutput(s interfaces.CucumberGame) string {
	return actionLogOutputTextForSeats[*domain.CucumberPlayer](s)
}
