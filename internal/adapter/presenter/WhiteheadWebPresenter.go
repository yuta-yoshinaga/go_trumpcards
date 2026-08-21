//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WhiteheadWebPresenter ホワイトヘッドWebプレゼンタークラス
type WhiteheadWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WhiteheadWebPresenter) Output(k interfaces.WhiteheadGame, lastErr error) string {
	resObj := new(controller.WhiteheadWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, k, int(k.GetPhase()))
	resObj.StockCount = k.GetStockCount()
	resObj.DrawCount = k.GetDrawCount()
	resObj.Score = k.GetScore()
	resObj.ScoringMode = int(k.GetScoringMode())

	// ウェイスト
	waste := k.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// タブロー
	tableau := k.GetTableau()
	resObj.Tableau = make([][]*controller.WhiteheadWebOutputTableauCard, domain.WhiteheadTableauCnt)
	for i := 0; i < domain.WhiteheadTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WhiteheadWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.WhiteheadWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := k.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.WhiteheadFoundationCnt)
	for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if k.GetPhase() == domain.WhiteheadPhasePlaying && !k.IsStalemate() {
		if hint := k.GetHint(); hint != nil {
			resObj.Hint = &controller.WhiteheadWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := k.GetPhase()
		switch phase {
		case domain.WhiteheadPhasePlaying:
			if k.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "whitehead.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "whitehead.stalemate"
				}
			} else {
				resObj.MessageCode = "whitehead.playing"
			}
		case domain.WhiteheadPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", k.GetMoveCount())
			resObj.MessageCode = "whitehead.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", k.GetMoveCount())}
		case domain.WhiteheadPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "whitehead.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *WhiteheadWebPresenter) HintOutput(k interfaces.WhiteheadGame) string {
	hint := k.GetHint()
	resObj := new(controller.WhiteheadWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, k, int(k.GetPhase()))
	resObj.StockCount = k.GetStockCount()
	resObj.DrawCount = k.GetDrawCount()
	resObj.Score = k.GetScore()
	resObj.ScoringMode = int(k.GetScoringMode())
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WhiteheadWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.WhiteheadWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "whitehead.hintAvailable"
	} else {
		resObj.MessageCode = "whitehead.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WhiteheadWebPresenter) ActionLogOutput(k interfaces.WhiteheadGame) string {
	return actionLogOutputJSON(k)
}
