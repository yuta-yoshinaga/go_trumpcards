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

// LooCuiPresenter renders the Loo CUI view.
type LooCuiPresenter struct{}

// looTrumpLabel は切り札スートの表示名を返す (0=未確定)。
func looTrumpLabel(suit int) string {
	if suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return "-"
	}
	return suitNames[suit]
}

// looSignedInt は符号付きでチップ増減を表示する (正なら "+" を前置)。
func looSignedInt(n int) string {
	if n > 0 {
		return "+" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// looPlayingLabel は参加状態の表示を返す。
func looPlayingLabel(playing bool) string {
	if playing {
		return i18n.T("loo.playing")
	}
	return i18n.T("loo.passed")
}

func looPlayerStr(g interfaces.LooGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("loo.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"status", looPlayingLabel(player.GetPlaying()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"chips", strconv.Itoa(player.GetChips())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *LooCuiPresenter) Output(g interfaces.LooGame, lastErr error) string {
	return buildCuiOutput(i18n.T("loo.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("loo.dealLine",
			"deal", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"trump", looTrumpLabel(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(looPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.LooTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.LooTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.LooPhaseDecide:
			deciderIdx := g.GetDecidePlayerIdx()
			b.WriteString(i18n.Tf("loo.promptDecide",
				"name", cuiPlayerName(g.GetPlayer(deciderIdx), deciderIdx)) + "\n")
		case domain.LooPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("loo.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.LooPhaseTrickEnd:
			b.WriteString(i18n.T("loo.promptTrickEnd") + "\n")
		case domain.LooPhaseRoundEnd:
			b.WriteString(i18n.T("loo.promptRoundEnd") + "\n")
			if det := g.GetLastDealDetail(); det != nil {
				// Name who was looed (penalised) and show each player's chip change
				// this deal, matching the web round-result breakdown.
				if len(det.Looed) > 0 {
					names := make([]string, len(det.Looed))
					for i, idx := range det.Looed {
						names[i] = cuiPlayerName(g.GetPlayer(idx), idx)
					}
					b.WriteString(color.Red(i18n.Tf("loo.looedList",
						"names", strings.Join(names, ", "))) + "\n")
				}
				for i := 0; i < g.GetPlayerCnt(); i++ {
					delta := det.Gained[i] // 0 when the player didn't participate
					b.WriteString(i18n.Tf("loo.chipDeltaLine",
						"name", cuiPlayerName(g.GetPlayer(i), i),
						"delta", looSignedInt(delta)) + "\n")
				}
			}
		}
		b.WriteString(i18n.T("loo.promptHelp") + "\n")
	})
}

// HintOutput emits the current Loo hint.
func (p *LooCuiPresenter) HintOutput(g interfaces.LooGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("loo.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, looHintReasonKeys)
	if hint.Decision != nil {
		decision := i18n.T("loo.passed")
		if *hint.Decision {
			decision = i18n.T("loo.playing")
		}
		return color.Yellow(i18n.Tf("loo.hintDecide",
			"decision", decision,
			"reason", reason)) + "\n"
	}
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentTurn()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("loo.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("loo.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// looHintReasonKeys maps Loo-specific hint-reason identifiers to i18n keys.
var looHintReasonKeys = map[string]string{
	"decide_play": "loo.hintReasonDecidePlay",
	"decide_pass": "loo.hintReasonDecidePass",
	"lead_high":   "loo.hintReasonLeadHigh",
	"follow_win":  "loo.hintReasonFollowWin",
	"discard_low": "loo.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LooCuiPresenter) ActionLogOutput(g interfaces.LooGame) string {
	return actionLogOutputText(g)
}
