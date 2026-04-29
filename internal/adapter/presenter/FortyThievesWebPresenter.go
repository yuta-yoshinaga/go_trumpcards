package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FortyThievesWebPresenter フォーティシーブスWebプレゼンタークラス
type FortyThievesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FortyThievesWebPresenter) Output(ft interfaces.FortyThievesGame, lastErr error) string {
	resObj := new(controller.FortyThievesWebOutput)
	resObj.Phase = int(ft.GetPhase())
	resObj.MoveCount = ft.GetMoveCount()
	resObj.StockCount = ft.GetStockCount()
	resObj.CanUndo = ft.CanUndo()
	resObj.IsStalemate = ft.IsStalemate()
	resObj.UndoToEscape = ft.UndoToEscape()

	// ウェイスト
	waste := ft.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// タブロー
	tableau := ft.GetTableau()
	resObj.Tableau = make([][]*controller.FortyThievesWebOutputTableauCard, domain.FortyThievesTableauCnt)
	for i := range domain.FortyThievesTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.FortyThievesWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.FortyThievesWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := ft.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FortyThievesFoundationCnt)
	for i := range domain.FortyThievesFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := ft.GetPhase()
		switch phase {
		case domain.FortyThievesPhasePlaying:
			if ft.IsStalemate() {
				resObj.MessageCode = "fortythieves.stalemate"
			} else {
				resObj.MessageCode = "fortythieves.playing"
			}
		case domain.FortyThievesPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", ft.GetMoveCount())
			resObj.MessageCode = "fortythieves.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", ft.GetMoveCount())}
		case domain.FortyThievesPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "fortythieves.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FortyThievesWebPresenter) HintOutput(ft interfaces.FortyThievesGame) string {
	hint := ft.GetHint()
	resObj := new(controller.FortyThievesWebOutput)
	resObj.Phase = int(ft.GetPhase())
	resObj.MoveCount = ft.GetMoveCount()
	resObj.StockCount = ft.GetStockCount()
	resObj.CanUndo = ft.CanUndo()
	resObj.UndoToEscape = ft.UndoToEscape()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.FortyThievesWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.FortyThievesWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "fortythieves.hintAvailable"
	} else {
		resObj.MessageCode = "fortythieves.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FortyThievesWebPresenter) ActionLogOutput(ft interfaces.FortyThievesGame) string {
	return actionLogOutputJSON(ft)
}
