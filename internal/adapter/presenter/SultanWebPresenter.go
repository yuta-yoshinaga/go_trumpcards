//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SultanWebPresenter スルタンWebプレゼンタークラス
type SultanWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SultanWebPresenter) Output(su interfaces.SultanGame, lastErr error) string {
	resObj := new(controller.SultanWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, su, int(su.GetPhase()))
	resObj.StockCount = su.GetStockCount()
	resObj.RedealCount = su.GetRedealCount()
	resObj.CanRedeal = su.CanRedeal()

	// ウェイスト
	waste := su.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, c := range waste {
		resObj.Waste[i] = cardToOutput(c)
	}

	// ディヴァン（プレイ済みスロットは nil）
	divan := su.GetDivan()
	resObj.Divan = make([]*controller.WebOutputCard, len(divan))
	for i, c := range divan {
		if c != nil {
			resObj.Divan[i] = cardToOutput(c)
		}
	}

	// ファンデーション
	foundation := su.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.SultanFoundationCnt)
	for i := range domain.SultanFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if su.GetPhase() == domain.SultanPhasePlaying && !su.IsStalemate() {
		if hint := su.GetHint(); hint != nil {
			resObj.Hint = &controller.SultanWebOutputHint{
				FromZone:     hint.FromZone,
				FromIdx:      hint.FromIdx,
				ToFoundation: hint.ToFoundation,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch su.GetPhase() {
		case domain.SultanPhasePlaying:
			if su.IsStalemate() {
				resObj.MessageCode = "sultan.stalemate"
			} else {
				resObj.MessageCode = "sultan.playing"
			}
		case domain.SultanPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", su.GetMoveCount())
			resObj.MessageCode = "sultan.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", su.GetMoveCount())}
		case domain.SultanPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "sultan.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SultanWebPresenter) HintOutput(su interfaces.SultanGame) string {
	hint := su.GetHint()
	resObj := new(controller.SultanWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, su, int(su.GetPhase()))
	resObj.StockCount = su.GetStockCount()
	resObj.RedealCount = su.GetRedealCount()
	resObj.CanRedeal = su.CanRedeal()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Divan = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SultanWebOutputHint{
			FromZone:     hint.FromZone,
			FromIdx:      hint.FromIdx,
			ToFoundation: hint.ToFoundation,
		}
		resObj.MessageCode = "sultan.hintAvailable"
	} else {
		resObj.MessageCode = "sultan.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SultanWebPresenter) ActionLogOutput(su interfaces.SultanGame) string {
	return actionLogOutputJSON(su)
}
