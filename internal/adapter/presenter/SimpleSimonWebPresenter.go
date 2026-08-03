//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SimpleSimonWebPresenter シンプル・サイモンのWebプレゼンタークラス。
type SimpleSimonWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
func (p *SimpleSimonWebPresenter) Output(g interfaces.SimpleSimonGame, lastErr error) string {
	resObj := p.buildBaseOutput(g)
	cols := g.GetColumns()
	resObj.Columns = pilesToOutput(cols[:])

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// このゲームは手詰まり判定を持たないので、ゲートは進行中かどうかだけ。
	if g.GetPhase() == domain.SimpleSimonPhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.SimpleSimonWebOutputHint{
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.SimpleSimonPhasePlaying:
			resObj.MessageCode = "simplesimon.playing"
		case domain.SimpleSimonPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "simplesimon.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.SimpleSimonPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "simplesimon.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力する。
func (p *SimpleSimonWebPresenter) HintOutput(g interfaces.SimpleSimonGame) string {
	resObj := p.buildBaseOutput(g)
	resObj.Columns = make([][]*controller.WebOutputCard, 0)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.SimpleSimonWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "simplesimon.hintAvailable"
	} else {
		resObj.MessageCode = "simplesimon.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *SimpleSimonWebPresenter) ActionLogOutput(g interfaces.SimpleSimonGame) string {
	return actionLogOutputJSON(g)
}

func (p *SimpleSimonWebPresenter) buildBaseOutput(g interfaces.SimpleSimonGame) *controller.SimpleSimonWebOutput {
	return &controller.SimpleSimonWebOutput{
		Phase:          int(g.GetPhase()),
		MoveCount:      g.GetMoveCount(),
		CompletedSuits: g.GetCompletedSuits(),
		CanUndo:        g.CanUndo(),
	}
}
