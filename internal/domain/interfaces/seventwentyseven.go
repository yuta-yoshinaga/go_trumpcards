//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SevenTwentySevenGame はセブン・トゥエンティセブン (SevenTwentySeven) のゲームインタフェース。
type SevenTwentySevenGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 次のラウンドを配る
	NextRound()
	// TakeCard 人間 (seat 0) の「引く / 止まる」を受け付ける。
	// **1 回では終わらない** —— 全員が止まるまで繰り返す。
	TakeCard(draw bool) error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SevenTwentySevenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SevenTwentySevenConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SevenTwentySevenPhase
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetCarryPot 次ラウンドへの持ち越し種銭を取得する
	GetCarryPot() int
	// GetCarryCount 全員アウトでポットが連続繰り越しになった回数を取得する
	GetCarryCount() int
	// GetAnte アンティ額を取得する
	GetAnte() int
	// GetLowWinner 7 側の勝者を取得する (-1 = 該当なし)
	GetLowWinner() int
	// GetHighWinner 27 側の勝者を取得する (-1 = 該当なし)
	GetHighWinner() int
	// GetDrawRound 何巡目の「引く / 止まる」かを取得する (1 始まり)
	GetDrawRound() int
	// GetScore playerIdx の side 側の得点 (内部値・×2) と、その側で生きているかを返す
	GetScore(playerIdx, side int) (int, bool)
	// GetMatchWinnerIdx ゲーム全体の勝者を取得する (-1 = 未確定)
	GetMatchWinnerIdx() int
	// GetResult 人間から見たラウンド結果を取得する
	GetResult() domain.SevenTwentySevenResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SevenTwentySevenPlayer
	// GetChips 人間 (seat 0) の保有チップを取得する
	GetChips() int
	// GetHint ヒントを取得する
	GetHint() *domain.SevenTwentySevenHint
}
