//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CurdsAndWheyWebPresenter カーズ・アンド・ホエイのWebプレゼンタークラス。
type CurdsAndWheyWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
func (p *CurdsAndWheyWebPresenter) Output(g interfaces.CurdsAndWheyGame, lastErr error) string {
	resObj := p.buildBaseOutput(g)
	cols := g.GetColumns()
	resObj.Columns = pilesToOutput(cols[:])

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// このゲームは手詰まり判定を持たないので、ゲートは進行中かどうかだけ。
	if g.GetPhase() == domain.CurdsAndWheyPhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.CurdsAndWheyWebOutputHint{
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
		case domain.CurdsAndWheyPhasePlaying:
			resObj.MessageCode = "curdsandwhey.playing"
		case domain.CurdsAndWheyPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "curdsandwhey.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.CurdsAndWheyPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "curdsandwhey.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力する。
func (p *CurdsAndWheyWebPresenter) HintOutput(g interfaces.CurdsAndWheyGame) string {
	resObj := p.buildBaseOutput(g)
	resObj.Columns = make([][]*controller.WebOutputCard, 0)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.CurdsAndWheyWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "curdsandwhey.hintAvailable"
	} else {
		resObj.MessageCode = "curdsandwhey.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *CurdsAndWheyWebPresenter) ActionLogOutput(g interfaces.CurdsAndWheyGame) string {
	return actionLogOutputJSON(g)
}

func (p *CurdsAndWheyWebPresenter) buildBaseOutput(g interfaces.CurdsAndWheyGame) *controller.CurdsAndWheyWebOutput {
	return &controller.CurdsAndWheyWebOutput{
		Phase:          int(g.GetPhase()),
		MoveCount:      g.GetMoveCount(),
		CompletedSuits: g.GetCompletedSuits(),
		CanUndo:        g.CanUndo(),
	}
}
