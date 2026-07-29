//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WindmillWebPresenter ウィンドミル Web プレゼンタークラス
type WindmillWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WindmillWebPresenter) Output(w interfaces.WindmillGame, lastErr error) string {
	resObj := new(controller.WindmillWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, w, int(w.GetPhase()))

	// 帆は 8 枠固定。補充が尽きた枠は null のまま送り、クライアント側で
	// 「枠はあるが札が無い」と描けるようにする。
	sails := w.GetSails()
	resObj.Sails = make([]*controller.WebOutputCard, domain.WindmillSailCnt)
	for i := range domain.WindmillSailCnt {
		resObj.Sails[i] = cardToOutput(sails[i])
	}

	center := w.GetCenter()
	resObj.Center = make([]*controller.WebOutputCard, len(center))
	for i, c := range center {
		resObj.Center[i] = cardToOutput(c)
	}

	corners := w.GetCorners()
	resObj.Corners = make([][]*controller.WebOutputCard, domain.WindmillCornerCnt)
	for i := range domain.WindmillCornerCnt {
		pile := corners[i]
		resObj.Corners[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Corners[i][j] = cardToOutput(c)
		}
	}

	resObj.StockCount = w.GetStockCount()
	waste := w.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, c := range waste {
		resObj.Waste[i] = cardToOutput(c)
	}
	resObj.TransferBlocked = w.IsTransferBlocked()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch w.GetPhase() {
		case domain.WindmillPhasePlaying:
			switch {
			case w.IsStalemate():
				resObj.MessageCode = "windmill.stalemate"
			case w.IsTransferBlocked():
				resObj.MessageCode = "windmill.transferBlocked"
			default:
				resObj.MessageCode = "windmill.playing"
			}
		case domain.WindmillPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", w.GetMoveCount())
			resObj.MessageCode = "windmill.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", w.GetMoveCount())}
		case domain.WindmillPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "windmill.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *WindmillWebPresenter) HintOutput(w interfaces.WindmillGame) string {
	hint := w.GetHint()
	resObj := new(controller.WindmillWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, w, int(w.GetPhase()))
	resObj.Sails = make([]*controller.WebOutputCard, 0)
	resObj.Center = make([]*controller.WebOutputCard, 0)
	resObj.Corners = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.WindmillWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "windmill.hintAvailable"
	} else {
		resObj.MessageCode = "windmill.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WindmillWebPresenter) ActionLogOutput(w interfaces.WindmillGame) string {
	return actionLogOutputJSON(w)
}
