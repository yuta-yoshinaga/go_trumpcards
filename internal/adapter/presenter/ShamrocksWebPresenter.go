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
	resObj.Fans = pilesToOutputSH(g.GetFans())
	foundation := g.GetFoundation()
	resObj.Foundation = pilesToOutputSH(foundation[:])

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
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
	resObj.Fans = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
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

func (p *ShamrocksWebPresenter) buildBaseOutput(g interfaces.ShamrocksGame) *controller.ShamrocksWebOutput {
	return &controller.ShamrocksWebOutput{
		Phase:       int(g.GetPhase()),
		MoveCount:   g.GetMoveCount(),
		RedealsLeft: g.GetRedealsLeft(),
		CanUndo:     g.CanUndo(),
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
