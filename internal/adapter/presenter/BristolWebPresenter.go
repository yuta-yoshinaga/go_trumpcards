//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BristolWebPresenter ブリストルWebプレゼンタークラス
type BristolWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BristolWebPresenter) Output(b interfaces.BristolGame, lastErr error) string {
	resObj := p.buildBaseOutput(b)

	// タブロー（8列）
	tableau := b.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.BristolTableauCnt)
	for i := 0; i < domain.BristolTableauCnt; i++ {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(col))
		for j, tc := range col {
			resObj.Tableau[i][j] = cardToOutput(tc)
		}
	}

	// ファン（3つ）
	fan := b.GetFan()
	resObj.Fan = make([][]*controller.WebOutputCard, domain.BristolFanCnt)
	for i := 0; i < domain.BristolFanCnt; i++ {
		pile := fan[i]
		resObj.Fan[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Fan[i][j] = cardToOutput(fc)
		}
	}

	// ファウンデーション（4つ）
	foundation := b.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.BristolFoundationCnt)
	for i := 0; i < domain.BristolFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Foundation[i][j] = cardToOutput(fc)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if b.GetPhase() == domain.BristolPhasePlaying {
		if hint := b.GetHint(); hint != nil {
			resObj.Hint = &controller.BristolWebOutputHint{
				FromZone: hint.FromZone,
				FromCol:  hint.FromCol,
				ToZone:   hint.ToZone,
				ToCol:    hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch b.GetPhase() {
		case domain.BristolPhasePlaying:
			resObj.MessageCode = "bristol.playing"
		case domain.BristolPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", b.GetMoveCount())
			resObj.MessageCode = "bristol.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", b.GetMoveCount())}
		case domain.BristolPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "bristol.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BristolWebPresenter) HintOutput(b interfaces.BristolGame) string {
	resObj := p.buildBaseOutput(b)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Fan = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	hint := b.GetHint()
	if hint != nil {
		resObj.Hint = &controller.BristolWebOutputHint{
			FromZone: hint.FromZone,
			FromCol:  hint.FromCol,
			ToZone:   hint.ToZone,
			ToCol:    hint.ToCol,
		}
		resObj.MessageCode = "bristol.hintAvailable"
	} else {
		resObj.MessageCode = "bristol.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BristolWebPresenter) ActionLogOutput(b interfaces.BristolGame) string {
	return actionLogOutputJSON(b)
}

func (p *BristolWebPresenter) buildBaseOutput(b interfaces.BristolGame) *controller.BristolWebOutput {
	return &controller.BristolWebOutput{
		Phase:      int(b.GetPhase()),
		MoveCount:  b.GetMoveCount(),
		StockCount: b.GetStockCount(),
		CanUndo:    b.CanUndo(),
	}
}
