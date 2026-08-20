package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WarWebPresenter 戦争Webプレゼンタークラス
type WarWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WarWebPresenter) Output(w interfaces.WarGame, lastErr error) string {
	resObj := new(controller.WarWebOutput)
	resObj.Phase = int(w.GetPhase())
	resObj.GameEndFlag = w.GetGameEndFlag()
	resObj.WinnerIdx = w.GetWinnerIdx()
	resObj.PlayerRevealed = cardToOutput(w.GetPlayerRevealed())
	resObj.CpuRevealed = cardToOutput(w.GetCpuRevealed())
	resObj.WarPotSize = w.GetWarPotSize()
	resObj.LastWinnerIdx = w.GetLastWinnerIdx()
	resObj.LastBurialCount = w.GetLastBurialCount()
	resObj.RoundsPlayed = w.GetRoundsPlayed()
	resObj.Config = controller.WarWebOutputConfig{
		MaxRounds: w.GetConfig().MaxRounds,
	}

	resObj.Players = make([]*controller.WarWebOutputPlayer, 0, w.GetPlayerCnt())
	for i := range w.GetPlayerCnt() {
		player := w.GetPlayer(i)
		resObj.Players = append(resObj.Players, &controller.WarWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			DrawPileSize:    player.GetDrawPileSize(),
			DiscardPileSize: player.GetDiscardPileSize(),
			TotalCards:      player.TotalCards(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildWarMessage(w, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *WarWebPresenter) ActionLogOutput(w interfaces.WarGame) string {
	return actionLogOutputJSON(w)
}

// buildWarMessage ゲーム状態に応じたメッセージを生成する
func buildWarMessage(w interfaces.WarGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if w.GetGameEndFlag() {
		if w.GetWinnerIdx() == 0 {
			return "", "war.result.humanWin", nil
		}
		return "", "war.result.cpuWin", nil
	}
	// **各ラウンドの決着も伝える。** 以前は最終勝敗のときしかコードを返さず、
	// 盤面の変化はリング色と不透明度だけだった。読み上げ利用者はどのラウンドで
	// 誰が勝ったのかも、いつ戦争になったのかも知りようがない (#5530)。
	// CUI の WarCuiPresenter は毎ラウンド promptResolved*/promptWarBury を出している。
	switch w.GetPhase() {
	case domain.WarPhaseResolved:
		if w.GetLastWinnerIdx() == 0 {
			return "", "war.round.humanWin", nil
		}
		return "", "war.round.cpuWin", nil
	case domain.WarPhaseWarBury:
		return "", "war.round.warBury", nil
	}
	return "", "", nil
}
