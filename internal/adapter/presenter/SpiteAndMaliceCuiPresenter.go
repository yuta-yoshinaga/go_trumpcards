package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SpiteAndMaliceCuiPresenter renders the Spite and Malice CUI view.
type SpiteAndMaliceCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SpiteAndMaliceCuiPresenter) Output(g interfaces.SpiteAndMaliceGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spiteandmalice.helpTitle"), func(b *strings.Builder) {
		// Foundations
		foundations := g.GetFoundations()
		maxStr := strconv.Itoa(domain.SpiteAndMaliceFoundationMax)
		for i := range domain.SpiteAndMaliceFoundationCnt {
			pile := foundations[i]
			b.WriteString(i18n.Tf("spiteandmalice.foundationLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("spiteandmalice.foundationEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("spiteandmalice.foundationFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)),
					"max", maxStr))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Per-player state
		for i := range domain.SpiteAndMalicePlayerCnt {
			pl := g.GetPlayer(i)
			if pl == nil {
				continue
			}
			label := i18n.T("spiteandmalice.labelHuman")
			if pl.GetIsCpu() {
				label = i18n.T("spiteandmalice.labelCpu")
			}
			b.WriteString(i18n.Tf("spiteandmalice.playerHeader",
				"idx", strconv.Itoa(i),
				"label", label))
			if top := pl.GoalTop(); top != nil {
				b.WriteString(i18n.Tf("spiteandmalice.goalLine",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(pl.GoalSize())))
				// **ゴール札を先に空にした方が勝ち。**Web は出せる状態のゴール札に
				// 警告色のリングを常時出しているのに、CUI は札と残り枚数だけで、
				// 毎ターン全基礎札と見比べる必要があった (#4876)。CPU 側は伏せた
				// 情報の扱いを変えないため印を付けない。
				if !pl.GetIsCpu() && g.IsGoalTopPlayable(i) {
					b.WriteString(color.Yellow(i18n.T("spiteandmalice.goalPlayable")))
				}
			} else {
				b.WriteString(i18n.T("spiteandmalice.goalEmpty"))
			}
			b.WriteString("\n")

			// Hand (face-up only for the human; CPU shows count only)
			hand := pl.GetHand()
			if !pl.GetIsCpu() {
				b.WriteString(i18n.T("spiteandmalice.humanHandLabel"))
				if len(hand) == 0 {
					b.WriteString(i18n.T("spiteandmalice.humanHandEmpty"))
				} else {
					parts := make([]string, len(hand))
					for k, c := range hand {
						parts[k] = i18n.Tf("spiteandmalice.humanHandEntry",
							"idx", strconv.Itoa(k),
							"card", cuiCardStr(c))
					}
					b.WriteString(strings.Join(parts, " "))
				}
				b.WriteString("\n")
			} else {
				b.WriteString(i18n.Tf("spiteandmalice.cpuHandLine",
					"count", strconv.Itoa(len(hand))) + "\n")
			}

			// Side piles
			for s := range domain.SpiteAndMaliceSideCnt {
				side := pl.GetSide(s)
				b.WriteString(i18n.Tf("spiteandmalice.sideLabel", "idx", strconv.Itoa(s)))
				if len(side) == 0 {
					b.WriteString(i18n.T("spiteandmalice.sideEmpty"))
				} else {
					top := side[len(side)-1]
					b.WriteString(i18n.Tf("spiteandmalice.sideFilled",
						"card", cuiCardStr(top),
						"count", strconv.Itoa(len(side))))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("----------\n")

		// Stock + completed
		b.WriteString(i18n.Tf("spiteandmalice.stockLine",
			"stock", strconv.Itoa(g.GetStockSize()),
			"completed", strconv.Itoa(g.GetCompletedSize())) + "\n")

		cuiErrorBlock(b, lastErr)

		movesStr := strconv.Itoa(g.GetMoveCount())
		switch g.GetPhase() {
		case domain.SpiteAndMalicePhasePlaying:
			b.WriteString(i18n.Tf("spiteandmalice.turnLine",
				"idx", strconv.Itoa(g.GetCurrent()),
				"moves", movesStr) + "\n")
		case domain.SpiteAndMalicePhaseGameOver:
			if g.GetWinner() == domain.SpiteAndMaliceHumanIdx {
				b.WriteString(color.Green(i18n.T("spiteandmalice.winHuman")) +
					i18n.Tf("cuiSolitaireMoves", "count", movesStr) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("spiteandmalice.winCpu")) +
					i18n.Tf("cuiSolitaireMoves", "count", movesStr) + "\n")
			}
		}
	})
}

// HintOutput emits the current Spite and Malice hint.
func (p *SpiteAndMaliceCuiPresenter) HintOutput(g interfaces.SpiteAndMaliceGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.Discard {
		return i18n.Tf("spiteandmalice.hintDiscard",
			"idx", strconv.Itoa(hint.Index),
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	}
	switch hint.Source {
	case domain.SpiteAndMaliceSourceGoal:
		return i18n.Tf("spiteandmalice.hintGoal",
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	case domain.SpiteAndMaliceSourceHand:
		return i18n.Tf("spiteandmalice.hintHand",
			"idx", strconv.Itoa(hint.Index),
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	case domain.SpiteAndMaliceSourceSide:
		return i18n.Tf("spiteandmalice.hintSide",
			"idx", strconv.Itoa(hint.Index),
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	}
	return i18n.T("cuiHintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpiteAndMaliceCuiPresenter) ActionLogOutput(g interfaces.SpiteAndMaliceGame) string {
	if g.GetPhase() == domain.SpiteAndMalicePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
