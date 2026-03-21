package presenter

import "fmt"

// buildWinnerResultMessage はゲーム終了メッセージ「ゲーム終了！ Xの勝ち！」を生成する。
// 単一の勝者が決まるゲーム（Hearts, Spades 等）で共通利用する。
func buildWinnerResultMessage(winnerIdx int, isHuman bool) string {
	var name string
	if isHuman {
		name = "あなた"
	} else {
		name = fmt.Sprintf("CPU %d", winnerIdx)
	}
	return fmt.Sprintf("ゲーム終了！ %sの勝ち！", name)
}

// buildWinnerWebMessage はゲーム終了時の message, messageCode, messageParams を構築する。
// humanWin/cpuWin パターンを持つゲームで共通利用する。
func buildWinnerWebMessage(resultMsg, gamePrefix string, winnerIdx int, isHuman bool) (string, string, map[string]string) {
	if isHuman {
		return resultMsg, gamePrefix + ".result.humanWin", nil
	}
	params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	return resultMsg, gamePrefix + ".result.cpuWin", params
}
