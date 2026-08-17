package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EgyptianRatscrewWebPresenter エジプシャン・ラットスクリュー Web プレゼンター
type EgyptianRatscrewWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *EgyptianRatscrewWebPresenter) Output(g interfaces.EgyptianRatscrewGame, lastErr error) string {
	resObj := new(controller.EgyptianRatscrewWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentTurnIdx = g.GetCurrentTurnIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsTopFaceCard = g.IsTopFaceCard()
	resObj.IsSlappable = g.IsSlappable()
	resObj.CenterPileSize = g.GetCenterPileSize()
	resObj.TopCard = cardToOutput(g.GetTopCard())
	resObj.CpuDifficulty = int(g.GetConfig().CpuDifficulty)
	resObj.ChanceRemaining = g.GetChanceRemaining()
	// 規則の説明を画面に書くための回数。ドメインの FaceCardChances から取る ──
	// 訳文に 1/2/3/4 を焼き込むと、回数を変えたとき説明だけが嘘になる (#5580)。
	resObj.FaceChances = controller.NewEgyptianRatscrewWebFaceChances()
	resObj.ChanceFromIdx = g.GetChanceFromIdx()

	pending := g.GetPending()
	resObj.PendingKind = int(pending.Kind)
	resObj.PendingDeadlineMs = pending.DeadlineMs

	last := g.GetLastEvent()
	resObj.LastEventKind = int(last.Kind)
	resObj.LastEventPlayerIdx = last.PlayerIdx
	resObj.LastSlapReason = int(last.SlapReason)

	resObj.Players = make([]*controller.EgyptianRatscrewWebPlayer, 0, g.GetPlayerCnt())
	for i := range g.GetPlayerCnt() {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := "あなた"
		if !player.GetIsHuman() {
			name = "CPU"
		}
		resObj.Players = append(resObj.Players, &controller.EgyptianRatscrewWebPlayer{
			Name:      name,
			IsHuman:   player.GetIsHuman(),
			StockSize: player.GetStockSize(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildEgyptianRatscrewMessage(g, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *EgyptianRatscrewWebPresenter) ActionLogOutput(g interfaces.EgyptianRatscrewGame) string {
	return actionLogOutputJSON(g)
}

// buildEgyptianRatscrewMessage ゲーム状態に応じたメッセージを生成する。
// 終局フラグはエラーより優先する。詰みで Step が ErrInvalidPlay を返してもゲーム終了画面を出すため。
func buildEgyptianRatscrewMessage(g interfaces.EgyptianRatscrewGame, lastErr error) (string, string, map[string]string) {
	if g.GetGameEndFlag() {
		switch g.GetWinnerIdx() {
		case 0:
			return "", "egyptianratscrew.result.humanWin", nil
		case -1:
			return "", "egyptianratscrew.result.draw", nil
		default:
			return "", "egyptianratscrew.result.cpuWin", nil
		}
	}
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	return "", "", nil
}
