//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DuchessWebPresenter ダッチェス Web プレゼンタークラス
type DuchessWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *DuchessWebPresenter) Output(d interfaces.DuchessGame, lastErr error) string {
	resObj := new(controller.DuchessWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, d, int(d.GetPhase()))

	// リザーブ扇
	reserve := d.GetReserve()
	resObj.Reserve = make([][]*controller.WebOutputCard, domain.DuchessReserveCnt)
	for i := range domain.DuchessReserveCnt {
		fan := reserve[i]
		resObj.Reserve[i] = make([]*controller.WebOutputCard, len(fan))
		for j, c := range fan {
			resObj.Reserve[i][j] = cardToOutput(c)
		}
	}

	// タブロー — 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れない
	// よう、ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := d.GetTableau()
	resObj.Tableau = make([][]*controller.DuchessWebOutputTableauCard, domain.DuchessTableauCnt)
	for i := range domain.DuchessTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.DuchessWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.DuchessWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 基礎札
	foundation := d.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.DuchessFoundationCnt)
	for i := range domain.DuchessFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// 山札・ウェイスト・開始ランク。baseRank と awaitingBaseRank を両方渡すのは、
	// クライアントに「0 は未選択」という約束を再実装させないため。
	resObj.StockCount = d.GetStockCount()
	waste := d.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, c := range waste {
		resObj.Waste[i] = cardToOutput(c)
	}
	resObj.BaseRank = d.GetBaseRank()
	resObj.AwaitingBaseRank = d.IsAwaitingBaseRank()
	resObj.CanAutoComplete = d.CanAutoComplete()

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// 開始ランク待ちは Terrace と同じ理由で除外する。フロントが
	// AwaitingBaseRank から別経路で文言を出すので、重ねても二重になるだけ。
	if d.GetPhase() == domain.DuchessPhasePlaying && !d.IsAwaitingBaseRank() {
		if hint := d.GetHint(); hint != nil {
			// **CardIndex まで運ぶ。**HintOutput 側と同じ形にしないと、同じヒント
			// なのに経路によって欠ける項目が出る。
			resObj.Hint = &controller.DuchessWebOutputHint{
				FromZone:  hint.FromZone,
				FromIdx:   hint.FromIdx,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToIdx:     hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch d.GetPhase() {
		case domain.DuchessPhasePlaying:
			switch {
			case d.IsAwaitingBaseRank():
				resObj.MessageCode = "duchess.chooseBase"
			case d.IsStalemate():
				resObj.MessageCode = "duchess.stalemate"
			default:
				resObj.MessageCode = "duchess.playing"
			}
		case domain.DuchessPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", d.GetMoveCount())
			resObj.MessageCode = "duchess.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", d.GetMoveCount())}
		case domain.DuchessPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "duchess.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *DuchessWebPresenter) HintOutput(d interfaces.DuchessGame) string {
	hint := d.GetHint()
	resObj := new(controller.DuchessWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, d, int(d.GetPhase()))
	resObj.Reserve = make([][]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.DuchessWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.DuchessWebOutputHint{
			FromZone:  hint.FromZone,
			FromIdx:   hint.FromIdx,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToIdx:     hint.ToIdx,
		}
		resObj.MessageCode = "duchess.hintAvailable"
	} else {
		resObj.MessageCode = "duchess.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *DuchessWebPresenter) ActionLogOutput(d interfaces.DuchessGame) string {
	return actionLogOutputJSON(d)
}
