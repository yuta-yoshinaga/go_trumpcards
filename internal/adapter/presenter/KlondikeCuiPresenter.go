//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// klondikeColumnStr returns the display string for a Klondike tableau column.
func klondikeColumnStr(colCards []*domain.KlondikeTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

// klondikeScoringModeLabel returns the localized scoring-mode name.
func klondikeScoringModeLabel(mode domain.KlondikeScoringMode) string {
	if mode == domain.KlondikeScoringVegas {
		return i18n.T("klondike.scoringVegas")
	}
	return i18n.T("klondike.scoringNone")
}

// KlondikeCuiPresenter renders the Klondike Solitaire CUI view.
//
// ベガス方式の最終得点にはタイムボーナスが乗る。それは Web だけの
// クライアントサイド機能で、ドメインにもプレゼンタにも存在しなかったため、
// **同じ盤で遊んでも CUI と Web で「最終得点」が常に食い違っていた** (#6287)。
// 経過時間はここで持つ ── GolfCuiPresenter がセッション状態を持つのと同じ形。
type KlondikeCuiPresenter struct {
	// now は差し替え可能な時計。nil のときは time.Now。テストが実時間に
	// 依存しないようにするためだけの継ぎ目。
	now       func() time.Time
	running   bool
	startedAt time.Time
	elapsed   time.Duration
}

// klondikeTimeBonus は Web の frontend/src/hooks/useKlondikeTimer.ts と
// **同じ式**: floor(700000 / 経過秒)。0 秒以下は 0 を返すのも Web と同じ。
// 式が割れると、この機能が直そうとしている食い違いをもう一度作ることになる。
func klondikeTimeBonus(d time.Duration) int {
	sec := int(d.Seconds())
	if sec <= 0 {
		return 0
	}
	return 700000 / sec
}

// clock returns the injected clock, or the wall clock.
func (p *KlondikeCuiPresenter) clock() time.Time {
	if p.now == nil {
		return time.Now()
	}
	return p.now()
}

// tickTimer starts the clock when a game begins and freezes it when one ends.
// 局面ごとに呼ばれるので、Output が何度呼ばれても計測は一度きり。
func (p *KlondikeCuiPresenter) tickTimer(phase domain.KlondikePhase) {
	switch {
	case phase == domain.KlondikePhasePlaying && !p.running:
		p.startedAt = p.clock()
		p.elapsed = 0
		p.running = true
	case phase != domain.KlondikePhasePlaying && p.running:
		p.elapsed = p.clock().Sub(p.startedAt)
		p.running = false
	}
}

// Output renders the current game state for the active locale (#1699).
func (p *KlondikeCuiPresenter) Output(k interfaces.KlondikeGame, lastErr error) string {
	p.tickTimer(k.GetPhase())
	return buildCuiOutput(i18n.T("klondike.helpTitle"), func(b *strings.Builder) {
		// Header: draw mode, scoring mode, and the running Vegas score, matching
		// the web header (none of which the CUI surfaced before).
		b.WriteString(i18n.Tf("klondike.settingsLine",
			"draw", strconv.Itoa(k.GetDrawCount()),
			"mode", klondikeScoringModeLabel(k.GetScoringMode()),
			"score", strconv.Itoa(k.GetScore())) + "\n")

		// Foundation
		b.WriteString(i18n.T("klondike.foundationHeader"))
		foundation := k.GetFoundation()
		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Stock + waste
		b.WriteString(i18n.Tf("klondike.stockLine",
			"count", strconv.Itoa(k.GetStockCount())))
		waste := k.GetWaste()
		switch {
		case len(waste) == 0:
			b.WriteString(i18n.T("klondike.wasteEmpty"))
		case k.GetDrawCount() == 3 && len(waste) > 1:
			// Three-draw: show the last up-to-3 cards as a fan; only the last
			// (top) card can be played, so mark it and note the restriction.
			start := len(waste) - 3
			if start < 0 {
				start = 0
			}
			shown := waste[start:]
			parts := make([]string, len(shown))
			for i, c := range shown {
				parts[i] = cuiCardStr(c)
			}
			parts[len(parts)-1] += "*"
			b.WriteString(i18n.Tf("klondike.wasteFan", "cards", strings.Join(parts, " ")))
		default:
			b.WriteString(i18n.Tf("klondike.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := k.GetTableau()
		for col := 0; col < domain.KlondikeTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("klondike.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch k.GetPhase() {
		case domain.KlondikePhasePlaying:
			if k.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := k.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			// **もう自動で揃えられることを能動的に知らせる (#4776)。**Web は
			// 条件が揃うとボタンを光らせバッジも出すのに、CUI は ac コマンドが
			// あること自体も、いま使えるかも出していなかった。タブロー全体を
			// 目で確認して自分で判断するしかない。
			if k.CanAutoComplete() {
				b.WriteString(color.Green(i18n.T("klondike.autoCompleteReady")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(k.GetMoveCount())) + "\n")
		case domain.KlondikePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(k.GetMoveCount())))
			if k.GetScoringMode() == domain.KlondikeScoringVegas {
				b.WriteString(" " + i18n.Tf("klondike.clearScore", "score", strconv.Itoa(k.GetScore())))
				// ボーナスは**ベガス方式のときだけ**。通常方式には最終得点が
				// 無いので、出すと存在しない数字を語ることになる。
				bonus := klondikeTimeBonus(p.elapsed)
				b.WriteString(" " + i18n.Tf("klondike.clearTimeBonus",
					"bonus", strconv.Itoa(bonus),
					"total", strconv.Itoa(k.GetScore()+bonus)))
			}
			b.WriteString("\n")
		case domain.KlondikePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Klondike hint.
func (p *KlondikeCuiPresenter) HintOutput(k interfaces.KlondikeGame) string {
	hint := k.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("klondike.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	} else {
		from = i18n.T("klondike.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("klondike.hintToFoundation")
	} else {
		to = i18n.Tf("klondike.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("klondike.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KlondikeCuiPresenter) ActionLogOutput(k interfaces.KlondikeGame) string {
	if k.GetPhase() == domain.KlondikePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(k.GetActionLog())
}
