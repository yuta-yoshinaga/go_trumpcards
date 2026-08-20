//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MissMilliganWebPresenter ミス・ミリガン Web プレゼンタークラス
type MissMilliganWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MissMilliganWebPresenter) Output(mm interfaces.MissMilliganGame, lastErr error) string {
	resObj := new(controller.MissMilliganWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, mm, int(mm.GetPhase()))

	// タブロー — 規則上すべて表向きだが、将来の伏せ札バリアントで状態が漏れない
	// よう、ハードコードせずドメインの FaceUp をそのまま流す。
	tableau := mm.GetTableau()
	resObj.Tableau = make([][]*controller.MissMilliganWebOutputTableauCard, domain.MissMilliganTableauCnt)
	for i := range domain.MissMilliganTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.MissMilliganWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.MissMilliganWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// 基礎札
	foundation := mm.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.MissMilliganFoundationCnt)
	for i := range domain.MissMilliganFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// 山札と保持中の札。canWaive はクライアントが「山札が空 かつ 未保持」を
	// 再計算しなくて済むよう、ドメインの判断をそのまま渡す。
	resObj.StockCount = mm.GetStockCount()
	waived := mm.GetWaived()
	resObj.Waived = make([]*controller.WebOutputCard, len(waived))
	for i, c := range waived {
		resObj.Waived[i] = cardToOutput(c)
	}
	resObj.CanWaive = mm.CanWaive()

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if mm.GetPhase() == domain.MissMilliganPhasePlaying && !mm.IsStalemate() {
		if hint := mm.GetHint(); hint != nil {
			resObj.Hint = &controller.MissMilliganWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToIdx:     hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		// コードを持つエラーはクライアントの i18n に組み立てさせる。以前は
		// lastErr.Error() をそのまま Message に入れており、ドメインが英語で
		// 書いた文がロケールに関係なく出ていた (#5556)。
		if code, params := domain.ErrorMessageCode(lastErr); code != "" {
			resObj.MessageCode = code
			resObj.MessageParams = params
		} else {
			resObj.Message = lastErr.Error()
		}
	} else {
		switch mm.GetPhase() {
		case domain.MissMilliganPhasePlaying:
			switch {
			case mm.IsStalemate():
				resObj.MessageCode = "missmilligan.stalemate"
			case len(waived) > 0:
				resObj.MessageCode = "missmilligan.waiving"
			default:
				resObj.MessageCode = "missmilligan.playing"
			}
		case domain.MissMilliganPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", mm.GetMoveCount())
			resObj.MessageCode = "missmilligan.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", mm.GetMoveCount())}
		case domain.MissMilliganPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "missmilligan.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *MissMilliganWebPresenter) HintOutput(mm interfaces.MissMilliganGame) string {
	hint := mm.GetHint()
	resObj := new(controller.MissMilliganWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, mm, int(mm.GetPhase()))
	resObj.Tableau = make([][]*controller.MissMilliganWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waived = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.MissMilliganWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToIdx:     hint.ToIdx,
		}
		resObj.MessageCode = "missmilligan.hintAvailable"
	} else {
		resObj.MessageCode = "missmilligan.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MissMilliganWebPresenter) ActionLogOutput(mm interfaces.MissMilliganGame) string {
	return actionLogOutputJSON(mm)
}
