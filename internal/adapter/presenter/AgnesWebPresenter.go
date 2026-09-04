//go:build !js || !wasm || extra5

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AgnesWebPresenter アグネス・ソレルWebプレゼンタークラス
type AgnesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AgnesWebPresenter) Output(c interfaces.AgnesGame, lastErr error) string {
	resObj := p.buildBaseOutput(c)
	p.fillBoard(resObj, c)

	// メッセージ
	// **受動ヒントは Output() でも埋める。**盤面のリングなど `state.hint` を読む
	// 分岐は、ヒントボタンを押していないときにも動く。ここで埋めないとそれらは
	// 全部死ぬ (#4483)。
	//
	// **「HintOutput の応答はページの state にマージされない」と書いてあったが、
	// このページでは誤り。**ヒントボタンは `useGameApi` の `exec('hint')` を呼び、
	// `setState(res)` が状態を丸ごと差し替える。だから HintOutput も
	// `fillBoard` で盤面を返す (#6800 / #6855)。
	if c.GetPhase() == domain.AgnesPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.AgnesWebOutputHint{
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
		switch c.GetPhase() {
		case domain.AgnesPhasePlaying:
			resObj.MessageCode = "agnes.playing"
		case domain.AgnesPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "agnes.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.AgnesPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "agnes.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *AgnesWebPresenter) HintOutput(c interfaces.AgnesGame) string {
	resObj := p.buildBaseOutput(c)
	p.fillBoard(resObj, c)

	hint := c.GetHint()
	if hint != nil {
		resObj.Hint = &controller.AgnesWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "agnes.hintAvailable"
	} else {
		resObj.MessageCode = "agnes.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AgnesWebPresenter) ActionLogOutput(c interfaces.AgnesGame) string {
	return actionLogOutputJSON(c)
}

// fillBoard は盤面を埋める。
//
// **Output と HintOutput が同じ盤面を返すためにある。**HintOutput は以前
// 盤面を空配列で潰していた ── `buildBaseOutput` は盤面を埋めないので、
// 潰さないと JSON に `null` が出てフロントの `.map` が壊れる、という理由だった。
// だがこのページのヒントボタン (と CLI の hint) は `useGameApi` の `exec` を
// 呼び、`useGameApi` は `setState(res)` で状態を**丸ごと差し替える**ので、
// その空配列がそのまま画面に流れ込んで盤面が消えていた (#6855、#6800 と同型)。
func (p *AgnesWebPresenter) fillBoard(resObj *controller.AgnesWebOutput, c interfaces.AgnesGame) {
	// タブロー
	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.AgnesWebOutputTableauCard, domain.AgnesTableauCnt)
	for i := 0; i < domain.AgnesTableauCnt; i++ {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.AgnesWebOutputTableauCard, len(col))
		for j, tc := range col {
			var card *controller.WebOutputCard
			if tc.FaceUp {
				card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = &controller.AgnesWebOutputTableauCard{Card: card, FaceUp: tc.FaceUp}
		}
	}

	// ファンデーション
	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.AgnesFoundationCnt)
	for i := 0; i < domain.AgnesFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Foundation[i][j] = cardToOutput(fc)
		}
	}
}

func (p *AgnesWebPresenter) buildBaseOutput(c interfaces.AgnesGame) *controller.AgnesWebOutput {
	return &controller.AgnesWebOutput{
		BaseRank:    c.GetBaseRank(),
		Phase:       int(c.GetPhase()),
		MoveCount:   c.GetMoveCount(),
		StockCount:  c.GetStockCount(),
		CanUndo:     c.CanUndo(),
		IsStalemate: c.IsStalemate(),
	}
}
