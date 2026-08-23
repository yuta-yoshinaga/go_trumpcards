//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// schafkopfSuitLabel returns the display label for a called suit integer.
func schafkopfSuitLabel(suit int) string {
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

// schafkopfPartnerLabel returns the partner display name.
func schafkopfPartnerLabel(g interfaces.SchafkopfGame) string {
	partnerIdx := g.GetPartnerIdx()
	if !g.IsPartnerRevealed() && partnerIdx >= 0 {
		return "???"
	}
	if partnerIdx < 0 {
		return "-"
	}
	return cuiPlayerName(g.GetPlayer(partnerIdx), partnerIdx)
}

// schafkopfPlayerStr returns the display string for a single Schafkopf player.
func schafkopfPlayerStr(g interfaces.SchafkopfGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("schafkopf.playerLine",
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

// schafkopfContractLabel は採用された契約の表示名を返す。Solo だけは
// 切り札スートまで見せないと、どの色が切り札か分からない。
func schafkopfContractLabel(g interfaces.SchafkopfGame) string {
	switch g.GetContract() {
	case domain.SchafkopfContractWenz:
		return i18n.T("schafkopf.contractWenz")
	case domain.SchafkopfContractSolo:
		return i18n.Tf("schafkopf.contractSolo", "suit", schafkopfSuitLabel(g.GetSoloSuit()))
	default:
		return i18n.T("schafkopf.contractRufspiel")
	}
}

// SchafkopfCuiPresenter renders the Schafkopf CUI view.
type SchafkopfCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SchafkopfCuiPresenter) Output(g interfaces.SchafkopfGame, lastErr error) string {
	return buildCuiOutput(i18n.T("schafkopf.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("schafkopf.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		if g.GetPickerIdx() >= 0 {
			b.WriteString(i18n.Tf("schafkopf.pickerLine",
				"name", cuiPlayerName(g.GetPlayer(g.GetPickerIdx()), g.GetPickerIdx()),
				"partner", schafkopfPartnerLabel(g),
				"suit", schafkopfSuitLabel(g.GetCalledSuit()),
			) + "\n")
			// **契約は切り札の構成そのもの。**出さないと、Wenz の盤面で
			// Ober が切り札でない理由が読み手に分からない。
			b.WriteString(i18n.Tf("schafkopf.contractLine",
				"contract", schafkopfContractLabel(g)) + "\n")
		} else {
			// **ブラインドは無い。** 32 枚を 4 人で配り切るので伏せ札が残らない。
			b.WriteString(i18n.Tf("schafkopf.passCount",
				"count", strconv.Itoa(g.GetPassCount())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(schafkopfPlayerStr(g, i))
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
			banner := i18n.Tf("schafkopf.gameEnd", "name", winnerName)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SchafkopfPhasePick:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("schafkopf.promptPick",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("schafkopf.promptPickHelp") + "\n")
		case domain.SchafkopfPhaseCall:
			pickerIdx := g.GetPickerIdx()
			b.WriteString(i18n.Tf("schafkopf.promptCall",
				"name", cuiPlayerName(g.GetPlayer(pickerIdx), pickerIdx)) + "\n")
			// **どのスートを呼べるかを出す。**Web は呼べるスートだけボタンを
			// 描くのに、CUI はコマンド構文しか示さず試行錯誤させていた (#4916)。
			if suits := g.GetCallableSuits(); len(suits) > 0 {
				labels := make([]string, len(suits))
				for i, s := range suits {
					// 番号を添える。c コマンドが取るのは記号ではなく数字。
					labels[i] = strconv.Itoa(s) + "=" + schafkopfSuitLabel(s)
				}
				b.WriteString(i18n.Tf("schafkopf.callableSuits",
					"suits", strings.Join(labels, " ")) + "\n")
			} else {
				// 呼べるスートが 1 つも無い局面もある (フェイル A を全部持っている等)。
				b.WriteString(i18n.T("schafkopf.callableNone") + "\n")
			}
			b.WriteString(i18n.T("schafkopf.promptCallHelp") + "\n")
		case domain.SchafkopfPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("schafkopf.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("schafkopf.promptPlayHelp") + "\n")
		case domain.SchafkopfPhaseTrickEnd:
			b.WriteString(i18n.T("schafkopf.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("schafkopf.promptTrickEndHelp") + "\n")
		case domain.SchafkopfPhaseRoundEnd:
			b.WriteString(i18n.Tf("schafkopf.promptRoundEnd",
				"pts", strconv.Itoa(g.GetRoundPickerPoints()),
				"mult", strconv.Itoa(g.GetRoundMultiplier())) + "\n")
			// Points and multiplier alone don't say who won; spell out the picker
			// team's result and the chip direction.
			if g.GetRoundPickerWon() {
				b.WriteString(color.Green(i18n.T("schafkopf.roundPickerWon")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("schafkopf.roundPickerLost")) + "\n")
			}
			b.WriteString(i18n.T("schafkopf.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Schafkopf hint.
func (p *SchafkopfCuiPresenter) HintOutput(g interfaces.SchafkopfGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("schafkopf.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, schafkopfHintReasonKeys)
	switch {
	case len(hint.CardIndices) > 0:
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
		return color.Yellow(i18n.Tf("schafkopf.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	case hint.Suit != 0:
		return color.Yellow(i18n.Tf("schafkopf.hintSuit",
			"suit", schafkopfSuitLabel(hint.Suit),
			"reason", reason)) + "\n"
	default:
		action := "pick"
		if !hint.Pick {
			action = "pass"
		}
		return color.Yellow(i18n.Tf("schafkopf.hintPick",
			"action", action,
			"reason", reason)) + "\n"
	}
}

// schafkopfHintReasonKeys maps Schafkopf-specific hint-reason identifiers to i18n keys.
var schafkopfHintReasonKeys = map[string]string{
	"pick_take":   "schafkopf.hintReasonPickTake",
	"pick_pass":   "schafkopf.hintReasonPickPass",
	"call_suit":   "schafkopf.hintReasonCallSuit",
	"lead_low":    "schafkopf.hintReasonLeadLow",
	"follow_win":  "schafkopf.hintReasonFollowWin",
	"follow_duck": "schafkopf.hintReasonFollowDuck",
	"discard_low": "schafkopf.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SchafkopfCuiPresenter) ActionLogOutput(g interfaces.SchafkopfGame) string {
	return actionLogOutputTextForSeats[*domain.SchafkopfPlayer](g)
}
