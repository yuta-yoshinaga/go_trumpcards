//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SakuraGame はさくら (肥後花) のゲームインタフェース。
type SakuraGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay 人間が手札を出す (fieldIdx で 2 枚一致時の獲得対象を指定; 不要なら -1)
	PlayerPlay(handIdx, fieldIdx int) error
	// CpuPlay CPU の手番を 1 ステップ実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SakuraConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SakuraConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SakuraPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// HumanSeat 人間の席を取得する
	HumanSeat() int
	// GetTurn 現在の手番の席を取得する
	GetTurn() int
	// GetDealer 親の席を取得する
	GetDealer() int
	// GetField 場札を取得する
	GetField() []*domain.Card
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetRound 現在のラウンド番号を取得する
	GetRound() int
	// GetWinner 終局時の勝者を取得する (-1=引き分け/未決)
	GetWinner() int
	// GetLastResult 直近のラウンド結果を取得する
	GetLastResult() *domain.SakuraRoundResult
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定した席を取得する
	GetPlayer(i int) *domain.SakuraPlayer
	// GetValidFieldIndices 各手札が合わせられる場札インデックスを取得する
	GetValidFieldIndices() map[int][]int
	// GetHint ヒントを取得する
	GetHint() domain.SakuraHint
}
