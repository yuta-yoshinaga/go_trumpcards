//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BristolWebPresenter ブリストルWebプレゼンタークラス
type BristolWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BristolWebPresenter) Output(b interfaces.BristolGame, lastErr error) string {
	resObj := p.buildBaseOutput(b)
	p.fillBoard(resObj, b)

	// メッセージ
	// **受動ヒントは Output() でも埋める。**盤面のリングなど `state.hint` を読む
	// 分岐は、ヒントボタンを押していないときにも動く。ここで埋めないとそれらは
	// 全部死ぬ (#4483)。
	//
	// **「HintOutput の応答はページの state にマージされない」と書いてあったが、
	// このページでは誤り。**ヒントボタンは `useGameApi` の `exec('hint')` を呼び、
	// `setState(res)` が状態を丸ごと差し替える。だから HintOutput も
	// `fillBoard` で盤面を返す (#6800 / #6855)。
	if b.GetPhase() == domain.BristolPhasePlaying {
		if hint := b.GetHint(); hint != nil {
			resObj.Hint = &controller.BristolWebOutputHint{
				FromZone: hint.FromZone,
				FromCol:  hint.FromCol,
				ToZone:   hint.ToZone,
				ToCol:    hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch b.GetPhase() {
		case domain.BristolPhasePlaying:
			resObj.MessageCode = "bristol.playing"
		case domain.BristolPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", b.GetMoveCount())
			resObj.MessageCode = "bristol.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", b.GetMoveCount())}
		case domain.BristolPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "bristol.gameOver"
		}
	}

	resObj.LegalTargets = p.buildLegalTargets(b)

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BristolWebPresenter) HintOutput(b interfaces.BristolGame) string {
	resObj := p.buildBaseOutput(b)
	p.fillBoard(resObj, b)

	hint := b.GetHint()
	if hint != nil {
		resObj.Hint = &controller.BristolWebOutputHint{
			FromZone: hint.FromZone,
			FromCol:  hint.FromCol,
			ToZone:   hint.ToZone,
			ToCol:    hint.ToCol,
		}
		resObj.MessageCode = "bristol.hintAvailable"
	} else {
		resObj.MessageCode = "bristol.noHint"
	}
	return marshalOrError(resObj)
}

// TargetsOutput は Web では通常の盤面をそのまま返す。
//
// 置ける先の強調は `buildLegalTargets` がこの盤面から作っている ──
// Web には `targets` に当たる操作が無く、選択した瞬間に見えている (#6427)。
// ここで別の形を返すと、CUI 専用の応答が Web の経路に紛れ込む。
func (p *BristolWebPresenter) TargetsOutput(b interfaces.BristolGame, _ string, _ int) string {
	return p.Output(b, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BristolWebPresenter) ActionLogOutput(b interfaces.BristolGame) string {
	return actionLogOutputJSON(b)
}

// fillBoard は盤面を埋める。
//
// **Output と HintOutput が同じ盤面を返すためにある。**HintOutput は以前
// 盤面を空配列で潰していた ── `buildBaseOutput` は盤面を埋めないので、
// 潰さないと JSON に `null` が出てフロントの `.map` が壊れる、という理由だった。
// だがこのページのヒントボタン (と CLI の hint) は `useGameApi` の `exec` を
// 呼び、`useGameApi` は `setState(res)` で状態を**丸ごと差し替える**ので、
// その空配列がそのまま画面に流れ込んで盤面が消えていた (#6855、#6800 と同型)。
func (p *BristolWebPresenter) fillBoard(resObj *controller.BristolWebOutput, b interfaces.BristolGame) {
	// タブロー（8列）
	tableau := b.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.BristolTableauCnt)
	for i := 0; i < domain.BristolTableauCnt; i++ {
		col := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(col))
		for j, tc := range col {
			resObj.Tableau[i][j] = cardToOutput(tc)
		}
	}

	// ファン（3つ）
	fan := b.GetFan()
	resObj.Fan = make([][]*controller.WebOutputCard, domain.BristolFanCnt)
	for i := 0; i < domain.BristolFanCnt; i++ {
		pile := fan[i]
		resObj.Fan[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Fan[i][j] = cardToOutput(fc)
		}
	}

	// ファウンデーション（4つ）
	foundation := b.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.BristolFoundationCnt)
	for i := 0; i < domain.BristolFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, fc := range pile {
			resObj.Foundation[i][j] = cardToOutput(fc)
		}
	}
}

func (p *BristolWebPresenter) buildBaseOutput(b interfaces.BristolGame) *controller.BristolWebOutput {
	return &controller.BristolWebOutput{
		Phase:      int(b.GetPhase()),
		MoveCount:  b.GetMoveCount(),
		StockCount: b.GetStockCount(),
		CanUndo:    b.CanUndo(),
		// 手詰まりは画面に出さないと、動かせる札を探し続けることになる (#5631)。
		IsStalemate:  b.IsStalemate(),
		UndoToEscape: b.UndoToEscape(),
	}
}

// buildLegalTargets は移動元ごとの合法な移動先を返す。キーは "tableau-0" / "fan-2"。
//
// **押すまで合法か分からなかった。**選択中は全ての移動先が同じ見た目で強調されて
// いて、不正な移動はサーバーが弾くまで分からなかった (#4813)。判定はドメインの
// canPlaceOnTableau / canPlaceOnFoundation をそのまま通す。
func (p *BristolWebPresenter) buildLegalTargets(g interfaces.BristolGame) map[string]controller.BristolWebOutputTargets {
	out := make(map[string]controller.BristolWebOutputTargets)
	add := func(zone string, col int) {
		tab, found := g.LegalTargets(zone, col)
		if len(tab) == 0 && len(found) == 0 {
			return
		}
		if tab == nil {
			tab = []int{}
		}
		if found == nil {
			found = []int{}
		}
		out[zone+"-"+strconv.Itoa(col)] = controller.BristolWebOutputTargets{Tableau: tab, Foundation: found}
	}
	for col := 0; col < domain.BristolTableauCnt; col++ {
		add("tableau", col)
	}
	for col := 0; col < domain.BristolFanCnt; col++ {
		add("fan", col)
	}
	return out
}
