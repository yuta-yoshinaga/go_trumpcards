//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CuckooGame Cuckoo (カッコー / Chase the Ace) ゲームインタフェース
type CuckooGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerKeep プレイヤーが手札を保持する
	PlayerKeep() error
	// PlayerSwap プレイヤーがスワップを試みる
	PlayerSwap() error
	// PlayerRefuse King を持つ隣人がスワップを拒否する
	PlayerRefuse() error
	// PlayerAcceptSwap King を持つ隣人がスワップを受け入れる
	PlayerAcceptSwap() error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CuckooConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CuckooConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CuckooPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetStockCount 残り山札の枚数を取得する
	GetStockCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CuckooPlayer
	// GetPendingSwapFrom スワップを要求中のプレイヤーインデックスを取得する
	GetPendingSwapFrom() int
	// GetPendingSwapTo スワップ要求先 (King 保持の隣人) を取得する
	GetPendingSwapTo() int
	// IsKingRevealed 指定プレイヤーの King が公開されているかを取得する
	IsKingRevealed(i int) bool
	// GetRoundLowest 直近ラウンドの最低カード値を取得する
	GetRoundLowest() int
	// GetRoundLosers 直近ラウンドでライフを失ったプレイヤーを取得する
	GetRoundLosers() []int
}
