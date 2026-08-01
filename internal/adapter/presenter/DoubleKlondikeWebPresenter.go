//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoubleKlondikeWebPresenter ダブル・クロンダイクのWebプレゼンタークラス。
type DoubleKlondikeWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
func (p *DoubleKlondikeWebPresenter) Output(g interfaces.DoubleKlondikeGame, lastErr error) string {
	resObj := p.buildBaseOutput(g)
	resObj.Tableau = p.tableauOutput(g)
	resObj.Waste = cardsToOutputOrEmpty(g.GetWaste())
	foundation := g.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, len(foundation))
	for i, f := range foundation {
		resObj.Foundation[i] = cardsToOutputOrEmpty(f)
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.DoubleKlondikePhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.DoubleKlondikeWebOutputHint{
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
		switch g.GetPhase() {
		case domain.DoubleKlondikePhasePlaying:
			resObj.MessageCode = "doubleklondike.playing"
		case domain.DoubleKlondikePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "doubleklondike.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.DoubleKlondikePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "doubleklondike.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// tableauOutput 表向きカードのみ可視化し、裏向きは隠す。
func (p *DoubleKlondikeWebPresenter) tableauOutput(g interfaces.DoubleKlondikeGame) [][]*controller.DoubleKlondikeWebOutputTableauCard {
	tableau := g.GetTableau()
	out := make([][]*controller.DoubleKlondikeWebOutputTableauCard, len(tableau))
	for i, col := range tableau {
		out[i] = make([]*controller.DoubleKlondikeWebOutputTableauCard, len(col))
		for j, tc := range col {
			cell := &controller.DoubleKlondikeWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				cell.Card = cardToOutput(tc.Card)
			}
			out[i][j] = cell
		}
	}
	return out
}

// HintOutput ヒントをJSON出力する。
func (p *DoubleKlondikeWebPresenter) HintOutput(g interfaces.DoubleKlondikeGame) string {
	resObj := p.buildBaseOutput(g)
	resObj.Tableau = make([][]*controller.DoubleKlondikeWebOutputTableauCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.DoubleKlondikeWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "doubleklondike.hintAvailable"
	} else {
		resObj.MessageCode = "doubleklondike.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *DoubleKlondikeWebPresenter) ActionLogOutput(g interfaces.DoubleKlondikeGame) string {
	return actionLogOutputJSON(g)
}

func (p *DoubleKlondikeWebPresenter) buildBaseOutput(g interfaces.DoubleKlondikeGame) *controller.DoubleKlondikeWebOutput {
	return &controller.DoubleKlondikeWebOutput{
		Phase:       int(g.GetPhase()),
		MoveCount:   g.GetMoveCount(),
		StockCount:  g.GetStockCount(),
		CanUndo:     g.CanUndo(),
		IsStalemate: g.IsStalemate(),
	}
}
