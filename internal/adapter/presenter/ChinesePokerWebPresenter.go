//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ChinesePokerWebPresenter チャイニーズポーカーWebプレゼンタークラス
type ChinesePokerWebPresenter struct {
}

// Output ゲーム状態を出力
func (pp *ChinesePokerWebPresenter) Output(cp interfaces.ChinesePokerGame, lastErr error) string {
	resObj := new(controller.ChinesePokerWebOutput)

	resObj.PlayerCards = cardsToOutputOrEmpty(cp.GetPlayerCards())
	resObj.DealerCards = cardsToOutputOrEmpty(cp.GetDealerCards())
	resObj.PlayerFront = cardsToOutputOrEmpty(cp.GetPlayerFront())
	resObj.PlayerMiddle = cardsToOutputOrEmpty(cp.GetPlayerMiddle())
	resObj.PlayerBack = cardsToOutputOrEmpty(cp.GetPlayerBack())
	resObj.DealerFront = cardsToOutputOrEmpty(cp.GetDealerFront())
	resObj.DealerMiddle = cardsToOutputOrEmpty(cp.GetDealerMiddle())
	resObj.DealerBack = cardsToOutputOrEmpty(cp.GetDealerBack())
	resObj.Phase = cp.GetPhase()
	resObj.Chips = cp.GetChips()
	resObj.Bet = cp.GetBet()
	resObj.Result = int(cp.GetResult())
	resObj.FrontResult = int(cp.GetFrontResult())
	resObj.MiddleResult = int(cp.GetMiddleResult())
	resObj.BackResult = int(cp.GetBackResult())
	resObj.Payout = cp.GetPayout()
	resObj.PlayerFrontRank = cp.GetPlayerFrontRank()
	resObj.PlayerMiddleRank = cp.GetPlayerMiddleRank()
	resObj.PlayerBackRank = cp.GetPlayerBackRank()
	resObj.DealerFrontRank = cp.GetDealerFrontRank()
	resObj.DealerMiddleRank = cp.GetDealerMiddleRank()
	resObj.DealerBackRank = cp.GetDealerBackRank()
	resObj.PlayerRoyalty = cp.GetPlayerRoyalty()
	resObj.DealerRoyalty = cp.GetDealerRoyalty()
	resObj.Scoop = cp.GetScoop()
	// 推奨分割はセットハンドで13枚そろっているときだけ入る (ドメイン側の判定)。
	resObj.SuggestedArrangement = chinesePokerArrangementOutput(cp.GetSuggestedArrangement())

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cp.GetGameEndFlag() {
		switch cp.GetResult() {
		case domain.GameResultWin:
			if cp.GetScoop() {
				resObj.Message = "Player scoop!"
				resObj.MessageCode = "chinesepoker.result.playerScoop"
			} else {
				resObj.Message = "Player wins!"
				resObj.MessageCode = "chinesepoker.result.playerWins"
			}
		case domain.GameResultLose:
			if cp.GetScoop() {
				resObj.Message = "Dealer scoop!"
				resObj.MessageCode = "chinesepoker.result.dealerScoop"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "chinesepoker.result.dealerWins"
			}
		default:
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを出力する。専用のレスポンスは持たず通常の状態出力を返すが、
// **その状態に推奨分割 (SuggestedArrangement) が載っている** (#5615)。フロントの
// useGameHint は、載っていればそれを使い、無いときだけ自前の計算に落ちる。
func (cp *ChinesePokerWebPresenter) HintOutput(g interfaces.ChinesePokerGame) string {
	return cp.Output(g, nil)
}

// chinesePokerArrangementOutput はドメインの推奨分割をレスポンス用に写す。
// **並べ直さない。**presenter で組み直すと、CUI (#4717) と Web が同じ手札で
// 違う分け方を勧めることになる (#5615)。
func chinesePokerArrangementOutput(a *domain.ChinesePokerSuggestedArrangement) *controller.ChinesePokerWebOutputArrangement {
	if a == nil {
		return nil
	}
	return &controller.ChinesePokerWebOutputArrangement{
		Front:  a.Front,
		Middle: a.Middle,
		Back:   a.Back,
		Foul:   a.Foul,
	}
}

// ActionLogOutput 棋譜をJSON出力
func (pp *ChinesePokerWebPresenter) ActionLogOutput(cp interfaces.ChinesePokerGame) string {
	return actionLogOutputJSON(cp)
}
