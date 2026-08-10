//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FourSeasonsWebPresenter フォーシーズンズWebプレゼンタークラス
type FourSeasonsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FourSeasonsWebPresenter) Output(f interfaces.FourSeasonsGame, lastErr error) string {
	resObj := p.buildBaseOutput(f)

	waste := f.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, w := range waste {
		resObj.Waste[i] = cardToOutput(w)
	}

	tableau := f.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.FourSeasonsTableauCnt)
	for i := range domain.FourSeasonsTableauCnt {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(col))
		for j, c := range col {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	foundation := f.GetFoundations()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FourSeasonsFoundationCnt)
	for i := range domain.FourSeasonsFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if f.GetPhase() == domain.FourSeasonsPhasePlaying {
		if hint := f.GetHint(); hint != nil {
			resObj.Hint = toFourSeasonsHintOutput(hint)
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch f.GetPhase() {
		case domain.FourSeasonsPhasePlaying:
			resObj.MessageCode = "fourseasons.playing"
		case domain.FourSeasonsPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", f.GetMoveCount())
			resObj.MessageCode = "fourseasons.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", f.GetMoveCount())}
		case domain.FourSeasonsPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "fourseasons.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FourSeasonsWebPresenter) HintOutput(f interfaces.FourSeasonsGame) string {
	resObj := p.buildBaseOutput(f)
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint := f.GetHint(); hint != nil {
		resObj.Hint = toFourSeasonsHintOutput(hint)
		resObj.MessageCode = "fourseasons.hintAvailable"
	} else {
		resObj.MessageCode = "fourseasons.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FourSeasonsWebPresenter) ActionLogOutput(f interfaces.FourSeasonsGame) string {
	return actionLogOutputJSON(f)
}

func toFourSeasonsHintOutput(h *domain.FourSeasonsHint) *controller.FourSeasonsWebOutputHint {
	return &controller.FourSeasonsWebOutputHint{
		FromZone: h.FromZone,
		FromIdx:  h.FromIdx,
		ToZone:   h.ToZone,
		ToIdx:    h.ToIdx,
	}
}

func (p *FourSeasonsWebPresenter) buildBaseOutput(f interfaces.FourSeasonsGame) *controller.FourSeasonsWebOutput {
	return &controller.FourSeasonsWebOutput{
		BaseRank:   f.GetBaseRank(),
		Phase:      int(f.GetPhase()),
		MoveCount:  f.GetMoveCount(),
		StockCount: f.GetStockCount(),
		CanUndo:    f.CanUndo(),
	}
}
