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

// golfAdjacentRank reports whether two ranks differ by one, treating King(13)
// and Ace(1) as adjacent — matching the domain's isAdjacentRank and the
// frontend's isGolfAdjacent (K-A wrap included).
func golfAdjacentRank(v1, v2 int) bool {
	diff := v1 - v2
	if diff < 0 {
		diff = -diff
	}
	return diff == 1 || diff == 12
}

// golfTotalHoles is the number of deals that make up a full round, matching the
// web GUI's GOLF_TOTAL_HOLES.
const golfTotalHoles = 9

// GolfCuiPresenter renders the Golf Solitaire CUI view.
//
// **スコアカードはプレゼンターが持つ (#4784)。**9ホールは Web では
// localStorage に置かれた表示層の状態で、ドメインには存在しない。同じものを
// ドメインへ足すと KV 復元やスナップショットの対象になり、CUI にしか要らない
// 値のために Web 側の永続化を触ることになる。CUI では GameManager が
// セッションごとに1つ生成するので、ここがちょうど同じ寿命になる。
type GolfCuiPresenter struct {
	// holes は記録済みホールのスコア (ディール終了時にタブローに残った枚数)。
	holes []int
	// dealRecorded は今のディールを既に記録したか。1回のディール終了で
	// Output が何度呼ばれても二重計上しない。
	dealRecorded bool
}

// golfRemainingCount counts the cards still on the tableau — the deal's score.
// Mirrors the web GUI's countGolfRemaining; lower is better, 0 on a clear.
func golfRemainingCount(layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard) int {
	n := 0
	for col := range domain.GolfColCnt {
		for _, gc := range layout[col] {
			if gc != nil && !gc.Removed {
				n++
			}
		}
	}
	return n
}

// recordHole appends the finished deal's score, ignoring calls once the round is
// full so an extra Output cannot inflate the card.
func (pr *GolfCuiPresenter) recordHole(score int) {
	if len(pr.holes) >= golfTotalHoles {
		return
	}
	pr.holes = append(pr.holes, score)
}

// golfHoleLines writes the hole number, this deal's score and the running total,
// plus the final line once all holes are in.
func (pr *GolfCuiPresenter) golfHoleLines(b *strings.Builder) {
	if len(pr.holes) == 0 {
		return
	}
	total := 0
	for _, s := range pr.holes {
		total += s
	}
	b.WriteString(i18n.Tf("golf.holeScore",
		"hole", strconv.Itoa(len(pr.holes)),
		"holes", strconv.Itoa(golfTotalHoles),
		"score", strconv.Itoa(pr.holes[len(pr.holes)-1]),
		"total", strconv.Itoa(total)) + "\n")
	if len(pr.holes) >= golfTotalHoles {
		b.WriteString(color.Green(i18n.Tf("golf.roundComplete",
			"total", strconv.Itoa(total),
			"holes", strconv.Itoa(golfTotalHoles))) + "\n")
	}
}

// Output renders the current game state for the active locale (#1699).
func (pr *GolfCuiPresenter) Output(g interfaces.GolfGame, lastErr error) string {
	return buildCuiOutput(i18n.T("golf.helpTitle"), func(b *strings.Builder) {
		layout := g.GetLayout()

		// Waste top drives the ±1 playable check for exposed tableau cards.
		waste := g.GetWaste()
		var wasteTop *domain.Card
		if len(waste) > 0 {
			wasteTop = waste[len(waste)-1]
		}

		// Tableau (each row prints 7 columns)
		for row := range domain.GolfRowCnt {
			for col := range domain.GolfColCnt {
				if col > 0 {
					b.WriteString("  ")
				}
				gc := layout[col][row]
				switch {
				case gc == nil || gc.Removed:
					b.WriteString("    ")
				case g.IsExposed(col, row) && wasteTop != nil && golfAdjacentRank(gc.Card.GetValue(), wasteTop.GetValue()):
					// Exposed and ±1 from the waste top: playable now (trailing *).
					b.WriteString(i18n.Tf("golf.playableCard",
						"col", strconv.Itoa(col),
						"card", cuiCardStr(gc.Card)))
				case g.IsExposed(col, row):
					b.WriteString(i18n.Tf("golf.exposedCard",
						"col", strconv.Itoa(col),
						"card", cuiCardStr(gc.Card)))
				default:
					b.WriteString("   " + cuiCardStr(gc.Card))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// Stock + waste
		b.WriteString(i18n.Tf("golf.stockLine",
			"count", strconv.Itoa(g.GetStockCount())))
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("golf.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("golf.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		// ディールが終わった最初の Output で1ホール分を記録する。次のディールが
		// 始まったら (= Playing に戻ったら) また記録できるようにする。
		if g.GetPhase() == domain.GolfPhasePlaying {
			pr.dealRecorded = false
		} else if !pr.dealRecorded {
			pr.recordHole(golfRemainingCount(layout))
			pr.dealRecorded = true
		}

		switch g.GetPhase() {
		case domain.GolfPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := g.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.GolfPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.GolfPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}

		pr.golfHoleLines(b)
	})
}

// HintOutput emits the current Golf hint.
func (pr *GolfCuiPresenter) HintOutput(g interfaces.GolfGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "remove":
		return i18n.Tf("golf.hintRemove", "col", strconv.Itoa(hint.Col)) + "\n"
	case "draw":
		return i18n.T("golf.hintDraw") + "\n"
	default:
		return i18n.T("golf.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *GolfCuiPresenter) ActionLogOutput(g interfaces.GolfGame) string {
	if g.GetPhase() == domain.GolfPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
