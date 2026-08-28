//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GolfWebPresenter ゴルフソリティアWebプレゼンタークラス
type GolfWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *GolfWebPresenter) Output(g interfaces.GolfGame, lastErr error) string {
	resObj := new(controller.GolfWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()

	// ウェイスト
	waste := g.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// レイアウト (7列×5段)
	layout := g.GetLayout()
	resObj.Layout = make([][]*controller.GolfWebOutputCard, domain.GolfColCnt)
	for col := range domain.GolfColCnt {
		resObj.Layout[col] = make([]*controller.GolfWebOutputCard, domain.GolfRowCnt)
		for row := range domain.GolfRowCnt {
			gc := layout[col][row]
			outGC := &controller.GolfWebOutputCard{}
			if gc != nil {
				outGC.Removed = gc.Removed
				outGC.Exposed = g.IsExposed(col, row)
				if !gc.Removed {
					outGC.Card = cardToOutput(gc.Card)
				}
			} else {
				outGC.Removed = true // nil positions are treated as removed
			}
			resObj.Layout[col][row] = outGC
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.GolfPhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.GolfWebOutputHint{
				Type: hint.Type,
				Col:  hint.Col,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := g.GetPhase()
		switch phase {
		case domain.GolfPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "golf.stalemate"
			} else {
				resObj.MessageCode = "golf.playing"
			}
		case domain.GolfPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "golf.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.GolfPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "golf.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (pr *GolfWebPresenter) HintOutput(g interfaces.GolfGame) string {
	hint := g.GetHint()
	resObj := new(controller.GolfWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Layout = make([][]*controller.GolfWebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.GolfWebOutputHint{
			Type: hint.Type,
			Col:  hint.Col,
		}
		resObj.MessageCode = "golf.hintAvailable"
	} else {
		resObj.MessageCode = "golf.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *GolfWebPresenter) ActionLogOutput(g interfaces.GolfGame) string {
	return actionLogOutputJSON(g)
}

// ResetNineHole 9ホールスコアをリセット (Webはフロントで管理するため何もしないかOutputを返す)
func (pr *GolfWebPresenter) ResetNineHole(g interfaces.GolfGame) string {
	return pr.Output(g, nil)
}
