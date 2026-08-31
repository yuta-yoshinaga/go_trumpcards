//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ShamrocksWebPresenter シャムロックスのWebプレゼンタークラス。
type ShamrocksWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
func (p *ShamrocksWebPresenter) Output(g interfaces.ShamrocksGame, lastErr error) string {
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
	if g.GetPhase() == domain.ShamrocksPhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.ShamrocksWebOutputHint{
				FromFan:      hint.FromFan,
				ToFan:        hint.ToFan,
				ToFoundation: hint.ToFoundation,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.ShamrocksPhasePlaying:
			resObj.MessageCode = "shamrocks.playing"
		case domain.ShamrocksPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "shamrocks.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.ShamrocksPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "shamrocks.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力する。
func (p *ShamrocksWebPresenter) HintOutput(g interfaces.ShamrocksGame) string {
	resObj := p.buildBaseOutput(g)
	p.fillBoard(resObj, g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.ShamrocksWebOutputHint{
			FromFan:      hint.FromFan,
			ToFan:        hint.ToFan,
			ToFoundation: hint.ToFoundation,
		}
		resObj.MessageCode = "shamrocks.hintAvailable"
	} else {
		resObj.MessageCode = "shamrocks.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *ShamrocksWebPresenter) ActionLogOutput(g interfaces.ShamrocksGame) string {
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
func (p *ShamrocksWebPresenter) fillBoard(resObj *controller.ShamrocksWebOutput, g interfaces.ShamrocksGame) {
	resObj.Fans = pilesToOutputSH(g.GetFans())
	foundation := g.GetFoundation()
	resObj.Foundation = pilesToOutputSH(foundation[:])
}

func (p *ShamrocksWebPresenter) buildBaseOutput(g interfaces.ShamrocksGame) *controller.ShamrocksWebOutput {
	return &controller.ShamrocksWebOutput{
		Phase:     int(g.GetPhase()),
		MoveCount: g.GetMoveCount(),
		CanUndo:   g.CanUndo(),
	}
}

// pilesToOutputSH converts a slice of card piles to web output cards.
func pilesToOutputSH(piles [][]*domain.Card) [][]*controller.WebOutputCard {
	out := make([][]*controller.WebOutputCard, len(piles))
	for i, pile := range piles {
		out[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			out[i][j] = cardToOutput(c)
		}
	}
	return out
}
