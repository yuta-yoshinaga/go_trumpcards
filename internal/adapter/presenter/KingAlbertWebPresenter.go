//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KingAlbertWebPresenter King Albert Web プレゼンタークラス
type KingAlbertWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KingAlbertWebPresenter) Output(bc interfaces.KingAlbertGame, lastErr error) string {
	resObj := new(controller.KingAlbertWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in King Albert is face-up by rule, but we surface
	// the domain's FaceUp field rather than hardcoding it.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.KingAlbertWebOutputTableauCard, domain.KingAlbertTableauCnt)
	for i := range domain.KingAlbertTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.KingAlbertWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.KingAlbertWebOutputTableauCard{
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
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.KingAlbertFoundationCnt)
	for i := range domain.KingAlbertFoundationCnt {
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
	if bc.GetPhase() == domain.KingAlbertPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.KingAlbertWebOutputHint{
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
		case domain.KingAlbertPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "kingalbert.stalemate"
			} else {
				resObj.MessageCode = "kingalbert.playing"
			}
		case domain.KingAlbertPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "kingalbert.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.KingAlbertPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "kingalbert.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *KingAlbertWebPresenter) HintOutput(bc interfaces.KingAlbertGame) string {
	hint := bc.GetHint()
	resObj := new(controller.KingAlbertWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.KingAlbertWebOutputTableauCard, 0)
	resObj.Reserve = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.KingAlbertWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "kingalbert.hintAvailable"
	} else {
		resObj.MessageCode = "kingalbert.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *KingAlbertWebPresenter) ActionLogOutput(bc interfaces.KingAlbertGame) string {
	return actionLogOutputJSON(bc)
}
