//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BigBenWebPresenter ビッグ・ベン Web プレゼンタークラス
type BigBenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BigBenWebPresenter) Output(gc interfaces.BigBenGame, lastErr error) string {
	resObj := new(controller.BigBenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, gc, int(gc.GetPhase()))

	// タブロー — 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れない
	// よう、ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := gc.GetTableau()
	resObj.StockCount = gc.GetStockCount()
	resObj.Tableau = make([][]*controller.BigBenWebOutputTableauCard, domain.BigBenTableauCnt)
	for i := range domain.BigBenTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.BigBenWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.BigBenWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 文字盤。目標ランクと完成フラグを一緒に返す — クライアントが時計の並びを
	// 自前で再計算しなくて済むし、ドメインと食い違いようがない。
	foundation := gc.GetFoundation()
	resObj.Foundation = make([]*controller.BigBenWebOutputFoundation, domain.BigBenFoundationCnt)
	for i := range domain.BigBenFoundationCnt {
		pile := foundation[i]
		cards := make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			cards[j] = cardToOutput(c)
		}
		resObj.Foundation[i] = &controller.BigBenWebOutputFoundation{
			Cards:      cards,
			TargetRank: domain.BigBenTargetRank(i),
			Complete:   gc.IsFoundationComplete(i),
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if gc.GetPhase() == domain.BigBenPhasePlaying && !gc.IsStalemate() {
		if hint := gc.GetHint(); hint != nil {
			resObj.Hint = &controller.BigBenWebOutputHint{
				FromCol: hint.FromCol,
				ToZone:  hint.ToZone,
				ToIdx:   hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch gc.GetPhase() {
		case domain.BigBenPhasePlaying:
			if gc.IsStalemate() {
				resObj.MessageCode = "bigben.stalemate"
			} else {
				resObj.MessageCode = "bigben.playing"
			}
		case domain.BigBenPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", gc.GetMoveCount())
			resObj.MessageCode = "bigben.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", gc.GetMoveCount())}
		case domain.BigBenPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "bigben.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BigBenWebPresenter) HintOutput(gc interfaces.BigBenGame) string {
	hint := gc.GetHint()
	resObj := new(controller.BigBenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, gc, int(gc.GetPhase()))
	resObj.Tableau = make([][]*controller.BigBenWebOutputTableauCard, 0)
	resObj.Foundation = make([]*controller.BigBenWebOutputFoundation, 0)

	if hint != nil {
		resObj.Hint = &controller.BigBenWebOutputHint{
			FromCol: hint.FromCol,
			ToZone:  hint.ToZone,
			ToIdx:   hint.ToIdx,
		}
		resObj.MessageCode = "bigben.hintAvailable"
	} else {
		resObj.MessageCode = "bigben.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BigBenWebPresenter) ActionLogOutput(gc interfaces.BigBenGame) string {
	return actionLogOutputJSON(gc)
}
