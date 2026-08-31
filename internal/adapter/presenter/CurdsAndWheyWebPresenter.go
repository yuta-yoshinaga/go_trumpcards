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
	p.fillBoard(resObj, g)

	// **受動ヒントは Output() でも埋める。**盤面のリングなど `state.hint` を読む
	// 分岐は、ヒントボタンを押していないときにも動く。ここで埋めないとそれらは
	// 全部死ぬ (#4483)。
	//
	// **「HintOutput の応答はページの state にマージされない」と書いてあったが、
	// このページでは誤り。**ヒントボタンは `useGameApi` の `exec('hint')` を呼び、
	// `setState(res)` が状態を丸ごと差し替える。だから HintOutput も
	// `fillBoard` で盤面を返す (#6800 / #6855)。
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
	p.fillBoard(resObj, g)
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

// fillBoard は盤面を埋める。
//
// **Output と HintOutput が同じ盤面を返すためにある。**HintOutput は以前
// 盤面を空配列で潰していた ── `buildBaseOutput` は盤面を埋めないので、
// 潰さないと JSON に `null` が出てフロントの `.map` が壊れる、という理由だった。
// だがこのページのヒントボタン (と CLI の hint) は `useGameApi` の `exec` を
// 呼び、`useGameApi` は `setState(res)` で状態を**丸ごと差し替える**ので、
// その空配列がそのまま画面に流れ込んで盤面が消えていた (#6855、#6800 と同型)。
func (p *CurdsAndWheyWebPresenter) fillBoard(resObj *controller.CurdsAndWheyWebOutput, g interfaces.CurdsAndWheyGame) {
	cols := g.GetColumns()
	resObj.Columns = pilesToOutput(cols[:])
}

func (p *CurdsAndWheyWebPresenter) buildBaseOutput(g interfaces.CurdsAndWheyGame) *controller.CurdsAndWheyWebOutput {
	return &controller.CurdsAndWheyWebOutput{
		Phase:          int(g.GetPhase()),
		MoveCount:      g.GetMoveCount(),
		CompletedSuits: g.GetCompletedSuits(),
		CanUndo:        g.CanUndo(),
	}
}
