//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AmericanToadWebPresenter アメリカン・トード Web プレゼンタークラス
type AmericanToadWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AmericanToadWebPresenter) Output(at interfaces.AmericanToadGame, lastErr error) string {
	resObj := new(controller.AmericanToadWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, at, int(at.GetPhase()))

	reserve := at.GetReserve()
	resObj.Reserve = make([]*controller.WebOutputCard, len(reserve))
	for i, c := range reserve {
		resObj.Reserve[i] = cardToOutput(c)
	}

	// 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れないよう、
	// ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := at.GetTableau()
	resObj.Tableau = make([][]*controller.AmericanToadWebOutputTableauCard, domain.AmericanToadTableauCnt)
	for i := range domain.AmericanToadTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.AmericanToadWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.AmericanToadWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	foundation := at.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.AmericanToadFoundationCnt)
	for i := range domain.AmericanToadFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	resObj.StockCount = at.GetStockCount()
	waste := at.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, c := range waste {
		resObj.Waste[i] = cardToOutput(c)
	}
	resObj.BaseRank = at.GetBaseRank()
	resObj.PassesUsed = at.GetPassesUsed()
	resObj.CanRedeal = at.CanRedeal()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch at.GetPhase() {
		case domain.AmericanToadPhasePlaying:
			switch {
			case at.IsStalemate():
				resObj.MessageCode = "americantoad.stalemate"
			case at.CanRedeal():
				resObj.MessageCode = "americantoad.redealAvailable"
			default:
				resObj.MessageCode = "americantoad.playing"
			}
		case domain.AmericanToadPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", at.GetMoveCount())
			resObj.MessageCode = "americantoad.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", at.GetMoveCount())}
		case domain.AmericanToadPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "americantoad.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *AmericanToadWebPresenter) HintOutput(at interfaces.AmericanToadGame) string {
	hint := at.GetHint()
	resObj := new(controller.AmericanToadWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, at, int(at.GetPhase()))
	resObj.Reserve = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.AmericanToadWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.AmericanToadWebOutputHint{
			FromZone:  hint.FromZone,
			FromIdx:   hint.FromIdx,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToIdx:     hint.ToIdx,
		}
		resObj.MessageCode = "americantoad.hintAvailable"
	} else {
		resObj.MessageCode = "americantoad.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AmericanToadWebPresenter) ActionLogOutput(at interfaces.AmericanToadGame) string {
	return actionLogOutputJSON(at)
}
