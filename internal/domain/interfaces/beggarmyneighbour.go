package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BeggarMyNeighbourGame Beggar-My-Neighbour ゲームインタフェース
type BeggarMyNeighbourGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Step 状態機械を1ステップ進める
	Step() error
	// AutoPlay 決着まで Step を自動で繰り返す
	AutoPlay() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BeggarMyNeighbourConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BeggarMyNeighbourConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BeggarMyNeighbourPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BeggarMyNeighbourPlayer
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetPenaltyOwnerIdx ペナルティ所有者インデックスを取得する
	GetPenaltyOwnerIdx() int
	// GetPenaltyRemaining 残りペナルティ枚数を取得する
	GetPenaltyRemaining() int
	// GetCentralPileSize 場の山の枚数を取得する
	GetCentralPileSize() int
	// GetLastCardPlayed 最後に出されたカードを取得する
	GetLastCardPlayed() *domain.Card
	// GetRoundsPlayed 消化ラウンド数を取得する
	GetRoundsPlayed() int
}
