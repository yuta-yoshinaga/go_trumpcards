//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BakersDozenWebPresenter ベーカーズダズンWebプレゼンタークラス
type BakersDozenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BakersDozenWebPresenter) Output(bd interfaces.BakersDozenGame, lastErr error) string {
	resObj := new(controller.BakersDozenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bd, int(bd.GetPhase()))

	// タブロー — every card in Baker's Dozen is face-up by rule, so we always
	// serialise tc.Card and the FaceUp field is effectively a constant true.
	tableau := bd.GetTableau()
	resObj.Tableau = make([][]*controller.BakersDozenWebOutputTableauCard, domain.BakersDozenTableauCnt)
	for i := range domain.BakersDozenTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.BakersDozenWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.BakersDozenWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: true,
			}
		}
	}
	resObj.LegalTargets = p.buildLegalTargets(bd)

	// ファンデーション
	foundation := bd.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.BakersDozenFoundationCnt)
	for i := range domain.BakersDozenFoundationCnt {
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
	if bd.GetPhase() == domain.BakersDozenPhasePlaying && !bd.IsStalemate() {
		if hint := bd.GetHint(); hint != nil {
			resObj.Hint = &controller.BakersDozenWebOutputHint{
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
		phase := bd.GetPhase()
		switch phase {
		case domain.BakersDozenPhasePlaying:
			if bd.IsStalemate() {
				resObj.MessageCode = "bakersdozen.stalemate"
			} else {
				resObj.MessageCode = "bakersdozen.playing"
			}
		case domain.BakersDozenPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bd.GetMoveCount())
			resObj.MessageCode = "bakersdozen.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bd.GetMoveCount())}
		case domain.BakersDozenPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "bakersdozen.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// buildLegalTargets returns the domain's legal destinations for every source column.
func (p *BakersDozenWebPresenter) buildLegalTargets(g interfaces.BakersDozenGame) map[string]controller.BakersDozenWebOutputTargets {
	out := make(map[string]controller.BakersDozenWebOutputTargets)
	tableauState := g.GetTableau()
	for col := 0; col < domain.BakersDozenTableauCnt; col++ {
		if len(tableauState[col]) == 0 {
			continue
		}
		tableau, foundation := g.LegalTargets(col)
		if len(tableau) == 0 && len(foundation) == 0 {
			continue
		}
		if tableau == nil {
			tableau = []int{}
		}
		if foundation == nil {
			foundation = []int{}
		}
		out["tableau-"+strconv.Itoa(col)] = controller.BakersDozenWebOutputTargets{Tableau: tableau, Foundation: foundation}
	}
	return out
}

// HintOutput ヒントをJSON出力
func (p *BakersDozenWebPresenter) HintOutput(bd interfaces.BakersDozenGame) string {
	hint := bd.GetHint()
	resObj := new(controller.BakersDozenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bd, int(bd.GetPhase()))
	resObj.Tableau = make([][]*controller.BakersDozenWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.BakersDozenWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "bakersdozen.hintAvailable"
	} else {
		resObj.MessageCode = "bakersdozen.noHint"
	}

	return marshalOrError(resObj)
}

// TargetsOutput は Web では通常の盤面をそのまま返す。
//
// 合法な移動先は通常の盤面レスポンスの `legalTargets` で返す。
// Web には CUI の `targets` に当たる操作はない (#5581)。
// ここで別の形を返すと、CUI 専用の応答が Web の経路に紛れ込む。
func (p *BakersDozenWebPresenter) TargetsOutput(bd interfaces.BakersDozenGame, _ int) string {
	return p.Output(bd, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BakersDozenWebPresenter) ActionLogOutput(bd interfaces.BakersDozenGame) string {
	return actionLogOutputJSON(bd)
}
