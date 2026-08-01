//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KlondikeWebPresenter クロンダイクWebプレゼンタークラス
type KlondikeWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KlondikeWebPresenter) Output(k interfaces.KlondikeGame, lastErr error) string {
	resObj := new(controller.KlondikeWebOutput)
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
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.KlondikeTableauCnt)
	for i := 0; i < domain.KlondikeTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.KlondikeWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.KlondikeWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := k.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.KlondikeFoundationCnt)
	for i := 0; i < domain.KlondikeFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if k.GetPhase() == domain.KlondikePhasePlaying && !k.IsStalemate() {
		if hint := k.GetHint(); hint != nil {
			resObj.Hint = &controller.KlondikeWebOutputHint{
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
		case domain.KlondikePhasePlaying:
			if k.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "klondike.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "klondike.stalemate"
				}
			} else {
				resObj.MessageCode = "klondike.playing"
			}
		case domain.KlondikePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", k.GetMoveCount())
			resObj.MessageCode = "klondike.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", k.GetMoveCount())}
		case domain.KlondikePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "klondike.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *KlondikeWebPresenter) HintOutput(k interfaces.KlondikeGame) string {
	hint := k.GetHint()
	resObj := new(controller.KlondikeWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, k, int(k.GetPhase()))
	resObj.StockCount = k.GetStockCount()
	resObj.DrawCount = k.GetDrawCount()
	resObj.Score = k.GetScore()
	resObj.ScoringMode = int(k.GetScoringMode())
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.KlondikeWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "klondike.hintAvailable"
	} else {
		resObj.MessageCode = "klondike.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *KlondikeWebPresenter) ActionLogOutput(k interfaces.KlondikeGame) string {
	return actionLogOutputJSON(k)
}
