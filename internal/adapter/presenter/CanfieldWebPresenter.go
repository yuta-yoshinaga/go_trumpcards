//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CanfieldWebPresenter キャンフィールドWebプレゼンタークラス
type CanfieldWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CanfieldWebPresenter) Output(c interfaces.CanfieldGame, lastErr error) string {
	resObj := p.buildBaseOutput(c)

	// ウェイスト
	waste := c.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, w := range waste {
			resObj.Waste[i] = cardToOutput(w)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// リザーブ
	reserve := c.GetReserve()
	resObj.Reserve = make([]*controller.WebOutputCard, len(reserve))
	for i, rc := range reserve {
		resObj.Reserve[i] = cardToOutput(rc)
	}

	// タブロー
	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.CanfieldWebOutputTableauCard, domain.CanfieldTableauCnt)
	for i := 0; i < domain.CanfieldTableauCnt; i++ {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.CanfieldWebOutputTableauCard, len(col))
		for j, tc := range col {
			resObj.Tableau[i][j] = &controller.CanfieldWebOutputTableauCard{Card: cardToOutput(tc.Card)}
		}
	}

	// ファンデーション
	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CanfieldFoundationCnt)
	for i := 0; i < domain.CanfieldFoundationCnt; i++ {
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
	if c.GetPhase() == domain.CanfieldPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.CanfieldWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch c.GetPhase() {
		case domain.CanfieldPhasePlaying:
			resObj.MessageCode = "canfield.playing"
		case domain.CanfieldPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "canfield.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CanfieldPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "canfield.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CanfieldWebPresenter) HintOutput(c interfaces.CanfieldGame) string {
	resObj := p.buildBaseOutput(c)
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Reserve = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.CanfieldWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	hint := c.GetHint()
	if hint != nil {
		resObj.Hint = &controller.CanfieldWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "canfield.hintAvailable"
	} else {
		resObj.MessageCode = "canfield.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CanfieldWebPresenter) ActionLogOutput(c interfaces.CanfieldGame) string {
	return actionLogOutputJSON(c)
}

func (p *CanfieldWebPresenter) buildBaseOutput(c interfaces.CanfieldGame) *controller.CanfieldWebOutput {
	return &controller.CanfieldWebOutput{
		BaseRank:   c.GetBaseRank(),
		Phase:      int(c.GetPhase()),
		MoveCount:  c.GetMoveCount(),
		StockCount: c.GetStockCount(),
		CanUndo:    c.CanUndo(),
	}
}
