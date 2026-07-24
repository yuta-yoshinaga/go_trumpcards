//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GutsGame はガッツ (Guts) のゲームインタフェース。
type GutsGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 次のラウンドを配る
	NextRound()
	// Declare 人間 (seat 0) の in/out 宣言を受け付けラウンドを解決する
	Declare(stay bool) error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GutsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GutsConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GutsPhase
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
	// GetWinnerIdx 直近ラウンドの勝者を取得する (-1 = なし)
	GetWinnerIdx() int
	// GetMatchWinnerIdx ゲーム全体の勝者を取得する (-1 = 未確定)
	GetMatchWinnerIdx() int
	// GetResult 人間から見たラウンド結果を取得する
	GetResult() domain.GutsResult
	// GetMatchers このラウンドでマッチしたプレイヤーのインデックス列を取得する
	GetMatchers() []int
	// IsMatcher 指定プレイヤーがこのラウンドでマッチしたかを返す
	IsMatcher(idx int) bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GutsPlayer
	// GetChips 人間 (seat 0) の保有チップを取得する
	GetChips() int
	// GetHint ヒントを取得する
	GetHint() *domain.GutsHint
}
