package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// WarGame 戦争ゲームインタフェース
type WarGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Step 状態機械を1ステップ進める
	Step() error
	// AutoPlay 決着まで Step を自動で繰り返す
	AutoPlay() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.WarConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.WarConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.WarPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.WarPlayer
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerRevealed 人間側の表札を取得する
	GetPlayerRevealed() *domain.Card
	// GetCpuRevealed CPU側の表札を取得する
	GetCpuRevealed() *domain.Card
	// GetWarPotSize 場に出ている総枚数を取得する
	GetWarPotSize() int
	// GetLastWinnerIdx 直近ラウンドの勝者インデックスを取得する
	GetLastWinnerIdx() int
	// GetLastBurialCount 直近の戦争で伏せた枚数を取得する
	GetLastBurialCount() int
	// GetRoundsPlayed 消化ラウンド数を取得する
	GetRoundsPlayed() int
}
