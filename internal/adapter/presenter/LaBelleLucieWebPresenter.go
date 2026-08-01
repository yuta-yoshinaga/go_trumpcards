//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LaBelleLucieWebPresenter ラ・ベル・ルーシーのWebプレゼンタークラス。
type LaBelleLucieWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
func (p *LaBelleLucieWebPresenter) Output(g interfaces.LaBelleLucieGame, lastErr error) string {
	resObj := p.buildBaseOutput(g)
	resObj.Fans = pilesToOutput(g.GetFans())
	foundation := g.GetFoundation()
	resObj.Foundation = pilesToOutput(foundation[:])

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// このゲームは手詰まり判定を持たないので、ゲートは進行中かどうかだけ。
	if g.GetPhase() == domain.LaBelleLuciePhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.LaBelleLucieWebOutputHint{
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
		case domain.LaBelleLuciePhasePlaying:
			resObj.MessageCode = "labellelucie.playing"
		case domain.LaBelleLuciePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "labellelucie.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.LaBelleLuciePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "labellelucie.gameOver"
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力する。
func (p *LaBelleLucieWebPresenter) HintOutput(g interfaces.LaBelleLucieGame) string {
	resObj := p.buildBaseOutput(g)
	resObj.Fans = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.LaBelleLucieWebOutputHint{
			FromFan:      hint.FromFan,
			ToFan:        hint.ToFan,
			ToFoundation: hint.ToFoundation,
		}
		resObj.MessageCode = "labellelucie.hintAvailable"
	} else {
		resObj.MessageCode = "labellelucie.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (p *LaBelleLucieWebPresenter) ActionLogOutput(g interfaces.LaBelleLucieGame) string {
	return actionLogOutputJSON(g)
}

func (p *LaBelleLucieWebPresenter) buildBaseOutput(g interfaces.LaBelleLucieGame) *controller.LaBelleLucieWebOutput {
	return &controller.LaBelleLucieWebOutput{
		Phase:       int(g.GetPhase()),
		MoveCount:   g.GetMoveCount(),
		RedealsLeft: g.GetRedealsLeft(),
		CanUndo:     g.CanUndo(),
	}
}

// pilesToOutput converts a slice of card piles to web output cards.
func pilesToOutput(piles [][]*domain.Card) [][]*controller.WebOutputCard {
	out := make([][]*controller.WebOutputCard, len(piles))
	for i, pile := range piles {
		out[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			out[i][j] = cardToOutput(c)
		}
	}
	return out
}
