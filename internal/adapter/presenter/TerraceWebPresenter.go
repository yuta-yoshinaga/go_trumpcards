//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TerraceWebPresenter テラス Web プレゼンタークラス
type TerraceWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TerraceWebPresenter) Output(t interfaces.TerraceGame, lastErr error) string {
	resObj := new(controller.TerraceWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, t, int(t.GetPhase()))

	reserve := t.GetReserve()
	resObj.Reserve = make([]*controller.WebOutputCard, len(reserve))
	for i, card := range reserve {
		resObj.Reserve[i] = cardToOutput(card)
	}

	tableau := t.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.TerraceTableauCnt)
	for i := range domain.TerraceTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := t.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.TerraceFoundationCnt)
	for i := range domain.TerraceFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = t.GetStockCount()
	waste := t.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}
	// baseRank と awaitingBaseRank を両方渡すのは、クライアントに「0 は未決定」
	// という約束を再実装させないため。
	resObj.BaseRank = t.GetBaseRank()
	resObj.AwaitingBaseRank = t.IsAwaitingBaseRank()

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// 開始ランク待ちと手詰まりでは、フロントが別経路 (AwaitingBaseRank /
	// IsStalemate) で文言を出すので、ここでヒントを重ねる意味が無い。
	if t.GetPhase() == domain.TerracePhasePlaying && !t.IsAwaitingBaseRank() && !t.IsStalemate() {
		if hint := t.GetHint(); hint != nil {
			resObj.Hint = &controller.TerraceWebOutputHint{
				FromZone: hint.FromZone,
				FromIdx:  hint.FromIdx,
				ToZone:   hint.ToZone,
				ToIdx:    hint.ToIdx,
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
		switch t.GetPhase() {
		case domain.TerracePhasePlaying:
			switch {
			case t.IsAwaitingBaseRank():
				resObj.MessageCode = "terrace.chooseBase"
			case t.IsStalemate():
				resObj.MessageCode = "terrace.stalemate"
			default:
				resObj.MessageCode = "terrace.playing"
			}
		case domain.TerracePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", t.GetMoveCount())
			resObj.MessageCode = "terrace.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", t.GetMoveCount())}
		case domain.TerracePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "terrace.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *TerraceWebPresenter) HintOutput(t interfaces.TerraceGame) string {
	hint := t.GetHint()
	resObj := new(controller.TerraceWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, t, int(t.GetPhase()))
	resObj.Reserve = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.TerraceWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "terrace.hintAvailable"
	} else {
		resObj.MessageCode = "terrace.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TerraceWebPresenter) ActionLogOutput(t interfaces.TerraceGame) string {
	return actionLogOutputJSON(t)
}
