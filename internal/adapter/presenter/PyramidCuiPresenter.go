//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// PyramidCuiPresenter renders the Pyramid Solitaire CUI view.
type PyramidCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
// pyramidSessionStatsLine renders this run's play/clear counts and best move
// count, or "" before the first game finishes.
//
// **The web keeps a persistent record in localStorage; this does not.** The CUI
// has no store of its own, so the figures cover the current process only -- the
// wording says so rather than implying an all-time best.
func pyramidSessionStatsLine(p interfaces.PyramidGame) string {
	plays := p.GetSessionPlays()
	if plays == 0 {
		return ""
	}
	best := i18n.T("pyramid.sessionNoBest")
	if fm := p.GetSessionFewestMoves(); fm > 0 {
		best = strconv.Itoa(fm)
	}
	return i18n.Tf("pyramid.sessionStats",
		"plays", strconv.Itoa(plays),
		"wins", strconv.Itoa(p.GetSessionWins()),
		"best", best) + "\n"
}

func (pr *PyramidCuiPresenter) Output(p interfaces.PyramidGame, lastErr error) string {
	const pyramidIndent = "  "
	const pyramidRemovedPlaceholder = "    "
	return buildCuiOutput(i18n.T("pyramid.helpTitle"), func(b *strings.Builder) {
		// Pyramid layout (triangular)
		pyramid := p.GetPyramid()
		for row := range domain.PyramidRowCnt {
			indent := strings.Repeat(pyramidIndent, domain.PyramidRowCnt-1-row)
			b.WriteString(indent)
			for col := range row + 1 {
				if col > 0 {
					b.WriteString(pyramidIndent)
				}
				pc := pyramid[row][col]
				switch {
				case pc.Removed:
					b.WriteString(pyramidRemovedPlaceholder)
				case p.IsRemovableKing(row, col):
					// **キングは相方が要らず単独で消せる (#4782)。**Web は常時
					// ハイライトしているのに、CUI は数値を1枚ずつ見て 13 を
					// 自分で探すしかなかった。
					b.WriteString(i18n.Tf("pyramid.exposedKing",
						"row", strconv.Itoa(row),
						"col", strconv.Itoa(col),
						"card", cuiCardStr(pc.Card)))
				case p.IsExposed(row, col):
					// Exposed cards carry their coordinates so they can be played.
					b.WriteString(i18n.Tf("pyramid.exposedCard",
						"row", strconv.Itoa(row),
						"col", strconv.Itoa(col),
						"card", cuiCardStr(pc.Card)))
				default:
					// Blocked cards hide coordinates to signal they cannot be taken yet.
					b.WriteString(i18n.Tf("pyramid.blockedCard",
						"card", cuiCardStr(pc.Card)))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// Stock + waste
		b.WriteString(i18n.Tf("pyramid.stockLine",
			"count", strconv.Itoa(p.GetStockCount())))
		// **配り直しは無い。** Draw は山札を引き切ると二度と引けない設計なのに、
		// 残り枚数しか出しておらず、標準のピラミッド (3回配り直し可) を知っている
		// プレイヤーほど手詰まりの原因を誤解する (#5510)。
		if p.GetStockCount() == 0 {
			b.WriteString(" " + color.Yellow(i18n.T("pyramid.noRedeal")))
		}
		waste := p.GetWaste()
		if len(waste) > 0 {
			wasteKey := "pyramid.wasteCard"
			if p.IsWasteKingRemovable() {
				wasteKey = "pyramid.wasteKing"
			}
			b.WriteString(i18n.Tf(wasteKey,
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("pyramid.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch p.GetPhase() {
		case domain.PyramidPhasePlaying:
			if p.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := p.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(p.GetMoveCount())) +
				cuiSolitaireUndoHint(p.CanUndo()) + "\n")
			b.WriteString(pyramidSessionStatsLine(p))
		case domain.PyramidPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(p.GetMoveCount())) + "\n")
		case domain.PyramidPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Pyramid hint.
func (pr *PyramidCuiPresenter) HintOutput(p interfaces.PyramidGame) string {
	hint := p.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "king":
		return i18n.Tf("pyramid.hintKing",
			"row", strconv.Itoa(hint.Row1),
			"col", strconv.Itoa(hint.Col1)) + "\n"
	case "pair":
		return i18n.Tf("pyramid.hintPair",
			"row1", strconv.Itoa(hint.Row1),
			"col1", strconv.Itoa(hint.Col1),
			"row2", strconv.Itoa(hint.Row2),
			"col2", strconv.Itoa(hint.Col2)) + "\n"
	case "waste_king":
		return i18n.T("pyramid.hintWasteKing") + "\n"
	case "waste_pair":
		return i18n.Tf("pyramid.hintWastePair",
			"row", strconv.Itoa(hint.Row1),
			"col", strconv.Itoa(hint.Col1)) + "\n"
	default:
		return i18n.T("pyramid.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *PyramidCuiPresenter) ActionLogOutput(p interfaces.PyramidGame) string {
	if p.GetPhase() == domain.PyramidPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
