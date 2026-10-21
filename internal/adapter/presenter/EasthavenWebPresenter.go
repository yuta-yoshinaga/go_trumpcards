//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EasthavenWebPresenter イーストヘイブンWebプレゼンタークラス
type EasthavenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *EasthavenWebPresenter) Output(e interfaces.EasthavenGame, lastErr error) string {
	resObj := new(controller.EasthavenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, e, int(e.GetPhase()))
	resObj.StockCount = e.GetStockCount()
	resObj.LegalTargets = p.buildLegalTargets(e)

	// タブロー
	tableau := e.GetTableau()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.EasthavenTableauCnt)
	for i := range domain.EasthavenTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.KlondikeWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.KlondikeWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := e.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.EasthavenFoundationCnt)
	for i := range domain.EasthavenFoundationCnt {
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
	if e.GetPhase() == domain.EasthavenPhasePlaying && !e.IsStalemate() {
		if hint := e.GetHint(); hint != nil {
			resObj.Hint = &controller.EasthavenWebOutputHint{
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
		switch e.GetPhase() {
		case domain.EasthavenPhasePlaying:
			if e.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "easthaven.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "easthaven.stalemate"
				}
			} else {
				resObj.MessageCode = "easthaven.playing"
			}
		case domain.EasthavenPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", e.GetMoveCount())
			resObj.MessageCode = "easthaven.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", e.GetMoveCount())}
		case domain.EasthavenPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "easthaven.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *EasthavenWebPresenter) HintOutput(e interfaces.EasthavenGame) string {
	hint := e.GetHint()
	resObj := new(controller.EasthavenWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, e, int(e.GetPhase()))
	resObj.StockCount = e.GetStockCount()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.LegalTargets = p.buildLegalTargets(e)

	if hint != nil {
		resObj.Hint = &controller.EasthavenWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "easthaven.hintAvailable"
	} else {
		resObj.MessageCode = "easthaven.noHint"
	}

	return marshalOrError(resObj)
}

func (p *EasthavenWebPresenter) buildLegalTargets(e interfaces.EasthavenGame) map[string]controller.EasthavenWebOutputTargets {
	result := make(map[string]controller.EasthavenWebOutputTargets)
	tableau := e.GetTableau()
	for col := range domain.EasthavenTableauCnt {
		for index, card := range tableau[col] {
			if card == nil || !card.FaceUp {
				continue
			}
			tab, foundation := e.LegalTargetsFrom(col, index)
			if len(tab)+len(foundation) == 0 {
				continue
			}
			if tab == nil {
				tab = []int{}
			}
			if foundation == nil {
				foundation = []int{}
			}
			result["tableau-"+strconv.Itoa(col)+"-"+strconv.Itoa(index)] = controller.EasthavenWebOutputTargets{Tableau: tab, Foundation: foundation}
		}
	}
	return result
}

// ActionLogOutput 棋譜をJSON出力
func (p *EasthavenWebPresenter) ActionLogOutput(e interfaces.EasthavenGame) string {
	return actionLogOutputJSON(e)
}
