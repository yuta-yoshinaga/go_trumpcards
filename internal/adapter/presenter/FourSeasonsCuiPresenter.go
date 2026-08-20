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

// FourSeasonsCuiPresenter renders the Four Seasons Solitaire CUI view.
type FourSeasonsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FourSeasonsCuiPresenter) Output(f interfaces.FourSeasonsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fourseasons.helpTitle"), func(b *strings.Builder) {
		// **ベースランクを最初に出す。** 配りごとに変わり、置ける／置けないの
		// すべてがこれに乗るので、盤面より先に見えていないと読めない。
		b.WriteString(i18n.Tf("fourseasons.baseRank",
			"rank", cuiRankLabel(f.GetBaseRank())) + "\n")

		// 四隅のファンデーション。次に必要なランクも出す（暗算させない）。
		foundation := f.GetFoundations()
		maxStr := strconv.Itoa(domain.CardValueMax)
		for i := range domain.FourSeasonsFoundationCnt {
			pile := foundation[i]
			b.WriteString(i18n.Tf("fourseasons.foundationLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("fourseasons.foundationEmpty"))
				b.WriteString(i18n.Tf("fourseasons.foundationNext", "rank", cuiRankLabel(f.GetBaseRank())))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("fourseasons.foundationFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)),
					"max", maxStr))
				if len(pile) >= domain.CardValueMax {
					b.WriteString(i18n.T("fourseasons.foundationComplete"))
				} else {
					b.WriteString(i18n.Tf("fourseasons.foundationNext",
						"rank", cuiRankLabel(fourSeasonsCuiNextRank(top.GetValue()))))
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// 十字のタブロー
		tableau := f.GetTableau()
		for col := range domain.FourSeasonsTableauCnt {
			pile := tableau[col]
			b.WriteString(i18n.Tf("fourseasons.columnLabel", "col", strconv.Itoa(col)))
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("fourseasons.columnCard",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile))))
				// **組札では暗算させないのに、十字だけ暗算させていた** (#5738)。
				// タブローは下り (A の下は K) なので、折り返しを毎回自分で
				// 数えることになる。
				b.WriteString(i18n.Tf("fourseasons.columnAccepts",
					"rank", cuiRankLabel(domain.FourSeasonsPrevRank(top.GetValue()))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		b.WriteString(i18n.Tf("fourseasons.stockLine", "count", strconv.Itoa(f.GetStockCount())))
		waste := f.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("fourseasons.wasteCard", "card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("fourseasons.wasteEmpty"))
		}
		b.WriteString("\n")

		cuiErrorBlock(b, lastErr)

		switch f.GetPhase() {
		case domain.FourSeasonsPhasePlaying:
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(f.GetMoveCount())) +
				cuiSolitaireUndoHint(f.CanUndo()) + "\n")
		case domain.FourSeasonsPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(f.GetMoveCount())) + "\n")
		case domain.FourSeasonsPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// fourSeasonsCuiNextRank は表示用に K の次を A へ戻す。
// ドメイン側の同名ロジックと同じ規則だが、あちらは solo タグ内の非公開関数なので
// ここで同じ式を持つ。ずれると「次: 」の案内だけが嘘になる。
func fourSeasonsCuiNextRank(r int) int { return (r % domain.CardValueMax) + 1 }

// HintOutput emits the current Four Seasons hint.
func (p *FourSeasonsCuiPresenter) HintOutput(f interfaces.FourSeasonsGame) string {
	hint := f.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("fourseasons.hintFromTableau", "col", strconv.Itoa(hint.FromIdx))
	} else {
		from = i18n.T("fourseasons.hintFromWaste")
	}
	to := i18n.Tf("fourseasons.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	return i18n.Tf("fourseasons.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FourSeasonsCuiPresenter) ActionLogOutput(f interfaces.FourSeasonsGame) string {
	if f.GetPhase() == domain.FourSeasonsPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(f.GetActionLog())
}
