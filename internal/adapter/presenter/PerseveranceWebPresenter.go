//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PerseveranceWebPresenter パーシビアランスWebプレゼンタークラス
type PerseveranceWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PerseveranceWebPresenter) Output(bd interfaces.PerseveranceGame, lastErr error) string {
	resObj := new(controller.PerseveranceWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bd, int(bd.GetPhase()))

	// タブロー — every card in Perseverance is face-up by rule, so we always
	// serialise tc.Card and the FaceUp field is effectively a constant true.
	tableau := bd.GetTableau()
	resObj.Tableau = make([][]*controller.PerseveranceWebOutputTableauCard, domain.PerseveranceTableauCnt)
	for i := range domain.PerseveranceTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.PerseveranceWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.PerseveranceWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: true,
			}
		}
	}

	// ファンデーション
	foundation := bd.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.PerseveranceFoundationCnt)
	for i := range domain.PerseveranceFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	resObj.RedealsLeft = bd.GetRedealsLeft()

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if bd.GetPhase() == domain.PerseverancePhasePlaying && !bd.IsStalemate() {
		if hint := bd.GetHint(); hint != nil {
			resObj.Hint = &controller.PerseveranceWebOutputHint{
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
		phase := bd.GetPhase()
		switch phase {
		case domain.PerseverancePhasePlaying:
			if bd.IsStalemate() {
				resObj.MessageCode = "perseverance.stalemate"
			} else {
				resObj.MessageCode = "perseverance.playing"
			}
		case domain.PerseverancePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bd.GetMoveCount())
			resObj.MessageCode = "perseverance.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bd.GetMoveCount())}
		case domain.PerseverancePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "perseverance.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *PerseveranceWebPresenter) HintOutput(bd interfaces.PerseveranceGame) string {
	hint := bd.GetHint()
	resObj := new(controller.PerseveranceWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bd, int(bd.GetPhase()))
	resObj.Tableau = make([][]*controller.PerseveranceWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.PerseveranceWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "perseverance.hintAvailable"
	} else {
		resObj.MessageCode = "perseverance.noHint"
	}

	return marshalOrError(resObj)
}

// TargetsOutput は Web では通常の盤面をそのまま返す。
//
// 置ける先の強調は `perseveranceLegalTargets` がこの盤面から作っている ──
// Web には `targets` に当たる操作が無く、選択した瞬間に見えている (#5581)。
// ここで別の形を返すと、CUI 専用の応答が Web の経路に紛れ込む。
func (p *PerseveranceWebPresenter) TargetsOutput(bd interfaces.PerseveranceGame, _ int) string {
	return p.Output(bd, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PerseveranceWebPresenter) ActionLogOutput(bd interfaces.PerseveranceGame) string {
	return actionLogOutputJSON(bd)
}
