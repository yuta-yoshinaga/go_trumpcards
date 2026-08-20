//go:build !js || !wasm || casino

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BourreCuiPresenter renders the Bourré CUI view.
type BourreCuiPresenter struct{}

// Output renders the current game state.
func (p *BourreCuiPresenter) Output(bg interfaces.BourreGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bourre.helpTitle"), func(b *strings.Builder) {
		phase := bg.GetPhase()
		b.WriteString(color.BoldYellow(bourrePhaseLabel(phase)) + "\n")
		fmt.Fprintf(b, "%s: %s   %s: %d", i18n.T("bourre.trumpSuit"),
			bourreSuitLabel(bg.GetTrumpSuit()), i18n.T("bourre.pot"), bg.GetPot())
		if bg.GetCarryPot() > 0 {
			fmt.Fprintf(b, " (+%d %s)", bg.GetCarryPot(), i18n.T("bourre.carryPot"))
		}
		// **参加の代償が画面のどこにも出ていなかった** (#5637)。0 トリックで
		// 「ブーレ」となり、domain は min(ポット, 手持ち) を罰金として取る。
		// 繰越ポットは配り直しの時点で pot に畳み込まれている (Bourre.nextHand)
		// ので、ここで足すと二重に数える。
		if phase == domain.BourrePhaseDecide && bg.IsHumanTurn() {
			penalty := bg.GetPot()
			if human := bg.GetPlayer(bg.GetCurrentPlayerIdx()); human != nil {
				penalty = min(penalty, human.GetChips())
			}
			b.WriteString("\n" + color.Yellow(i18n.Tf("bourre.decidePenalty",
				"penalty", strconv.Itoa(penalty))))
		}
		b.WriteString("\n----------\n")

		for idx := 0; idx < bg.GetPlayerCnt(); idx++ {
			b.WriteString(bourrePlayerStr(bg, bg.GetPlayer(idx), idx))
		}

		if phase == domain.BourrePhasePlay {
			b.WriteString("----------\n")
			cuiTrickBlock(b, bg.GetCurrentTrick(),
				func(tc *domain.TrickCard) int { return tc.PlayerIdx },
				func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
				func(i int) string { return bourreName(bg, i) })
		}

		if phase == domain.BourrePhaseRoundEnd || phase == domain.BourrePhaseGameEnd {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("bourre.handResult") + "\n")
			for _, r := range bg.GetLastResults() {
				line := fmt.Sprintf("  %s: %d %s", bourreName(bg, r.PlayerIdx), r.Tricks, i18n.T("bourre.tricks"))
				if r.Folded {
					line += " (" + i18n.T("bourre.folded") + ")"
				} else if r.Bourreed {
					line += " (" + i18n.T("bourre.bourreed") + ")"
				} else if r.WonAmount > 0 {
					line += fmt.Sprintf(" (+%d)", r.WonAmount)
				}
				b.WriteString(line + "\n")
			}
		}

		if phase == domain.BourrePhaseGameEnd {
			b.WriteString(color.BoldYellow(i18n.Tf("bourre.gameEndWinner", "name", bourreName(bg, bg.GetWinnerIdx()))) + "\n")
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput 棋譜をCUI出力
func (p *BourreCuiPresenter) ActionLogOutput(bg interfaces.BourreGame) string {
	return actionLogOutputTextForSeats[*domain.BourrePlayer](bg)
}

func bourrePlayerStr(bg interfaces.BourreGame, player *domain.BourrePlayer, idx int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, idx))
	fmt.Fprintf(&b, " [%s %d]", i18n.T("bourre.chips"), player.GetChips())
	if idx == bg.GetDealerIdx() {
		b.WriteString(" (D)")
	}
	switch {
	case player.GetIsFinished():
		b.WriteString(" " + i18n.T("bourre.out"))
	case player.GetFolded():
		b.WriteString(" " + i18n.T("bourre.folded"))
	default:
		b.WriteString(i18n.Tf("bourre.tricksTaken", "count", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString("\n")
	if player.GetIsHuman() && !player.GetIsFinished() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// bourreName はプレイヤー名を返す。
//
// CPU 名を自前で組むと英語リテラルがそのまま日本語ロケールに混ざるので、
// 他ゲームと同じ cuiPlayerName に任せる (#4719)。
func bourreName(bg interfaces.BourreGame, idx int) string {
	return cuiPlayerName(bg.GetPlayer(idx), idx)
}

func bourrePhaseLabel(phase domain.BourrePhase) string {
	switch phase {
	case domain.BourrePhaseDecide:
		return i18n.T("bourre.phaseDecide")
	case domain.BourrePhaseDraw:
		return i18n.T("bourre.phaseDraw")
	case domain.BourrePhasePlay:
		return i18n.T("bourre.phasePlay")
	case domain.BourrePhaseRoundEnd:
		return i18n.T("bourre.phaseRoundEnd")
	default:
		return i18n.T("bourre.phaseGameEnd")
	}
}

func bourreSuitLabel(design int) string {
	switch design {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}
