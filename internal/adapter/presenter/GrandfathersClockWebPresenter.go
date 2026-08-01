//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GrandfathersClockWebPresenter グランドファーザーズ・クロック Web プレゼンタークラス
type GrandfathersClockWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GrandfathersClockWebPresenter) Output(gc interfaces.GrandfathersClockGame, lastErr error) string {
	resObj := new(controller.GrandfathersClockWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, gc, int(gc.GetPhase()))

	// タブロー — 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れない
	// よう、ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := gc.GetTableau()
	resObj.Tableau = make([][]*controller.GrandfathersClockWebOutputTableauCard, domain.GrandfathersClockTableauCnt)
	for i := range domain.GrandfathersClockTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.GrandfathersClockWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.GrandfathersClockWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 文字盤。目標ランクと完成フラグを一緒に返す — クライアントが時計の並びを
	// 自前で再計算しなくて済むし、ドメインと食い違いようがない。
	foundation := gc.GetFoundation()
	resObj.Foundation = make([]*controller.GrandfathersClockWebOutputFoundation, domain.GrandfathersClockFoundationCnt)
	for i := range domain.GrandfathersClockFoundationCnt {
		pile := foundation[i]
		cards := make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			cards[j] = cardToOutput(c)
		}
		resObj.Foundation[i] = &controller.GrandfathersClockWebOutputFoundation{
			Cards:      cards,
			TargetRank: domain.GrandfathersClockTargetRank(i),
			Complete:   gc.IsFoundationComplete(i),
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if gc.GetPhase() == domain.GrandfathersClockPhasePlaying && !gc.IsStalemate() {
		if hint := gc.GetHint(); hint != nil {
			resObj.Hint = &controller.GrandfathersClockWebOutputHint{
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
		case domain.GrandfathersClockPhasePlaying:
			if gc.IsStalemate() {
				resObj.MessageCode = "grandfathersclock.stalemate"
			} else {
				resObj.MessageCode = "grandfathersclock.playing"
			}
		case domain.GrandfathersClockPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", gc.GetMoveCount())
			resObj.MessageCode = "grandfathersclock.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", gc.GetMoveCount())}
		case domain.GrandfathersClockPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "grandfathersclock.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *GrandfathersClockWebPresenter) HintOutput(gc interfaces.GrandfathersClockGame) string {
	hint := gc.GetHint()
	resObj := new(controller.GrandfathersClockWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, gc, int(gc.GetPhase()))
	resObj.Tableau = make([][]*controller.GrandfathersClockWebOutputTableauCard, 0)
	resObj.Foundation = make([]*controller.GrandfathersClockWebOutputFoundation, 0)

	if hint != nil {
		resObj.Hint = &controller.GrandfathersClockWebOutputHint{
			FromCol: hint.FromCol,
			ToZone:  hint.ToZone,
			ToIdx:   hint.ToIdx,
		}
		resObj.MessageCode = "grandfathersclock.hintAvailable"
	} else {
		resObj.MessageCode = "grandfathersclock.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GrandfathersClockWebPresenter) ActionLogOutput(gc interfaces.GrandfathersClockGame) string {
	return actionLogOutputJSON(gc)
}
