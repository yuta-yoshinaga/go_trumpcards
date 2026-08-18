//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tienLenPlayerStr returns the display string for a single Tien Len player.
func tienLenPlayerStr(player *domain.TienLenPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("tienlen.playerFinished", "rank", strconv.Itoa(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("tienlen.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// tienLenPlayTypeLabel returns the i18n label for a table combo type.
func tienLenPlayTypeLabel(pt domain.TienLenPlayType) string {
	switch pt {
	case domain.TienLenPlaySingle:
		return i18n.T("tienlen.comboSingle")
	case domain.TienLenPlayPair:
		return i18n.T("tienlen.comboPair")
	case domain.TienLenPlayTriple:
		return i18n.T("tienlen.comboTriple")
	case domain.TienLenPlayStraight:
		return i18n.T("tienlen.comboStraight")
	case domain.TienLenPlayThreePairRun:
		return i18n.T("tienlen.comboThreePairRun")
	case domain.TienLenPlayFourOfAKind:
		return i18n.T("tienlen.comboFourOfAKind")
	default:
		return i18n.T("tienlen.comboInvalid")
	}
}

// TienLenCuiPresenter renders the Tien Len CUI view.
type TienLenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TienLenCuiPresenter) Output(tg interfaces.TienLenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tienlen.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < tg.GetPlayerCnt(); i++ {
			b.WriteString(tienLenPlayerStr(tg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if tg.GetTableCards() != nil {
			// Name the combo on the table (pair/straight/…) so the player knows
			// what shape they must beat, matching the web tableCombo display.
			b.WriteString(i18n.T("tienlen.tableCards") + ": " + cuiCardSliceStr(tg.GetTableCards()) +
				" " + i18n.Tf("tienlen.tableComboType", "type", tienLenPlayTypeLabel(tg.GetTablePlayType())) + "\n")
		} else {
			b.WriteString(i18n.T("tienlen.tableEmpty") + "\n")
		}

		// Unify error display with other presenters (red, right after the board).
		cuiErrorBlock(b, lastErr)

		if tg.GetGameEndFlag() {
			b.WriteString(i18n.T("tienlen.gameEnd") + "\n")
			for i := 0; i < tg.GetPlayerCnt(); i++ {
				player := tg.GetPlayer(i)
				if player.GetRank() >= 1 {
					// 名前の組み立ては他の表示と同じ cuiPlayerName に任せる。
					// 自前で組むと太字も i18n も抜ける (#4807)。
					name := cuiPlayerName(player, i)
					fmt.Fprintf(b, "  %s: %s\n", name, i18n.Tf("tienlen.rankN", "rank", strconv.Itoa(player.GetRank())))
				}
			}
		} else {
			for _, action := range tg.GetCpuActions() {
				var actionStr string
				if action.PlayedCards == nil {
					actionStr = i18n.T("tienlen.actionPass")
				} else {
					actionStr = cuiCardSliceStr(action.PlayedCards)
				}
				fmt.Fprintf(b, "%s: %s\n", cuiPlayerName(tg.GetPlayer(action.PlayerIdx), action.PlayerIdx), actionStr)
			}
			if tg.IsHumanTurn() {
				b.WriteString(i18n.T("tienlen.yourTurn") + "\n")
				// Combo-strength rules, so CLI players can judge a legal play
				// (and self-diagnose an invalid combo shown in the error line).
				b.WriteString(i18n.T("tienlen.comboRulesHint") + "\n")
			}
		}
	})
}

// ActionLogOutput 棋譜を出力
func (p *TienLenCuiPresenter) ActionLogOutput(tg interfaces.TienLenGame) string {
	return actionLogOutputText(tg)
}

// HintOutput emits the recommended play for the human player.
//
// **ドメインの CPU 戦略をそのまま読む。**同じ盤面の判断を presenter 側に
// 書き直すと、ヒントと CPU が別のことを言い出す (#5624)。
func (p *TienLenCuiPresenter) HintOutput(tg interfaces.TienLenGame) string {
	hint := tg.GetHint()
	if hint == nil {
		return i18n.T("tienlen.hintNone") + "\n"
	}
	if hint.Pass {
		return i18n.T("tienlen.hintPass") + "\n"
	}
	player := tg.GetPlayer(tg.GetCurrentTurn())
	cards := make([]string, 0, len(hint.Indices))
	for _, idx := range hint.Indices {
		// 範囲チェックは置かない。インデックスは直前の GetHint() が**同じ手札から**
		// 作ったもので、その間に手札は変わらない。起きない場合の分岐を足すと、
		// テストの通らない行が増えるだけになる。
		//
		// 番号も出す ── `p <idx...>` にそのまま打ち込めるように。
		cards = append(cards, fmt.Sprintf("[%d]%s", idx, cuiCardStr(player.GetCard(idx))))
	}
	return i18n.Tf("tienlen.hintPlay", "cards", strings.Join(cards, " ")) + "\n"
}
