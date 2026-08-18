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

// sheepsheadSuitLabel returns the display label for a called suit integer.
func sheepsheadSuitLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	default:
		return "-"
	}
}

// sheepsheadPartnerLabel returns the partner display name.
func sheepsheadPartnerLabel(g interfaces.SheepsheadGame) string {
	partnerIdx := g.GetPartnerIdx()
	if !g.IsPartnerRevealed() && partnerIdx >= 0 {
		return "???"
	}
	if partnerIdx < 0 {
		return "-"
	}
	return cuiPlayerName(g.GetPlayer(partnerIdx), partnerIdx)
}

// sheepsheadPlayerStr returns the display string for a single Sheepshead player.
func sheepsheadPlayerStr(g interfaces.SheepsheadGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("sheepshead.playerLine",
		"name", cuiPlayerName(player, i),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"chips", strconv.Itoa(player.GetChips()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// SheepsheadCuiPresenter renders the Sheepshead CUI view.
type SheepsheadCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SheepsheadCuiPresenter) Output(g interfaces.SheepsheadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sheepshead.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("sheepshead.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		if g.GetPickerIdx() >= 0 {
			b.WriteString(i18n.Tf("sheepshead.pickerLine",
				"name", cuiPlayerName(g.GetPlayer(g.GetPickerIdx()), g.GetPickerIdx()),
				"partner", sheepsheadPartnerLabel(g),
				"suit", sheepsheadSuitLabel(g.GetCalledSuit()),
			) + "\n")
		} else {
			b.WriteString(i18n.Tf("sheepshead.blindCount",
				"count", strconv.Itoa(len(g.GetBlind()))) + "  " +
				i18n.Tf("sheepshead.passCount",
					"count", strconv.Itoa(g.GetPassCount())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sheepsheadPlayerStr(g, i))
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
			banner := i18n.Tf("sheepshead.gameEnd", "name", winnerName)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SheepsheadPhasePick:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sheepshead.promptPick",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sheepshead.promptPickHelp") + "\n")
		case domain.SheepsheadPhaseBury:
			pickerIdx := g.GetPickerIdx()
			b.WriteString(i18n.Tf("sheepshead.promptBury",
				"name", cuiPlayerName(g.GetPlayer(pickerIdx), pickerIdx)) + "\n")
			b.WriteString(i18n.T("sheepshead.promptBuryHelp") + "\n")
		case domain.SheepsheadPhaseCall:
			pickerIdx := g.GetPickerIdx()
			b.WriteString(i18n.Tf("sheepshead.promptCall",
				"name", cuiPlayerName(g.GetPlayer(pickerIdx), pickerIdx)) + "\n")
			// **どのスートを呼べるかを出す。**Web は呼べるスートだけボタンを
			// 描くのに、CUI はコマンド構文しか示さず試行錯誤させていた (#4916)。
			if suits := g.GetCallableSuits(); len(suits) > 0 {
				labels := make([]string, len(suits))
				for i, s := range suits {
					// 番号を添える。c コマンドが取るのは記号ではなく数字。
					labels[i] = strconv.Itoa(s) + "=" + sheepsheadSuitLabel(s)
				}
				b.WriteString(i18n.Tf("sheepshead.callableSuits",
					"suits", strings.Join(labels, " ")) + "\n")
			} else {
				// 呼べるスートが 1 つも無い局面もある (フェイル A を全部持っている等)。
				b.WriteString(i18n.T("sheepshead.callableNone") + "\n")
			}
			b.WriteString(i18n.T("sheepshead.promptCallHelp") + "\n")
		case domain.SheepsheadPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sheepshead.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sheepshead.promptPlayHelp") + "\n")
		case domain.SheepsheadPhaseTrickEnd:
			b.WriteString(i18n.T("sheepshead.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("sheepshead.promptTrickEndHelp") + "\n")
		case domain.SheepsheadPhaseRoundEnd:
			b.WriteString(i18n.Tf("sheepshead.promptRoundEnd",
				"pts", strconv.Itoa(g.GetRoundPickerPoints()),
				"mult", strconv.Itoa(g.GetRoundMultiplier())) + "\n")
			// Points and multiplier alone don't say who won; spell out the picker
			// team's result and the chip direction.
			if g.GetRoundPickerWon() {
				b.WriteString(color.Green(i18n.T("sheepshead.roundPickerWon")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("sheepshead.roundPickerLost")) + "\n")
			}
			// **ピッカーが何を埋めたかは得点の内訳そのもの** (#5638)。domain は
			// 保持していて Web API も送っているのに、CUI は最後まで出さなかった。
			// 公開されるのはラウンド終了以降なので、ここでだけ出す。
			if buried := g.GetBuried(); len(buried) > 0 {
				cards := make([]string, len(buried))
				for i, c := range buried {
					cards[i] = cuiCardStr(c)
				}
				b.WriteString(i18n.T("sheepshead.roundEndBuried") + ": " +
					strings.Join(cards, "  ") + "\n")
			}
			b.WriteString(i18n.T("sheepshead.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Sheepshead hint.
func (p *SheepsheadCuiPresenter) HintOutput(g interfaces.SheepsheadGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("sheepshead.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, sheepsheadHintReasonKeys)
	switch {
	case len(hint.CardIndices) > 0:
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.SheepsheadPhaseBury {
			playerIdx = g.GetPickerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("sheepshead.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	case hint.Suit != 0:
		return color.Yellow(i18n.Tf("sheepshead.hintSuit",
			"suit", sheepsheadSuitLabel(hint.Suit),
			"reason", reason)) + "\n"
	default:
		action := "pick"
		if !hint.Pick {
			action = "pass"
		}
		return color.Yellow(i18n.Tf("sheepshead.hintPick",
			"action", action,
			"reason", reason)) + "\n"
	}
}

// sheepsheadHintReasonKeys maps Sheepshead-specific hint-reason identifiers to i18n keys.
var sheepsheadHintReasonKeys = map[string]string{
	"pick_take":   "sheepshead.hintReasonPickTake",
	"pick_pass":   "sheepshead.hintReasonPickPass",
	"bury_low":    "sheepshead.hintReasonBuryLow",
	"call_suit":   "sheepshead.hintReasonCallSuit",
	"lead_low":    "sheepshead.hintReasonLeadLow",
	"follow_win":  "sheepshead.hintReasonFollowWin",
	"follow_duck": "sheepshead.hintReasonFollowDuck",
	"discard_low": "sheepshead.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SheepsheadCuiPresenter) ActionLogOutput(g interfaces.SheepsheadGame) string {
	return actionLogOutputText(g)
}
