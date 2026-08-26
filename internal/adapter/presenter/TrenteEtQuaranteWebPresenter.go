//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrenteEtQuaranteWebPresenter はトラント・エ・カラント (Trente et Quarante) の
// Web プレゼンタークラス。
type TrenteEtQuaranteWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *TrenteEtQuaranteWebPresenter) Output(g interfaces.TrenteEtQuaranteGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetPhase() == domain.TrenteEtQuarantePhaseResult {
		resObj.Message, resObj.MessageCode = trenteEtQuaranteEndMessage(g)
	}
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**TrenteEtQuarante.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TrenteEtQuaranteWebOutputHint{
			Bet:    int(hint.Bet),
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *TrenteEtQuaranteWebPresenter) buildBase(g interfaces.TrenteEtQuaranteGame) *controller.TrenteEtQuaranteWebOutput {
	resObj := new(controller.TrenteEtQuaranteWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.Chips = g.GetChips()
	resObj.CurrentBet = int(g.GetCurrentBet())
	resObj.Stake = g.GetStake()
	resObj.NoirRow = cardsToOutputOrEmpty(g.GetNoirRow())
	resObj.RougeRow = cardsToOutputOrEmpty(g.GetRougeRow())
	resObj.NoirTotal = g.GetNoirTotal()
	resObj.RougeTotal = g.GetRougeTotal()
	resObj.WinningRow = g.GetWinningRow()
	resObj.FirstCardRed = g.GetFirstCardRed()
	resObj.Refait = g.GetRefait()
	resObj.Result = trenteEtQuaranteWireResult(g)
	resObj.Payout = g.GetPayout()
	resObj.RemainingDeck = g.GetRemainingDeck()
	resObj.GameEndFlag = g.GetGameEndFlag()

	cfg := g.GetConfig()
	resObj.Config = controller.TrenteEtQuaranteWebConfigOutput{
		DefaultBet: int(cfg.DefaultBet),
	}
	return resObj
}

// trenteEtQuaranteWireResult は勝敗結果をフロントエンド契約のワイヤー値に変換する
// (1=win, -1=lose, 0=push/refait)。ドメインでは Refait を「半額負け」として Lose で
// 表現するが、ワイヤーでは push とともに 0 として送る (refait フラグで区別可能)。
func trenteEtQuaranteWireResult(g interfaces.TrenteEtQuaranteGame) int {
	if g.GetRefait() {
		return 0
	}
	return int(g.GetResult())
}

// trenteEtQuaranteEndMessage はラウンド終了時の表示メッセージと i18n キーを返す。
func trenteEtQuaranteEndMessage(g interfaces.TrenteEtQuaranteGame) (string, string) {
	if g.GetRefait() {
		return "", "trenteetquarante.result.refait"
	}
	switch g.GetResult() {
	case domain.TrenteEtQuaranteResultWin:
		return "", "trenteetquarante.result.win"
	case domain.TrenteEtQuaranteResultDraw:
		return "", "trenteetquarante.result.push"
	default:
		return "", "trenteetquarante.result.lose"
	}
}

// HintOutput はヒント情報を JSON 出力する。
func (p *TrenteEtQuaranteWebPresenter) HintOutput(g interfaces.TrenteEtQuaranteGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TrenteEtQuaranteWebOutputHint{
			Bet:    int(hint.Bet),
			Reason: hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *TrenteEtQuaranteWebPresenter) ActionLogOutput(g interfaces.TrenteEtQuaranteGame) string {
	return actionLogOutputJSON(g)
}
