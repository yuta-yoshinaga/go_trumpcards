package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DurakWebPresenter ドゥラークWebプレゼンタークラス
type DurakWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (dwp *DurakWebPresenter) Output(dg interfaces.DurakGame, lastErr error) string {
	resObj := new(controller.DurakWebOutput)
	resObj.Players = make([]*controller.DurakWebOutputPlayer, 0)
	resObj.CurrentTurn = dg.GetCurrentTurn()
	resObj.Phase = int(dg.GetPhase())
	resObj.AttackerIdx = dg.GetAttackerIdx()
	resObj.DefenderIdx = dg.GetDefenderIdx()
	resObj.StockCount = dg.GetStockCount()
	resObj.LoserIdx = dg.GetLoserIdx()
	resObj.GameEndFlag = dg.GetGameEndFlag()
	resObj.BoutNumber = dg.GetBoutNumber()
	resObj.SortMode = int(dg.GetSortMode())

	// 切り札
	resObj.TrumpSuit = cardDesignToString(dg.GetTrumpSuit())
	resObj.TrumpCard = cardToOutput(dg.GetTrumpCard())

	// 設定
	config := dg.GetConfig()
	resObj.Config = controller.DurakWebConfig{
		PlayerCount:     config.PlayerCount,
		CpuDifficulty:   int(config.CpuDifficulty),
		TransferEnabled: config.TransferEnabled,
	}

	// テーブルペア
	resObj.TablePairs = make([]*controller.DurakWebOutputTablePair, 0)
	for _, pair := range dg.GetTablePairs() {
		resObj.TablePairs = append(resObj.TablePairs, &controller.DurakWebOutputTablePair{
			Attack:  cardToOutput(pair.Attack),
			Defense: cardToOutput(pair.Defense),
		})
	}

	// CPU行動記録
	resObj.CpuActions = make([]*controller.DurakWebOutputAction, 0)
	for _, action := range dg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, &controller.DurakWebOutputAction{
			PlayerIdx:  action.PlayerIdx,
			ActionType: action.ActionType,
			Card:       cardToOutput(action.Card),
			AttackIdx:  action.AttackIdx,
		})
	}

	// 人間の最後の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.DurakWebOutputAction{
			PlayerIdx:  humanAction.PlayerIdx,
			ActionType: humanAction.ActionType,
			Card:       cardToOutput(humanAction.Card),
			AttackIdx:  humanAction.AttackIdx,
		}
	}

	// プレイヤー情報
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := &controller.DurakWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// エラーメッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if dg.GetGameEndFlag() {
		resObj.Message = dwp.buildResultMessage(dg)
		resObj.MessageCode = "durak.result"
		loserIdx := dg.GetLoserIdx()
		if loserIdx >= 0 {
			resObj.MessageParams = map[string]string{"loserIdx": fmt.Sprintf("%d", loserIdx)}
		}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (dwp *DurakWebPresenter) buildResultMessage(dg interfaces.DurakGame) string {
	loserIdx := dg.GetLoserIdx()
	if loserIdx < 0 {
		return "引き分け！"
	}
	player := dg.GetPlayer(loserIdx)
	if player.GetIsHuman() {
		return "ゲーム終了！ あなたがドゥラーク（負け）です"
	}
	return fmt.Sprintf("ゲーム終了！ CPU %d がドゥラーク（負け）です", loserIdx)
}

// ActionLogOutput 棋譜をJSON出力
func (dwp *DurakWebPresenter) ActionLogOutput(dg interfaces.DurakGame) string {
	return actionLogOutputJSON(dg)
}

// HintOutput はサーバー計算のヒントを返す (`command: "hint"` 専用のレスポンス)。
func (p *DurakWebPresenter) HintOutput(g interfaces.DurakGame) string {
	return p.Output(g, nil)
}
