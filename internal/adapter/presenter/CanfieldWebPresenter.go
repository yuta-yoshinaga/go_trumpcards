//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CanfieldWebPresenter キャンフィールドWebプレゼンタークラス
type CanfieldWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CanfieldWebPresenter) Output(c interfaces.CanfieldGame, lastErr error) string {
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
	if c.GetPhase() == domain.CanfieldPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.CanfieldWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
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
		switch c.GetPhase() {
		case domain.CanfieldPhasePlaying:
			resObj.MessageCode = "canfield.playing"
		case domain.CanfieldPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "canfield.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CanfieldPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "canfield.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CanfieldWebPresenter) HintOutput(c interfaces.CanfieldGame) string {
	resObj := p.buildBaseOutput(c)
	p.fillBoard(resObj, c)

	hint := c.GetHint()
	if hint != nil {
		resObj.Hint = &controller.CanfieldWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "canfield.hintAvailable"
	} else {
		resObj.MessageCode = "canfield.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CanfieldWebPresenter) ActionLogOutput(c interfaces.CanfieldGame) string {
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
func (p *CanfieldWebPresenter) fillBoard(resObj *controller.CanfieldWebOutput, c interfaces.CanfieldGame) {
	// ウェイスト
	waste := c.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, w := range waste {
			resObj.Waste[i] = cardToOutput(w)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// リザーブ
	reserve := c.GetReserve()
	resObj.Reserve = make([]*controller.WebOutputCard, len(reserve))
	for i, rc := range reserve {
		resObj.Reserve[i] = cardToOutput(rc)
	}

	// タブロー
	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.CanfieldWebOutputTableauCard, domain.CanfieldTableauCnt)
	for i := 0; i < domain.CanfieldTableauCnt; i++ {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.CanfieldWebOutputTableauCard, len(col))
		for j, tc := range col {
			resObj.Tableau[i][j] = &controller.CanfieldWebOutputTableauCard{Card: cardToOutput(tc.Card)}
		}
	}

	// ファンデーション
	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CanfieldFoundationCnt)
	for i := 0; i < domain.CanfieldFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Foundation[i][j] = cardToOutput(fc)
		}
	}
}

func (p *CanfieldWebPresenter) buildBaseOutput(c interfaces.CanfieldGame) *controller.CanfieldWebOutput {
	return &controller.CanfieldWebOutput{
		BaseRank:   c.GetBaseRank(),
		Phase:      int(c.GetPhase()),
		MoveCount:  c.GetMoveCount(),
		StockCount: c.GetStockCount(),
		CanUndo:    c.CanUndo(),
	}
}
