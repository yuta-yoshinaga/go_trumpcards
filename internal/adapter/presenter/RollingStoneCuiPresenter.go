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

// rollingStonePlayerStr returns the display string for a single player.
func rollingStonePlayerStr(s interfaces.RollingStoneGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	role := ""
	switch {
	case player.HasFinished():
		role = i18n.Tf("rollingstone.roleFinished", "rank", strconv.Itoa(player.GetFinishedAt()))
	case idx == s.GetLastPickupIdx():
		role = i18n.T("rollingstone.rolePickedUp")
	}
	b.WriteString(marker + i18n.Tf("rollingstone.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"pickups", strconv.Itoa(player.GetPickups()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// RollingStoneCuiPresenter renders the Rolling Stone CUI view.
type RollingStoneCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *RollingStoneCuiPresenter) Output(s interfaces.RollingStoneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("rollingstone.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("rollingstone.header",
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"deck", strconv.Itoa(s.GetDeckSize()),
			"left", strconv.Itoa(s.GetDeckSize()-s.GetDiscarded())) + "\n")
		// **勝利条件が逆さまなのが規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("rollingstone.rule") + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(rollingStonePlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			switch {
			case s.GetWinnerIdx() >= 0 && s.GetPlayer(s.GetWinnerIdx()).GetCardsSize() > 0:
				// **上限で切った局は「上がった」わけではない。**
				banner = i18n.Tf("rollingstone.gameEndStalemate",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()),
					"n", strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetCardsSize()))
			case s.GetWinnerIdx() == 0:
				banner = i18n.T("rollingstone.gameEndYou")
			default:
				banner = i18n.Tf("rollingstone.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := s.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("rollingstone.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		// **出せる札が無い局面はそう名乗る。** 黙っていると打てない理由が分からない。
		if s.MustPickUp(0) && currentIdx == 0 {
			sb.WriteString(color.Yellow(i18n.Tf("rollingstone.promptPickUp",
				"suit", cuiSuitName(s.GetLeadSuit()),
				"n", strconv.Itoa(len(s.GetCurrentTrick())))) + "\n")
			return
		}
		sb.WriteString(i18n.T("rollingstone.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *RollingStoneCuiPresenter) HintOutput(s interfaces.RollingStoneGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("rollingstone.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, rollingStoneHintReasonKeys)
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("rollingstone.hintPickUp", "reason", reason)) + "\n"
	}
	card := s.GetPlayer(s.GetCurrentPlayerIdx()).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("rollingstone.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// rollingStoneHintReasonKeys maps hint-reason identifiers to their i18n keys.
var rollingStoneHintReasonKeys = map[string]string{
	"rollingstoneLead":   "rollingstone.hintReasonLead",
	"rollingstoneFollow": "rollingstone.hintReasonFollow",
	"rollingstonePickUp": "rollingstone.hintReasonPickUp",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RollingStoneCuiPresenter) ActionLogOutput(s interfaces.RollingStoneGame) string {
	return actionLogOutputTextForSeats[*domain.RollingStonePlayer](s)
}
