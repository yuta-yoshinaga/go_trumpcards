//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RoyalCotillionWebPresenter ロイヤルコティヨン Web プレゼンタークラス
type RoyalCotillionWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *RoyalCotillionWebPresenter) Output(c interfaces.RoyalCotillionGame, lastErr error) string {
	resObj := new(controller.RoyalCotillionWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	// タブローは 1 枠 1 枚。空き枠は null で送る。
	tableau := c.GetTableau()
	resObj.Tableau = make([]*controller.WebOutputCard, domain.RoyalCotillionTableauCnt)
	for i := range domain.RoyalCotillionTableauCnt {
		resObj.Tableau[i] = cardToOutput(tableau[i])
	}

	reserve := c.GetReserve()
	resObj.Reserve = make([][]*controller.WebOutputCard, domain.RoyalCotillionReserveCnt)
	for i := range domain.RoyalCotillionReserveCnt {
		pile := reserve[i]
		resObj.Reserve[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Reserve[i][j] = cardToOutput(card)
		}
	}

	// A 始まり / 2 始まりはワイヤに載せる。添字から推測させると、並びを変えた
	// ときに表示だけが静かにずれる。
	resObj.FoundationOdd = make([]bool, domain.RoyalCotillionFoundationCnt)
	for i := range domain.RoyalCotillionFoundationCnt {
		resObj.FoundationOdd[i] = c.IsOddFoundation(i)
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.RoyalCotillionFoundationCnt)
	for i := range domain.RoyalCotillionFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = c.GetStockCount()
	waste := c.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.RoyalCotillionPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.RoyalCotillionWebOutputHint{
				FromZone: hint.FromZone,
				FromIdx:  hint.FromIdx,
				ToZone:   hint.ToZone,
				ToIdx:    hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch c.GetPhase() {
		case domain.RoyalCotillionPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "royalcotillion.stalemate"
			} else {
				resObj.MessageCode = "royalcotillion.playing"
			}
		case domain.RoyalCotillionPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "royalcotillion.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.RoyalCotillionPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "royalcotillion.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *RoyalCotillionWebPresenter) HintOutput(c interfaces.RoyalCotillionGame) string {
	hint := c.GetHint()
	resObj := new(controller.RoyalCotillionWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([]*controller.WebOutputCard, 0)
	resObj.Reserve = make([][]*controller.WebOutputCard, 0)
	resObj.FoundationOdd = make([]bool, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.RoyalCotillionWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "royalcotillion.hintAvailable"
	} else {
		resObj.MessageCode = "royalcotillion.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *RoyalCotillionWebPresenter) ActionLogOutput(c interfaces.RoyalCotillionGame) string {
	return actionLogOutputJSON(c)
}
