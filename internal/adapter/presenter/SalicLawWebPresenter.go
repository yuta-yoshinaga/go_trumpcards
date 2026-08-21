//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SalicLawWebPresenter サリカ法典 Web プレゼンタークラス
type SalicLawWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SalicLawWebPresenter) Output(c interfaces.SalicLawGame, lastErr error) string {
	resObj := new(controller.SalicLawWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.SalicLawTableauCnt)
	for i := range domain.SalicLawTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.SalicLawFoundationCnt)
	for i := range domain.SalicLawFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = c.GetStockCount()
	resObj.OpenPiles = c.GetOpenPiles()
	queens := c.GetQueens()
	resObj.Queens = make([]*controller.WebOutputCard, len(queens))
	for i, card := range queens {
		resObj.Queens[i] = cardToOutput(card)
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.SalicLawPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.SalicLawWebOutputHint{
				FromZone: hint.FromZone,
				FromIdx:  hint.FromIdx,
				ToZone:   hint.ToZone,
				ToIdx:    hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		// コードを持つエラーはクライアントの i18n に組み立てさせる。ここで
		// Error() をそのまま入れると、キーを名乗るエラーはキー文字列が画面に
		// 出る (#5562)。
		if code, params := domain.ErrorMessageCode(lastErr); code != "" {
			resObj.MessageCode = code
			resObj.MessageParams = params
		} else {
			resObj.Message = lastErr.Error()
		}
	} else {
		switch c.GetPhase() {
		case domain.SalicLawPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "saliclaw.stalemate"
			} else {
				resObj.MessageCode = "saliclaw.playing"
			}
		case domain.SalicLawPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "saliclaw.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.SalicLawPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "saliclaw.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SalicLawWebPresenter) HintOutput(c interfaces.SalicLawGame) string {
	hint := c.GetHint()
	resObj := new(controller.SalicLawWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Queens = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SalicLawWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "saliclaw.hintAvailable"
	} else {
		resObj.MessageCode = "saliclaw.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SalicLawWebPresenter) ActionLogOutput(c interfaces.SalicLawGame) string {
	return actionLogOutputJSON(c)
}
