//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NapoleonsSquareWebPresenter ナポレオンズ・スクエア Web プレゼンタークラス
type NapoleonsSquareWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *NapoleonsSquareWebPresenter) Output(ns interfaces.NapoleonsSquareGame, lastErr error) string {
	resObj := new(controller.NapoleonsSquareWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ns, int(ns.GetPhase()))

	// タブロー — 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れないよう、
	// ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := ns.GetTableau()
	resObj.Tableau = make([][]*controller.NapoleonsSquareWebOutputTableauCard, domain.NapoleonsSquareTableauCnt)
	for i := range domain.NapoleonsSquareTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.NapoleonsSquareWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.NapoleonsSquareWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 基礎札
	foundation := ns.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.NapoleonsSquareFoundationCnt)
	for i := range domain.NapoleonsSquareFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// 山札・ウェイスト
	resObj.StockCount = ns.GetStockCount()
	waste := ns.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, c := range waste {
		resObj.Waste[i] = cardToOutput(c)
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if ns.GetPhase() == domain.NapoleonsSquarePhasePlaying && !ns.IsStalemate() {
		if hint := ns.GetHint(); hint != nil {
			resObj.Hint = &controller.NapoleonsSquareWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		if code, params := domain.ErrorMessageCode(lastErr); code != "" {
			resObj.MessageCode = code
			resObj.MessageParams = params
		} else {
			resObj.Message = lastErr.Error()
		}
	} else {
		switch ns.GetPhase() {
		case domain.NapoleonsSquarePhasePlaying:
			if ns.IsStalemate() {
				resObj.MessageCode = "napoleonssquare.stalemate"
			} else {
				resObj.MessageCode = "napoleonssquare.playing"
			}
		case domain.NapoleonsSquarePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", ns.GetMoveCount())
			resObj.MessageCode = "napoleonssquare.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", ns.GetMoveCount())}
		case domain.NapoleonsSquarePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "napoleonssquare.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *NapoleonsSquareWebPresenter) HintOutput(ns interfaces.NapoleonsSquareGame) string {
	hint := ns.GetHint()
	resObj := new(controller.NapoleonsSquareWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ns, int(ns.GetPhase()))
	resObj.Tableau = make([][]*controller.NapoleonsSquareWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.NapoleonsSquareWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "napoleonssquare.hintAvailable"
	} else {
		resObj.MessageCode = "napoleonssquare.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *NapoleonsSquareWebPresenter) ActionLogOutput(ns interfaces.NapoleonsSquareGame) string {
	return actionLogOutputJSON(ns)
}
