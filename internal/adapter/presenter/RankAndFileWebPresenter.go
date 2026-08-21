//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RankAndFileWebPresenter ランク・アンド・ファイルWebプレゼンタークラス
type RankAndFileWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *RankAndFileWebPresenter) Output(ft interfaces.RankAndFileGame, lastErr error) string {
	resObj := new(controller.RankAndFileWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ft, int(ft.GetPhase()))
	resObj.StockCount = ft.GetStockCount()

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
	resObj.Tableau = make([][]*controller.RankAndFileWebOutputTableauCard, domain.RankAndFileTableauCnt)
	for i := range domain.RankAndFileTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.RankAndFileWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.RankAndFileWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := ft.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.RankAndFileFoundationCnt)
	for i := range domain.RankAndFileFoundationCnt {
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
	if ft.GetPhase() == domain.RankAndFilePhasePlaying && !ft.IsStalemate() {
		if hint := ft.GetHint(); hint != nil {
			resObj.Hint = &controller.RankAndFileWebOutputHint{
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
		case domain.RankAndFilePhasePlaying:
			if ft.IsStalemate() {
				resObj.MessageCode = "rankandfile.stalemate"
			} else {
				resObj.MessageCode = "rankandfile.playing"
			}
		case domain.RankAndFilePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", ft.GetMoveCount())
			resObj.MessageCode = "rankandfile.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", ft.GetMoveCount())}
		case domain.RankAndFilePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "rankandfile.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *RankAndFileWebPresenter) HintOutput(ft interfaces.RankAndFileGame) string {
	hint := ft.GetHint()
	resObj := new(controller.RankAndFileWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, ft, int(ft.GetPhase()))
	resObj.StockCount = ft.GetStockCount()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.RankAndFileWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.RankAndFileWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "rankandfile.hintAvailable"
	} else {
		resObj.MessageCode = "rankandfile.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *RankAndFileWebPresenter) ActionLogOutput(ft interfaces.RankAndFileGame) string {
	return actionLogOutputJSON(ft)
}
