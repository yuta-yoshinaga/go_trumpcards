//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CongressWebPresenter コングレス Web プレゼンタークラス
type CongressWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CongressWebPresenter) Output(c interfaces.CongressGame, lastErr error) string {
	resObj := new(controller.CongressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.LegalTargets = p.buildLegalTargets(c)

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.CongressTableauCnt)
	for i := range domain.CongressTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CongressFoundationCnt)
	for i := range domain.CongressFoundationCnt {
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
	if c.GetPhase() == domain.CongressPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.CongressWebOutputHint{
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
		case domain.CongressPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "congress.stalemate"
			} else {
				resObj.MessageCode = "congress.playing"
			}
		case domain.CongressPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "congress.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CongressPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "congress.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CongressWebPresenter) HintOutput(c interfaces.CongressGame) string {
	hint := c.GetHint()
	resObj := new(controller.CongressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.LegalTargets = p.buildLegalTargets(c)

	if hint != nil {
		resObj.Hint = &controller.CongressWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "congress.hintAvailable"
	} else {
		resObj.MessageCode = "congress.noHint"
	}

	return marshalOrError(resObj)
}

func (p *CongressWebPresenter) buildLegalTargets(c interfaces.CongressGame) map[string]controller.CongressWebOutputTargets {
	result := make(map[string]controller.CongressWebOutputTargets)
	for pile := 0; pile < domain.CongressTableauCnt; pile++ {
		tableau, foundation := c.LegalTargets(pile)
		if len(tableau)+len(foundation) == 0 {
			continue
		}
		if tableau == nil {
			tableau = []int{}
		}
		if foundation == nil {
			foundation = []int{}
		}
		result["tableau-"+strconv.Itoa(pile)] = controller.CongressWebOutputTargets{Tableau: tableau, Foundation: foundation}
	}
	tableau, foundation := c.WasteLegalTargets()
	if len(tableau)+len(foundation) > 0 {
		if tableau == nil {
			tableau = []int{}
		}
		if foundation == nil {
			foundation = []int{}
		}
		result["waste"] = controller.CongressWebOutputTargets{Tableau: tableau, Foundation: foundation}
	}
	tableau = c.StockLegalTargets()
	if len(tableau) > 0 {
		result["stock"] = controller.CongressWebOutputTargets{Tableau: tableau, Foundation: []int{}}
	}
	return result
}

// ActionLogOutput 棋譜をJSON出力
func (p *CongressWebPresenter) ActionLogOutput(c interfaces.CongressGame) string {
	return actionLogToJSON(c.GetActionLog())
}
