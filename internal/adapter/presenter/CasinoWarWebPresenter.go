package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CasinoWarWebPresenter カジノウォーWebプレゼンタークラス
type CasinoWarWebPresenter struct{}

// Output ゲーム状態を出力
func (cp *CasinoWarWebPresenter) Output(cw interfaces.CasinoWarGame, lastErr error) string {
	resObj := new(controller.CasinoWarWebOutput)

	if cw.GetPlayerCard() != nil {
		resObj.PlayerCard = cardToOutput(cw.GetPlayerCard())
	}
	if cw.GetDealerCard() != nil {
		resObj.DealerCard = cardToOutput(cw.GetDealerCard())
	}
	if cw.GetPlayerWarCard() != nil {
		resObj.PlayerWarCard = cardToOutput(cw.GetPlayerWarCard())
	}
	if cw.GetDealerWarCard() != nil {
		resObj.DealerWarCard = cardToOutput(cw.GetDealerWarCard())
	}
	resObj.BurnCards = cardsToOutputOrEmpty(cw.GetBurnCards())
	resObj.Phase = cw.GetPhase()
	resObj.Chips = cw.GetChips()
	resObj.Ante = cw.GetAnte()
	resObj.WarBet = cw.GetWarBet()
	resObj.Result = int(cw.GetResult())
	resObj.TotalPayout = cw.GetTotalPayout()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cw.GetGameEndFlag() {
		resObj.Message, resObj.MessageCode = casinoWarEndMessage(cw)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (cp *CasinoWarWebPresenter) ActionLogOutput(cw interfaces.CasinoWarGame) string {
	return actionLogOutputJSON(cw)
}

// casinoWarEndMessage は終了時の表示メッセージと i18n キーを返す
func casinoWarEndMessage(cw interfaces.CasinoWarGame) (string, string) {
	switch cw.GetResult() {
	case domain.GameResultWin:
		if cw.GetWarBet() > 0 {
			pr, dr := casinoWarRanks(cw)
			if pr == dr {
				return "Player wins (war tie).", "casinowar.result.warTieWin"
			}
			return "Player wins the war!", "casinowar.result.warWin"
		}
		return "Player wins!", "casinowar.result.playerWins"
	case domain.GameResultLose:
		if cw.GetWarBet() > 0 {
			return "Player loses the war.", "casinowar.result.warLoss"
		}
		if cw.GetTotalPayout() > 0 {
			return "Surrendered.", "casinowar.result.surrender"
		}
		return "Player loses.", "casinowar.result.dealerWins"
	default:
		return "Push.", "casinowar.result.tie"
	}
}

// casinoWarRanks は war 後の両者ランクを返す（同値判定用）
func casinoWarRanks(cw interfaces.CasinoWarGame) (int, int) {
	pc, dc := cw.GetPlayerWarCard(), cw.GetDealerWarCard()
	if pc == nil || dc == nil {
		return 0, 0
	}
	return casinoWarRankOf(pc), casinoWarRankOf(dc)
}

// casinoWarRankOf は A=14 のランクに変換する（プレゼンター層用ヘルパー）
func casinoWarRankOf(c *domain.Card) int {
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}
