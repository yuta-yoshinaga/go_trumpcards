//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BlackHoleWebPresenter ブラックホールのWebプレゼンタークラス。
type BlackHoleWebPresenter struct{}

// blackHoleFansOutput converts the fans to web output cards.
func blackHoleFansOutput(fans [][]*domain.Card) [][]*controller.WebOutputCard {
	out := make([][]*controller.WebOutputCard, len(fans))
	for i, fan := range fans {
		out[i] = cardsToOutputOrEmpty(fan)
	}
	return out
}

// Output ゲーム状態をJSON出力する。
func (p *BlackHoleWebPresenter) Output(g interfaces.BlackHoleGame, lastErr error) string {
	resObj := p.buildBaseOutput(g)
	resObj.Fans = blackHoleFansOutput(g.GetFans())
	resObj.BlackHole = cardsToOutputOrEmpty(g.GetBlackHole())

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.BlackHolePhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.BlackHoleWebOutputHint{Fan: hint.Fan}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.BlackHolePhasePlaying:
			resObj.MessageCode = "blackhole.playing"
		case domain.BlackHolePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "blackhole.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.BlackHolePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "blackhole.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力する。
//
// The board (fans + black hole) is included so the client can keep rendering the
// tableau after a hint request: the Black Hole page replaces its whole state from
// this response, so an empty board would blank the tableau. The recommended fan is
// surfaced via Hint for a two-tier highlight (strong ring on the recommended fan,
// weaker ring on the other legal fans).
func (p *BlackHoleWebPresenter) HintOutput(g interfaces.BlackHoleGame) string {
	resObj := p.buildBaseOutput(g)
	resObj.Fans = blackHoleFansOutput(g.GetFans())
	resObj.BlackHole = cardsToOutputOrEmpty(g.GetBlackHole())
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BlackHoleWebOutputHint{Fan: hint.Fan}
		resObj.MessageCode = "blackhole.hintAvailable"
	} else {
		resObj.MessageCode = "blackhole.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *BlackHoleWebPresenter) ActionLogOutput(g interfaces.BlackHoleGame) string {
	return actionLogOutputJSON(g)
}

func (p *BlackHoleWebPresenter) buildBaseOutput(g interfaces.BlackHoleGame) *controller.BlackHoleWebOutput {
	return &controller.BlackHoleWebOutput{
		Phase:       int(g.GetPhase()),
		MoveCount:   g.GetMoveCount(),
		CanUndo:     g.CanUndo(),
		IsStalemate: g.IsStalemate(),
	}
}
