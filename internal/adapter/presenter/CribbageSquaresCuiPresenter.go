//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CribbageSquaresCuiPresenter renders the Cribbage Squares CUI view.
type CribbageSquaresCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *CribbageSquaresCuiPresenter) Output(p interfaces.CribbageSquaresGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cribbagesquares.helpTitle"), func(b *strings.Builder) {
		board := p.GetBoard()
		for r := range domain.CribbageSquaresGridSize {
			for c := range domain.CribbageSquaresGridSize {
				if c > 0 {
					b.WriteString(" | ")
				}
				rs := strconv.Itoa(r)
				cs := strconv.Itoa(c)
				if card := board[r][c]; card == nil {
					b.WriteString(i18n.Tf("cribbagesquares.cellEmpty", "r", rs, "c", cs))
				} else {
					b.WriteString(i18n.Tf("cribbagesquares.cellCard",
						"r", rs, "c", cs, "card", cuiCardStr(card)))
				}
			}
			b.WriteString(i18n.Tf("cribbagesquares.rowScore",
				"score", strconv.Itoa(p.RowScore(r))) + "\n")
		}
		b.WriteString("----------\n")

		colParts := make([]string, domain.CribbageSquaresGridSize)
		for i := range domain.CribbageSquaresGridSize {
			colParts[i] = i18n.Tf("cribbagesquares.colScore",
				"idx", strconv.Itoa(i),
				"score", strconv.Itoa(p.ColScore(i)))
		}
		b.WriteString(strings.Join(colParts, " ") + "\n")
		b.WriteString("----------\n")

		if cc := p.GetCurrentCard(); cc != nil {
			b.WriteString(i18n.Tf("cribbagesquares.currentCard", "card", cuiCardStr(cc)) + "\n")
		} else {
			b.WriteString(i18n.T("cribbagesquares.currentCardNone") + "\n")
		}

		// スターターは 16 枚置き終えるまで伏せたまま。伏せていることを
		// 明示しないと、単に表示漏れに見える。
		if st := p.GetStarter(); st != nil {
			b.WriteString(i18n.Tf("cribbagesquares.starter", "card", cuiCardStr(st)) + "\n")
		} else {
			b.WriteString(i18n.T("cribbagesquares.starterHidden") + "\n")
		}
		b.WriteString(i18n.Tf("cribbagesquares.placedLine",
			"placed", strconv.Itoa(p.GetPlacedCount()),
			"total", strconv.Itoa(domain.CribbageSquaresTotalCells),
			"score", strconv.Itoa(p.TotalScore())) + "\n")

		if p.GetPhase() != domain.CribbageSquaresPhaseComplete {
			b.WriteString(i18n.T("cribbagesquares.cuiPlaceHint") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		complete := p.GetPhase() == domain.CribbageSquaresPhaseComplete
		// 8 手それぞれの内訳。合計だけでは、どの手が効いたのか分からない。
		//
		// **途中でも出す** (#5740)。ただしスターターは 16 枚置き終えるまで
		// めくらないので、途中経過は「スターター抜きで確定している点」を出す。
		// 0 点の行・列は黙る。
		if complete {
			for r := range domain.CribbageSquaresGridSize {
				b.WriteString(cribbageSquaresDetailLine(
					i18n.Tf("cribbagesquares.rowLabel", "idx", strconv.Itoa(r)), p.RowDetail(r)))
			}
			for c := range domain.CribbageSquaresGridSize {
				b.WriteString(cribbageSquaresDetailLine(
					i18n.Tf("cribbagesquares.colLabel", "idx", strconv.Itoa(c)), p.ColDetail(c)))
			}
		} else {
			// **確定ぶんはスターター抜きで数えられる。**RowDetail は
			// スターターが出るまで 0 のままなので、途中経過は語れない。
			// 置かれている札だけを数えた下限を出す (スターターは足すだけ)。
			partial := make([]string, 0, 2*domain.CribbageSquaresGridSize)
			for r := range domain.CribbageSquaresGridSize {
				if d := p.RowPartialDetail(r); cribbageSquaresHasPoints(d) {
					partial = append(partial, cribbageSquaresDetailLine(
						i18n.Tf("cribbagesquares.rowLabel", "idx", strconv.Itoa(r)), d))
				}
			}
			for c := range domain.CribbageSquaresGridSize {
				if d := p.ColPartialDetail(c); cribbageSquaresHasPoints(d) {
					partial = append(partial, cribbageSquaresDetailLine(
						i18n.Tf("cribbagesquares.colLabel", "idx", strconv.Itoa(c)), d))
				}
			}
			if len(partial) > 0 {
				b.WriteString(i18n.T("cribbagesquares.partialHeader") + "\n")
				for _, line := range partial {
					b.WriteString(line)
				}
			}
		}

		if complete {
			line := i18n.Tf("cribbagesquares.gameComplete",
				"score", strconv.Itoa(p.TotalScore()),
				"target", strconv.Itoa(domain.CribbageSquaresWinScore))
			if p.IsWin() {
				b.WriteString(color.Green(line) + "\n")
			} else {
				b.WriteString(color.Red(line) + "\n")
			}
		}
	})
}

