//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TriPeaksWebPresenter トリピークスWebプレゼンタークラス
type TriPeaksWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *TriPeaksWebPresenter) Output(t interfaces.TriPeaksGame, lastErr error) string {
	resObj := new(controller.TriPeaksWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, t, int(t.GetPhase()))
	resObj.StockCount = t.GetStockCount()
	resObj.Score, resObj.Combo = t.GetScore(), t.GetCombo()

	// ウェイスト
	waste := t.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// レイアウト
	layout := t.GetLayout()
	resObj.Layout = make([][]*controller.TriPeaksWebOutputCard, domain.TriPeaksRowCnt)
	for row := range domain.TriPeaksRowCnt {
		resObj.Layout[row] = make([]*controller.TriPeaksWebOutputCard, domain.TriPeaksColCnt)
		for col := range domain.TriPeaksColCnt {
			tc := layout[row][col]
			outTC := &controller.TriPeaksWebOutputCard{}
			if tc != nil {
				outTC.Removed = tc.Removed
				outTC.Exposed = t.IsExposed(row, col)
				if !tc.Removed {
					outTC.Card = cardToOutput(tc.Card)
				}
			} else {
				outTC.Removed = true // nil positions are treated as removed
			}
			resObj.Layout[row][col] = outTC
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if t.GetPhase() == domain.TriPeaksPhasePlaying && !t.IsStalemate() {
		if hint := t.GetHint(); hint != nil {
			resObj.Hint = &controller.TriPeaksWebOutputHint{
				Type: hint.Type,
				Row:  hint.Row,
				Col:  hint.Col,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := t.GetPhase()
		switch phase {
		case domain.TriPeaksPhasePlaying:
			if t.IsStalemate() {
				resObj.MessageCode = "tripeaks.stalemate"
			} else {
				resObj.MessageCode = "tripeaks.playing"
			}
		case domain.TriPeaksPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", t.GetMoveCount())
			resObj.MessageCode = "tripeaks.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", t.GetMoveCount())}
		case domain.TriPeaksPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "tripeaks.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (pr *TriPeaksWebPresenter) HintOutput(t interfaces.TriPeaksGame) string {
	hint := t.GetHint()
	resObj := new(controller.TriPeaksWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, t, int(t.GetPhase()))
	resObj.StockCount = t.GetStockCount()
	resObj.Score, resObj.Combo = t.GetScore(), t.GetCombo()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Layout = make([][]*controller.TriPeaksWebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.TriPeaksWebOutputHint{
			Type: hint.Type,
			Row:  hint.Row,
			Col:  hint.Col,
		}
		resObj.MessageCode = "tripeaks.hintAvailable"
	} else {
		resObj.MessageCode = "tripeaks.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *TriPeaksWebPresenter) ActionLogOutput(t interfaces.TriPeaksGame) string {
	return actionLogOutputJSON(t)
}
