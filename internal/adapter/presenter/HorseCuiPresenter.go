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

// HorseCuiPresenter renders the H.O.R.S.E. CUI view.
type HorseCuiPresenter struct{}

// Output renders the current game state for the active locale.
//
// **出すのは打つのに要るものだけ。** どの種目の何ハンド目か、席の残高、そして
// 見えている札 ── 役の判定や勝敗の内訳は種目側の実装が持っている。
func (p *HorseCuiPresenter) Output(g interfaces.HorseGame, lastErr error) string {
	return buildCuiOutput(i18n.T("horse.helpTitle"), func(b *strings.Builder) {
		cfg := g.GetConfig()
		b.WriteString(i18n.Tf("horse.round",
			"letter", g.GetDisciplineLetter(),
			"name", i18n.T("horse.discipline."+domain.HorseDisciplineName(g.GetDiscipline())),
			"hand", strconv.Itoa(g.GetHandInDiscipline()),
			"hands", strconv.Itoa(cfg.HandsPerDiscipline),
			"total", strconv.Itoa(g.GetHandNumber())) + "\n")

		if shared := g.GetCommunityCards(); len(shared) > 0 {
			b.WriteString(i18n.T("horse.community") + " " + horseCardsText(shared) + "\n")
		}
		for i := 0; i < g.GetSeatCount(); i++ {
			line := i18n.Tf("horse.seatLine",
				"name", g.GetSeatName(i),
				"chips", strconv.Itoa(g.GetSeatChips(i)))
			// **見えている札だけを並べる。** CPU の伏せ札はドメインが返さない。
			if cards := g.GetSeatCards(i); len(cards) > 0 {
				line += "  " + horseCardsText(cards)
			}
			if i == g.GetCurrentTurn() && g.GetPhase() == domain.HorsePhaseHand {
				line = color.Yellow(line + i18n.T("horse.turnMark"))
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch {
		case g.GetGameEndFlag():
			b.WriteString(color.Green(i18n.Tf("horse.gameEnd",
				"name", g.GetSeatName(g.WinnerSeat()),
				"chips", strconv.Itoa(g.GetSeatChips(g.WinnerSeat())))) + "\n")
		case g.GetPhase() == domain.HorsePhaseHandEnd:
			b.WriteString(i18n.T("horse.promptHandEnd") + "\n")
			b.WriteString(i18n.T("horse.promptHandEndHelp") + "\n")
		default:
			b.WriteString(i18n.Tf("horse.promptPlay", "pot", strconv.Itoa(g.GetPot())) + "\n")
			b.WriteString(i18n.T("horse.promptPlayHelp") + "\n")
		}
	})
}

// HintOutput emits the current H.O.R.S.E. hint.
//
// **助言は種目に踏み込まない。** いま何の種目かを告げるところまでで、手の良し悪しは
// 種目の画面が持つ ── ここで真似ると 5 種目ぶんの戦略を二重に持つことになる。
func (p *HorseCuiPresenter) HintOutput(g interfaces.HorseGame) string {
	if g.GetGameEndFlag() {
		return i18n.T("horse.hintNone") + "\n"
	}
	if g.GetPhase() == domain.HorsePhaseHandEnd {
		return color.Yellow(i18n.T("horse.hintNextHand")) + "\n"
	}
	return color.Yellow(i18n.Tf("horse.hintDiscipline",
		"name", i18n.T("horse.discipline."+domain.HorseDisciplineName(g.GetDiscipline())))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HorseCuiPresenter) ActionLogOutput(g interfaces.HorseGame) string {
	return actionLogOutputText(g)
}

// horseCardsText は札を 1 行に並べる。
func horseCardsText(cards []*domain.Card) string {
	return formatCardSlice(cards, cuiCardStr, " ")
}
