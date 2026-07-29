//go:build !js || !wasm || extra2

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

// duchessColumnStr returns the display string for one tableau column.
func duchessColumnStr(colCards []*domain.DuchessTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// DuchessCuiPresenter renders the Duchess CUI view.
type DuchessCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DuchessCuiPresenter) Output(d interfaces.DuchessGame, lastErr error) string {
	return buildCuiOutput(i18n.T("duchess.helpTitle"), func(b *strings.Builder) {
		// 開始ランク。未選択のうちは他に何もできないので最初に大きく出す。
		if d.IsAwaitingBaseRank() {
			b.WriteString(color.Yellow(i18n.T("duchess.awaitingBase")) + "\n")
		} else {
			b.WriteString(i18n.Tf("duchess.baseRankLine",
				"rank", strconv.Itoa(d.GetBaseRank())) + "\n")
		}

		// 基礎札
		b.WriteString(i18n.T("duchess.foundationHeader"))
		foundation := d.GetFoundation()
		for i := range domain.DuchessFoundationCnt {
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

		// リザーブ扇。一番上だけが使えるので、枚数と最上段を出す。
		b.WriteString(i18n.T("duchess.reserveHeader"))
		reserve := d.GetReserve()
		for i := range domain.DuchessReserveCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			fan := reserve[i]
			if len(fan) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(i18n.Tf("duchess.fanEntry",
					"idx", strconv.Itoa(i),
					"card", cuiCardStr(fan[len(fan)-1]),
					"count", strconv.Itoa(len(fan))))
			}
		}
		b.WriteString("\n")

		// 山札とウェイスト
		b.WriteString(i18n.Tf("duchess.stockLine", "count", strconv.Itoa(d.GetStockCount())))
		waste := d.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("duchess.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("duchess.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := d.GetTableau()
		for col := range domain.DuchessTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("duchess.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(duchessColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch d.GetPhase() {
		case domain.DuchessPhasePlaying:
			if d.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := d.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("duchess.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(d.GetMoveCount())) + "\n")
		case domain.DuchessPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(d.GetMoveCount())) + "\n")
		case domain.DuchessPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Duchess hint.
func (p *DuchessCuiPresenter) HintOutput(d interfaces.DuchessGame) string {
	hint := d.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// 開始ランク未選択のときは移動ではなく「選べ」という案内になる。
	if d.IsAwaitingBaseRank() {
		return i18n.Tf("duchess.hintChooseBase", "idx", strconv.Itoa(hint.FromIdx)) + "\n"
	}
	var from string
	switch hint.FromZone {
	case "reserve":
		from = i18n.Tf("duchess.hintFromReserve", "idx", strconv.Itoa(hint.FromIdx))
	case "waste":
		from = i18n.T("duchess.hintFromWaste")
	case "stock":
		from = i18n.T("duchess.hintFromStock")
	default:
		from = i18n.Tf("duchess.hintFromTableau",
			"col", strconv.Itoa(hint.FromIdx),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("duchess.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("duchess.hintToWaste")
	default:
		to = i18n.Tf("duchess.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("duchess.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DuchessCuiPresenter) ActionLogOutput(d interfaces.DuchessGame) string {
	if d.GetPhase() == domain.DuchessPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(d.GetActionLog())
}
