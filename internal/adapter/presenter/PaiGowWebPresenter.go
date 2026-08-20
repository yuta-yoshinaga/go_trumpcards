//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PaiGowWebPresenter パイガオポーカーWebプレゼンタークラス
type PaiGowWebPresenter struct {
}

// Output ゲーム状態を出力
func (pp *PaiGowWebPresenter) Output(pg interfaces.PaiGowGame, lastErr error) string {
	resObj := new(controller.PaiGowWebOutput)

	resObj.PlayerCards = cardsToOutputOrEmpty(pg.GetPlayerCards())
	resObj.DealerCards = cardsToOutputOrEmpty(pg.GetDealerCards())
	resObj.PlayerHighHand = cardsToOutputOrEmpty(pg.GetPlayerHighHand())
	resObj.PlayerLowHand = cardsToOutputOrEmpty(pg.GetPlayerLowHand())
	resObj.DealerHighHand = cardsToOutputOrEmpty(pg.GetDealerHighHand())
	resObj.DealerLowHand = cardsToOutputOrEmpty(pg.GetDealerLowHand())
	resObj.Phase = pg.GetPhase()
	resObj.Chips = pg.GetChips()
	resObj.Bet = pg.GetBet()
	resObj.Result = int(pg.GetResult())
	resObj.HighHandResult = int(pg.GetHighHandResult())
	resObj.LowHandResult = int(pg.GetLowHandResult())
	resObj.Payout = pg.GetPayout()
	resObj.Commission = pg.GetCommission()
	resObj.PlayerHighRank = pg.GetPlayerHighRank()
	resObj.PlayerLowRank = pg.GetPlayerLowRank()
	resObj.DealerHighRank = pg.GetDealerHighRank()
	resObj.DealerLowRank = pg.GetDealerLowRank()

	// **受動ヒントは Output() でも埋める。**HintOutput() は command:"hint" 専用の
	// レスポンスで、ページの state にはマージされない。フェーズ判定は
	// PaiGow.GetHint() 側が持つ。
	resObj.Hint = paiGowWebHint(pg)

	if lastErr != nil {
		// コードを持つエラーはクライアントの i18n に組み立てさせる。ここで
		// Error() をそのまま入れると、キーを名乗るエラーはキー文字列が
		// 画面に出る (#5526)。
		if code, params := domain.ErrorMessageCode(lastErr); code != "" {
			resObj.MessageCode = code
			resObj.MessageParams = params
		} else {
			resObj.Message = lastErr.Error()
		}
	} else if pg.GetGameEndFlag() {
		switch pg.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "paigow.result.playerWins"
		case domain.GameResultLose:
			resObj.Message = "Dealer wins!"
			resObj.MessageCode = "paigow.result.dealerWins"
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "paigow.result.push"
		default:
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒント情報をJSON出力する
func (pp *PaiGowWebPresenter) HintOutput(pg interfaces.PaiGowGame) string {
	resObj := new(controller.PaiGowWebOutput)
	resObj.Hint = paiGowWebHint(pg)
	if resObj.Hint == nil {
		resObj.Message = "No hint is available now."
		resObj.MessageCode = "paigow.hintNone"
	}
	return marshalOrError(resObj)
}

// paiGowWebHint はドメインのヒントを JSON 用の形に移す。
func paiGowWebHint(pg interfaces.PaiGowGame) *controller.PaiGowWebOutputHint {
	hint := pg.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.PaiGowWebOutputHint{
		LowIdx0:   hint.LowIdx0,
		LowIdx1:   hint.LowIdx1,
		LowIsPair: hint.LowIsPair,
		Reason:    hint.Reason,
	}
}

// ActionLogOutput 棋譜をJSON出力
func (pp *PaiGowWebPresenter) ActionLogOutput(pg interfaces.PaiGowGame) string {
	return actionLogOutputJSON(pg)
}
