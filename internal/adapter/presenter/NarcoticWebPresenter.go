//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NarcoticWebPresenter ナルコティックWebプレゼンタークラス
type NarcoticWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *NarcoticWebPresenter) Output(g interfaces.NarcoticGame, lastErr error) string {
	resObj := new(controller.NarcoticWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	resObj.DiscardCount = g.GetDiscardCount()
	resObj.DiscardTop = cardToOutput(g.GetDiscardTop())
	resObj.Columns = narcoticColumns(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.NarcoticPhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.NarcoticWebOutputHint{
				Type: hint.Type,
				Col:  hint.Col,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.NarcoticPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "narcotic.stalemate"
			} else {
				resObj.MessageCode = "narcotic.playing"
			}
		case domain.NarcoticPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "narcotic.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.NarcoticPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "narcotic.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// narcoticColumns はドメインの列表現をWeb出力に変換する。
func narcoticColumns(g interfaces.NarcoticGame) [][]*controller.NarcoticWebOutputCard {
	cols := g.GetColumns()
	out := make([][]*controller.NarcoticWebOutputCard, domain.NarcoticColCnt)
	for c := range domain.NarcoticColCnt {
		col := cols[c]
		out[c] = make([]*controller.NarcoticWebOutputCard, len(col))
		for i, card := range col {
			top := i == len(col)-1
			outCard := &controller.NarcoticWebOutputCard{
				Card: cardToOutput(card),
				Top:  top,
			}
			if top {
				// **Removable は盤面全体の性質。**4枚揃ったときだけ真になり、
				// そのときは4列とも真になる (クローン元は列ごとに違った)。
				outCard.Removable = g.CanRemoveSet()
				outCard.Movable = g.CanMove(c)
			}
			out[c][i] = outCard
		}
	}
	return out
}

// HintOutput ヒントをJSON出力
func (pr *NarcoticWebPresenter) HintOutput(g interfaces.NarcoticGame) string {
	hint := g.GetHint()
	resObj := new(controller.NarcoticWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	resObj.DiscardCount = g.GetDiscardCount()
	resObj.Columns = make([][]*controller.NarcoticWebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.NarcoticWebOutputHint{
			Type: hint.Type,
			Col:  hint.Col,
		}
		resObj.MessageCode = "narcotic.hintAvailable"
	} else {
		resObj.MessageCode = "narcotic.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *NarcoticWebPresenter) ActionLogOutput(g interfaces.NarcoticGame) string {
	return actionLogOutputJSON(g)
}