// cribbageSquaresHasPoints reports whether a hand scored anything at all.
//
// frontend の cribbageBreakdownParts と同じ判定: 1 つでも 0 でない要素があるか。
// 空マスだらけの行に「なし」を 8 行並べても読みづらいだけなので、途中経過では
// 点の付いた行・列だけ出す。
func cribbageSquaresHasPoints(d domain.CribbageScoreDetail) bool {
	return d.Fifteens > 0 || d.Pairs > 0 || d.Runs > 0 || d.Flush > 0 || d.Nobs > 0
}

// cribbageSquaresDetailLine renders one hand's cribbage breakdown.
//
// Zero components are dropped: a line reading "15:0 ペア:0 ラン:0" hides the
// two points that actually landed.
func cribbageSquaresDetailLine(label string, d domain.CribbageScoreDetail) string {
	parts := make([]string, 0, 5)
	for _, kv := range []struct {
		key   string
		value int
	}{
		{"cribbagesquares.partFifteens", d.Fifteens},
		{"cribbagesquares.partPairs", d.Pairs},
		{"cribbagesquares.partRuns", d.Runs},
		{"cribbagesquares.partFlush", d.Flush},
		{"cribbagesquares.partNobs", d.Nobs},
	} {
		if kv.value > 0 {
			parts = append(parts, i18n.Tf(kv.key, "n", strconv.Itoa(kv.value)))
		}
	}
	breakdown := i18n.T("cribbagesquares.partNone")
	if len(parts) > 0 {
		breakdown = strings.Join(parts, " ")
	}
	return i18n.Tf("cribbagesquares.detailLine",
		"label", label, "total", strconv.Itoa(d.Total), "breakdown", breakdown) + "\n"
}

// HintOutput emits the best-placement hint for the current card. It suggests
// the empty cell whose row and column gain the most cribbage points, or an
// explanatory line when no hint is available (game over or no current card).
func (pr *CribbageSquaresCuiPresenter) HintOutput(p interfaces.CribbageSquaresGame) string {
	hint := p.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	reason := i18n.T("cribbagesquares.hintAny")
	if hint.Synergy {
		reason = i18n.T("cribbagesquares.hintSynergy")
	}
	card := i18n.T("cribbagesquares.currentCardNone")
	if cc := p.GetCurrentCard(); cc != nil {
		card = cuiCardStr(cc)
	}
	return i18n.Tf("cribbagesquares.hintLine",
		"card", card,
		"r", strconv.Itoa(hint.Row),
		"c", strconv.Itoa(hint.Col),
		"reason", reason) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *CribbageSquaresCuiPresenter) ActionLogOutput(p interfaces.CribbageSquaresGame) string {
	if p.GetPhase() == domain.CribbageSquaresPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
