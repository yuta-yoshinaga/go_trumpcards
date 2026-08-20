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

// andarBaharHistoryMaxShown は罫線に出す直近の件数の上限。
const andarBaharHistoryMaxShown = 20

// andarBaharColumnMaxShown は各列に出す札の枚数の上限 (末尾から)。
const andarBaharColumnMaxShown = 12

// AndarBaharCuiPresenter アンダーバハールCUIプレゼンタークラス
type AndarBaharCuiPresenter struct{}

// Output ゲーム状態を出力
func (ap *AndarBaharCuiPresenter) Output(ab interfaces.AndarBaharGame, lastErr error) string {
	return buildCuiOutput(i18n.T("andarbahar.outputTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("andarbahar.chipsLine", "chips", strconv.Itoa(ab.GetChips())) + "\n")
		b.WriteString(i18n.Tf("andarbahar.phaseLine", "phase", ap.phaseStr(ab.GetPhase())) + "\n")

		if joker := ab.GetJoker(); joker != nil {
			b.WriteString(i18n.Tf("andarbahar.jokerLine", "card", cuiCardStr(joker)) + "\n")
		}
		// **先に配る列は 0.9:1 に下がります。** 賭ける前に分かるようにします。
		b.WriteString(i18n.Tf("andarbahar.firstColumnLine",
			"column", ap.columnStr(ab.GetFirstColumn())) + "\n")

		if ab.GetBetAmount() > 0 {
			b.WriteString(i18n.Tf("andarbahar.betLine",
				"amount", strconv.Itoa(ab.GetBetAmount()),
				"target", ap.columnStr(ab.GetBetTarget()),
			) + "\n")
		}
		if ab.GetSideBand() != domain.AndarBaharSideNone {
			b.WriteString(i18n.Tf("andarbahar.sideBetLine",
				"amount", strconv.Itoa(ab.GetSideAmount()),
				"band", ap.bandStr(ab.GetSideBand()),
			) + "\n")
		}

		andar, bahar := ab.GetAndarCards(), ab.GetBaharCards()
		if len(andar) > 0 || len(bahar) > 0 {
			b.WriteString("----------\n")
			b.WriteString(i18n.Tf("andarbahar.andarLine", "cards", ap.columnCards(andar)) + "\n")
			b.WriteString(i18n.Tf("andarbahar.baharLine", "cards", ap.columnCards(bahar)) + "\n")
			b.WriteString(i18n.Tf("andarbahar.dealtCountLine",
				"count", strconv.Itoa(ab.DealtCount())) + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if ab.GetGameEndFlag() {
			msg := i18n.Tf("andarbahar.winnerLine", "column", ap.columnStr(ab.GetWinner()))
			if ab.GetResult() == domain.GameResultWin {
				b.WriteString(color.Green(msg) + "\n")
			} else {
				b.WriteString(color.Red(msg) + "\n")
			}
			b.WriteString(i18n.Tf("andarbahar.payoutLine",
				"payout", strconv.Itoa(ab.GetPayout())) + "\n")
			// **サイドベットは別の賭け** (#5770)。合計だけでは、外したのが
			// メインなのかサイドなのか分からない。張った回だけ内訳を出す。
			if ab.GetSideBand() != domain.AndarBaharSideNone {
				b.WriteString(i18n.Tf("andarbahar.payoutBreakdownLine",
					"main", strconv.Itoa(ab.GetMainPayout()),
					"side", strconv.Itoa(ab.GetSidePayout())) + "\n")
			}
		}

		ap.writeHistory(b, ab.GetHistory())
	})
}

// columnCards は列の札を末尾から andarBaharColumnMaxShown 枚まで並べる。
func (ap *AndarBaharCuiPresenter) columnCards(cards []*domain.Card) string {
	if len(cards) == 0 {
		return i18n.T("andarbahar.noCards")
	}
	shown := cards
	prefix := ""
	if len(shown) > andarBaharColumnMaxShown {
		shown = shown[len(shown)-andarBaharColumnMaxShown:]
		prefix = "... "
	}
	parts := make([]string, len(shown))
	for i, c := range shown {
		parts[i] = cuiCardStr(c)
	}
	return prefix + strings.Join(parts, " ")
}

// writeHistory は罫線を書き出す。
func (ap *AndarBaharCuiPresenter) writeHistory(b *strings.Builder, history []int) {
	if len(history) == 0 {
		return
	}
	andarCount, baharCount := 0, 0
	for _, r := range history {
		if r == domain.AndarBaharBetAndar {
			andarCount++
		} else {
			baharCount++
		}
	}
	shown := history
	if len(shown) > andarBaharHistoryMaxShown {
		shown = shown[len(shown)-andarBaharHistoryMaxShown:]
	}
	syms := make([]string, len(shown))
	for i, r := range shown {
		if r == domain.AndarBaharBetAndar {
			syms[i] = color.Red("A")
		} else {
			syms[i] = color.Yellow("B")
		}
	}
	b.WriteString(i18n.Tf("andarbahar.historyLine", "symbols", strings.Join(syms, " ")) + "\n")
	b.WriteString(i18n.Tf("andarbahar.historyCounts",
		"andar", strconv.Itoa(andarCount),
		"bahar", strconv.Itoa(baharCount)) + "\n")
}

// ActionLogOutput 棋譜をテキスト出力
func (ap *AndarBaharCuiPresenter) ActionLogOutput(ab interfaces.AndarBaharGame) string {
	return actionLogOutputText(ab)
}

// HintOutput ヒントをテキスト出力
func (ap *AndarBaharCuiPresenter) HintOutput(ab interfaces.AndarBaharGame) string {
	return i18n.T("andarbahar." + ab.GetHint())
}

// phaseStr フェーズ文字列
func (ap *AndarBaharCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.AndarBaharPhaseBet:
		return i18n.T("andarbahar.phaseBet")
	case domain.AndarBaharPhaseEnd:
		return i18n.T("andarbahar.phaseEnd")
	default:
		return i18n.T("andarbahar.phaseUnknown")
	}
}

// columnStr 列名の文字列
func (ap *AndarBaharCuiPresenter) columnStr(col int) string {
	switch col {
	case domain.AndarBaharBetAndar:
		return i18n.T("andarbahar.columnAndar")
	case domain.AndarBaharBetBahar:
		return i18n.T("andarbahar.columnBahar")
	default:
		return i18n.T("andarbahar.columnUnknown")
	}
}

// bandStr サイドベットの帯の文字列
func (ap *AndarBaharCuiPresenter) bandStr(band int) string {
	lo, hi, ok := domain.AndarBaharSideBand(band)
	if !ok {
		return i18n.T("andarbahar.bandUnknown")
	}
	if lo == hi {
		return i18n.Tf("andarbahar.bandExact", "n", strconv.Itoa(lo))
	}
	return i18n.Tf("andarbahar.bandRange", "lo", strconv.Itoa(lo), "hi", strconv.Itoa(hi))
}
