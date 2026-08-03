//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FlowerGardenWebPresenter Flower Garden Web プレゼンタークラス
type FlowerGardenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FlowerGardenWebPresenter) Output(bc interfaces.FlowerGardenGame, lastErr error) string {
	resObj := new(controller.FlowerGardenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Flower Garden is face-up by rule, but we surface
	// the domain's FaceUp field rather than hardcoding it.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.FlowerGardenWebOutputTableauCard, domain.FlowerGardenTableauCnt)
	for i := range domain.FlowerGardenTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.FlowerGardenWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.FlowerGardenWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// リザーブ — nil entries (depleted cells) serialise as null so the UI can
	// render an empty slot.
	reserve := bc.GetReserve()
	resObj.Reserve = make([]*controller.WebOutputCard, len(reserve))
	for i, c := range reserve {
		resObj.Reserve[i] = cardToOutput(c)
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FlowerGardenFoundationCnt)
	for i := range domain.FlowerGardenFoundationCnt {
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
	if bc.GetPhase() == domain.FlowerGardenPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.FlowerGardenWebOutputHint{
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
		switch bc.GetPhase() {
		case domain.FlowerGardenPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "flowergarden.stalemate"
			} else {
				resObj.MessageCode = "flowergarden.playing"
			}
		case domain.FlowerGardenPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "flowergarden.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.FlowerGardenPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "flowergarden.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FlowerGardenWebPresenter) HintOutput(bc interfaces.FlowerGardenGame) string {
	hint := bc.GetHint()
	resObj := new(controller.FlowerGardenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.FlowerGardenWebOutputTableauCard, 0)
	resObj.Reserve = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.FlowerGardenWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "flowergarden.hintAvailable"
	} else {
		resObj.MessageCode = "flowergarden.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FlowerGardenWebPresenter) ActionLogOutput(bc interfaces.FlowerGardenGame) string {
	return actionLogOutputJSON(bc)
}
