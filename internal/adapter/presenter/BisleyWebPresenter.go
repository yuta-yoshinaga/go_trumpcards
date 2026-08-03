//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BisleyWebPresenter ビズリー Web プレゼンタークラス
type BisleyWebPresenter struct{}

// bisleyFoundationsToOutput converts one of the two foundation sets to wire format.
func bisleyFoundationsToOutput(piles [domain.BisleyFoundationCnt][]*domain.Card) [][]*controller.WebOutputCard {
	out := make([][]*controller.WebOutputCard, domain.BisleyFoundationCnt)
	for i := range domain.BisleyFoundationCnt {
		out[i] = make([]*controller.WebOutputCard, len(piles[i]))
		for j, c := range piles[i] {
			out[i][j] = cardToOutput(c)
		}
	}
	return out
}

// Output ゲーム状態をJSON出力
func (p *BisleyWebPresenter) Output(b interfaces.BisleyGame, lastErr error) string {
	resObj := new(controller.BisleyWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, b, int(b.GetPhase()))

	// タブロー — ビズリーは規則上すべて表向きだが、将来の伏せ札バリアントで状態が
	// 漏れないよう、ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := b.GetTableau()
	resObj.Tableau = make([][]*controller.BisleyWebOutputTableauCard, domain.BisleyTableauCnt)
	for i := range domain.BisleyTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.BisleyWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.BisleyWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 基礎札（昇順・降順）
	resObj.AceFoundations = bisleyFoundationsToOutput(b.GetAceFoundations())
	resObj.KingFoundations = bisleyFoundationsToOutput(b.GetKingFoundations())

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if b.GetPhase() == domain.BisleyPhasePlaying && !b.IsStalemate() {
		if hint := b.GetHint(); hint != nil {
			resObj.Hint = &controller.BisleyWebOutputHint{
				FromCol: hint.FromCol,
				ToZone:  hint.ToZone,
				ToIdx:   hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch b.GetPhase() {
		case domain.BisleyPhasePlaying:
			if b.IsStalemate() {
				resObj.MessageCode = "bisley.stalemate"
			} else {
				resObj.MessageCode = "bisley.playing"
			}
		case domain.BisleyPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", b.GetMoveCount())
			resObj.MessageCode = "bisley.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", b.GetMoveCount())}
		case domain.BisleyPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "bisley.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BisleyWebPresenter) HintOutput(b interfaces.BisleyGame) string {
	hint := b.GetHint()
	resObj := new(controller.BisleyWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, b, int(b.GetPhase()))
	resObj.Tableau = make([][]*controller.BisleyWebOutputTableauCard, 0)
	resObj.AceFoundations = make([][]*controller.WebOutputCard, 0)
	resObj.KingFoundations = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.BisleyWebOutputHint{
			FromCol: hint.FromCol,
			ToZone:  hint.ToZone,
			ToIdx:   hint.ToIdx,
		}
		resObj.MessageCode = "bisley.hintAvailable"
	} else {
		resObj.MessageCode = "bisley.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BisleyWebPresenter) ActionLogOutput(b interfaces.BisleyGame) string {
	return actionLogOutputJSON(b)
}
