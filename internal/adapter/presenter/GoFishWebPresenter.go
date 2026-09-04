package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GoFishWebPresenter Go FishWebプレゼン��ークラス
type GoFishWebPresenter struct{}

// Output ���ーム状態をJSON出���
func (p *GoFishWebPresenter) Output(gf interfaces.GoFishGame, lastErr error) string {
	resObj := new(controller.GoFishWebOutput)
	resObj.Phase = int(gf.GetPhase())
	resObj.CurrentTurn = gf.GetCurrentTurn()
	resObj.GameEndFlag = gf.GetGameEndFlag()
	resObj.WinnerIdx = gf.GetWinnerIdx()
	resObj.TurnNumber = gf.GetTurnNumber()
	resObj.DeckRemaining = gf.GetDeckRemaining()
	resObj.Config = controller.GoFishWebOutputConfig{
		CpuDifficulty: int(gf.GetConfig().CpuDifficulty),
	}

	// プレイヤー情報
	resObj.Players = make([]*controller.GoFishWebOutputPlayer, 0, gf.GetPlayerCnt())
	knownRanks := gf.GetKnownRanks()
	for i := 0; i < gf.GetPlayerCnt(); i++ {
		player := gf.GetPlayer(i)
		pObj := &controller.GoFishWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
			BookCount: player.GetBookCount(),
			Books:     booksToOutput(player.GetBooks()),
			// Sent per seat so a reload restores what the table already knows.
			KnownRanks: knownRanks[i],
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// 最後の要求情報
	if gf.GetLastAskPlayerIdx() >= 0 {
		lastAsk := &controller.GoFishWebOutputLastAsk{
			PlayerIdx:  gf.GetLastAskPlayerIdx(),
			TargetIdx:  gf.GetLastAskTargetIdx(),
			Rank:       gf.GetLastAskRank(),
			Success:    gf.GetLastAskSuccess(),
			BookFormed: gf.GetLastBookFormed(),
			BookRank:   gf.GetLastBookRank(),
		}
		if gf.GetLastAskSuccess() {
			lastAsk.CardsReceived = cardsToOutputOrEmpty(gf.GetLastCardsReceived())
		}
		// Show drawn card only if it was the human who drew
		if gf.GetLastDrawnCard() != nil {
			lastDrawPlayer := gf.GetPlayer(gf.GetLastAskPlayerIdx())
			if lastDrawPlayer != nil && lastDrawPlayer.GetIsHuman() {
				lastAsk.DrawnCard = cardToOutput(gf.GetLastDrawnCard())
			}
		}
		resObj.LastAsk = lastAsk
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.GoFishWebOutputCpuAction, 0)
	for _, action := range gf.GetCpuActions() {
		a := &controller.GoFishWebOutputCpuAction{
			AskPlayerIdx:  action.AskPlayerIdx,
			AskTargetIdx:  action.AskTargetIdx,
			AskRank:       action.AskRank,
			Success:       action.Success,
			CardsReceived: action.CardsReceived,
			BookFormed:    action.BookFormed,
			BookRank:      action.BookRank,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間プレイヤーの行動記録
	if ha := gf.GetHumanAction(); ha != nil {
		resObj.HumanAction = &controller.GoFishWebOutputCpuAction{
			AskPlayerIdx:  ha.AskPlayerIdx,
			AskTargetIdx:  ha.AskTargetIdx,
			AskRank:       ha.AskRank,
			Success:       ha.Success,
			CardsReceived: ha.CardsReceived,
			DrawnCard:     cardToOutput(ha.DrawnCard),
			BookFormed:    ha.BookFormed,
			BookRank:      ha.BookRank,
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if gf.GetGameEndFlag() {
		winnerIdx := gf.GetWinnerIdx()
		winner := gf.GetPlayer(winnerIdx)
		if winner != nil && winner.GetIsHuman() {
			resObj.Message = "ゲーム終了！ あなたの勝ち！"
			resObj.MessageCode = "gofish.result.humanWin"
		} else {
			resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx)
			resObj.MessageCode = "gofish.result.cpuWin"
			resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GoFishWebPresenter) ActionLogOutput(gf interfaces.GoFishGame) string {
	return actionLogOutputJSON(gf)
}

// booksToOutput ブック配列をWeb出力形式に変換する
func booksToOutput(books [][]*domain.Card) []*controller.GoFishWebOutputBook {
	result := make([]*controller.GoFishWebOutputBook, 0, len(books))
	for _, book := range books {
		if len(book) == 0 {
			continue
		}
		result = append(result, &controller.GoFishWebOutputBook{
			Rank:  book[0].GetValue(),
			Cards: cardsToOutputOrEmpty(book),
		})
	}
	return result
}
