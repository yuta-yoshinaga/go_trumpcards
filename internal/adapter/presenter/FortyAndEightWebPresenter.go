//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FortyAndEightWebPresenter フォーティ・アンド・エイトWebプレゼンタークラス
type FortyAndEightWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FortyAndEightWebPresenter) Output(ft interfaces.FortyAndEightGame, lastErr error) string {
	resObj := new(controller.FortyAndEightWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ft, int(ft.GetPhase()))
	resObj.StockCount = ft.GetStockCount()
	resObj.RedealUsed = ft.GetRedealUsed()
	resObj.CanRedeal = ft.CanRedeal()

	// ウェイスト
	waste := ft.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// タブロー
	tableau := ft.GetTableau()
	resObj.Tableau = make([][]*controller.FortyAndEightWebOutputTableauCard, domain.FortyAndEightTableauCnt)
	for i := range domain.FortyAndEightTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.FortyAndEightWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.FortyAndEightWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := ft.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FortyAndEightFoundationCnt)
	for i := range domain.FortyAndEightFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if ft.GetPhase() == domain.FortyAndEightPhasePlaying && !ft.IsStalemate() {
		if hint := ft.GetHint(); hint != nil {
			resObj.Hint = &controller.FortyAndEightWebOutputHint{
				FromZone:  hint.FromZone,
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
		phase := ft.GetPhase()
		switch phase {
		case domain.FortyAndEightPhasePlaying:
			if ft.IsStalemate() {
				resObj.MessageCode = "fortyandeight.stalemate"
			} else {
				resObj.MessageCode = "fortyandeight.playing"
			}
		case domain.FortyAndEightPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", ft.GetMoveCount())
			resObj.MessageCode = "fortyandeight.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", ft.GetMoveCount())}
		case domain.FortyAndEightPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "fortyandeight.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FortyAndEightWebPresenter) HintOutput(ft interfaces.FortyAndEightGame) string {
	hint := ft.GetHint()
	resObj := new(controller.FortyAndEightWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ft, int(ft.GetPhase()))
	resObj.StockCount = ft.GetStockCount()
	resObj.RedealUsed = ft.GetRedealUsed()
	resObj.CanRedeal = ft.CanRedeal()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.FortyAndEightWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.FortyAndEightWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "fortyandeight.hintAvailable"
	} else {
		resObj.MessageCode = "fortyandeight.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FortyAndEightWebPresenter) ActionLogOutput(ft interfaces.FortyAndEightGame) string {
	return actionLogOutputJSON(ft)
}
