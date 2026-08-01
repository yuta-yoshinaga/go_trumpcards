//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// StreetsAndAlleysWebPresenter Streets and Alleys Web プレゼンタークラス
type StreetsAndAlleysWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *StreetsAndAlleysWebPresenter) Output(bc interfaces.StreetsAndAlleysGame, lastErr error) string {
	resObj := new(controller.StreetsAndAlleysWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Streets and Alleys is face-up by rule, but we
	// surface the domain's FaceUp field rather than hardcoding it so a future
	// hidden-deal variant would not silently leak state through this presenter.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.StreetsAndAlleysWebOutputTableauCard, domain.StreetsAndAlleysTableauCnt)
	for i := range domain.StreetsAndAlleysTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.StreetsAndAlleysWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.StreetsAndAlleysWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.StreetsAndAlleysFoundationCnt)
	for i := range domain.StreetsAndAlleysFoundationCnt {
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
	if bc.GetPhase() == domain.StreetsAndAlleysPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.StreetsAndAlleysWebOutputHint{
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
		phase := bc.GetPhase()
		switch phase {
		case domain.StreetsAndAlleysPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "streetsandalleys.stalemate"
			} else {
				resObj.MessageCode = "streetsandalleys.playing"
			}
		case domain.StreetsAndAlleysPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "streetsandalleys.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.StreetsAndAlleysPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "streetsandalleys.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *StreetsAndAlleysWebPresenter) HintOutput(bc interfaces.StreetsAndAlleysGame) string {
	hint := bc.GetHint()
	resObj := new(controller.StreetsAndAlleysWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.StreetsAndAlleysWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.StreetsAndAlleysWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "streetsandalleys.hintAvailable"
	} else {
		resObj.MessageCode = "streetsandalleys.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *StreetsAndAlleysWebPresenter) ActionLogOutput(bc interfaces.StreetsAndAlleysGame) string {
	return actionLogOutputJSON(bc)
}
