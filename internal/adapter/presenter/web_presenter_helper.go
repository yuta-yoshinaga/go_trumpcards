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
// 内部で buildWinnerResultMessage を呼び出してメッセージを生成する。
func buildWinnerWebMessage(gamePrefix string, winnerIdx int, isHuman bool) (string, string, map[string]string) {
	resultMsg := buildWinnerResultMessage(winnerIdx, isHuman)
	if isHuman {
		return resultMsg, gamePrefix + ".result.humanWin", nil
	}
	params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	return resultMsg, gamePrefix + ".result.cpuWin", params
}

// intSliceOrEmpty nil スライスを空スライスに正規化する (JSON で null を避ける)。
//
// **この共通ファイルはビルドタグを持たないので 6 つの Worker すべてに入る。**
// Schnapsen の presenter (solo タグ) に置いたままだと、同じヘルパを使う別
// カテゴリのゲームが classic の Worker ビルドだけで落ちる。
func intSliceOrEmpty(in []int) []int {
	if in == nil {
		return make([]int, 0)
	}
	return in
}
