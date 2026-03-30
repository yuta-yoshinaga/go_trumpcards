package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpeedWebPresenter スピードWebプレゼンタークラス
type SpeedWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SpeedWebPresenter) Output(s interfaces.SpeedGame, lastErr error) string {
	resObj := new(controller.SpeedWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Config = controller.SpeedWebOutputConfig{
		CpuDifficulty: int(s.GetConfig().CpuDifficulty),
	}

	// 台札
	resObj.CenterPiles = make([]*controller.WebOutputCard, 0, 2)
	for i := range 2 {
		resObj.CenterPiles = append(resObj.CenterPiles, cardToOutput(s.GetCenterPile(i)))
	}

	// プレイヤー情報
	resObj.Players = make([]*controller.SpeedWebOutputPlayer, 0, s.GetPlayerCnt())
	for i := range s.GetPlayerCnt() {
		player := s.GetPlayer(i)
		pObj := &controller.SpeedWebOutputPlayer{
			ID:           i,
			IsHuman:      player.GetIsHuman(),
			CardCount:    player.GetCardsSize(),
			Cards:        playerCardsToOutput(player, player.GetIsHuman()),
			DrawPileSize: player.GetDrawPileSize(),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// ヒント
	ci, pi, found := s.GetHint()
	if found {
		resObj.Hint = &controller.SpeedWebOutputHint{
			CardIndex: ci,
			PileIndex: pi,
			Found:     true,
		}
	}

	// メッセージ
	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildSpeedMessage(s, lastErr)

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *SpeedWebPresenter) ActionLogOutput(s interfaces.SpeedGame) string {
	return actionLogOutputJSON(s)
}

// buildSpeedMessage ゲーム状態に応じたメッセージを生成する
func buildSpeedMessage(s interfaces.SpeedGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if s.GetGameEndFlag() {
		return "", "gameEnd", nil
	}
	if s.GetPhase() == domain.SpeedPhaseStuck {
		return "", "stuck", nil
	}
	return "", "play", nil
}
