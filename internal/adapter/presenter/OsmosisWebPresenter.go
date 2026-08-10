//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OsmosisWebPresenter オズモシスWebプレゼンタークラス
type OsmosisWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OsmosisWebPresenter) Output(o interfaces.OsmosisGame, lastErr error) string {
	resObj := p.buildBaseOutput(o)
	// 1回だけ数えて使い回す。3箇所で呼ぶと、将来 IsStalemate に副作用が入ったとき
	// 判定同士が食い違いうる。
	stalemate := resObj.IsStalemate

	// ウェイスト
	waste := o.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, wc := range waste {
		resObj.Waste[i] = cardToOutput(wc)
	}

	// リザーブ（4列）
	reserve := o.GetReserve()
	resObj.Reserve = make([][]*controller.WebOutputCard, domain.OsmosisReserveCnt)
	for i := 0; i < domain.OsmosisReserveCnt; i++ {
		pile := reserve[i]
		resObj.Reserve[i] = make([]*controller.WebOutputCard, len(pile))
		for j, rc := range pile {
			resObj.Reserve[i][j] = cardToOutput(rc)
		}
	}

	// ファンデーション（4段）
	foundation := o.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.OsmosisFoundationCnt)
	for i := 0; i < domain.OsmosisFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Foundation[i][j] = cardToOutput(fc)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// 手詰まりならもう置ける札は無いので、ヒントを探しに行くだけ無駄になる。
	if o.GetPhase() == domain.OsmosisPhasePlaying && !stalemate {
		if hint := o.GetHint(); hint != nil {
			resObj.Hint = &controller.OsmosisWebOutputHint{
				FromZone: hint.FromZone,
				FromCol:  hint.FromCol,
				ToCol:    hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch o.GetPhase() {
		case domain.OsmosisPhasePlaying:
			if stalemate {
				resObj.MessageCode = "osmosis.stalemate"
			} else {
				resObj.MessageCode = "osmosis.playing"
			}
		case domain.OsmosisPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", o.GetMoveCount())
			resObj.MessageCode = "osmosis.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", o.GetMoveCount())}
		case domain.OsmosisPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "osmosis.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *OsmosisWebPresenter) HintOutput(o interfaces.OsmosisGame) string {
	resObj := p.buildBaseOutput(o)
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Reserve = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	hint := o.GetHint()
	if hint != nil {
		resObj.Hint = &controller.OsmosisWebOutputHint{
			FromZone: hint.FromZone,
			FromCol:  hint.FromCol,
			ToCol:    hint.ToCol,
		}
		resObj.MessageCode = "osmosis.hintAvailable"
	} else {
		resObj.MessageCode = "osmosis.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OsmosisWebPresenter) ActionLogOutput(o interfaces.OsmosisGame) string {
	return actionLogOutputJSON(o)
}

func (p *OsmosisWebPresenter) buildBaseOutput(o interfaces.OsmosisGame) *controller.OsmosisWebOutput {
	return &controller.OsmosisWebOutput{
		BaseRank:    o.GetBaseRank(),
		Phase:       int(o.GetPhase()),
		MoveCount:   o.GetMoveCount(),
		StockCount:  o.GetStockCount(),
		CanUndo:     o.CanUndo(),
		IsStalemate: o.IsStalemate(),
	}
}
