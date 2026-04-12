package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FiftyOneGame フィフティワンゲームインタフェース
type FiftyOneGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SetConfig 設定をセットする
	SetConfig(cfg domain.FiftyOneConfig)
	// GetConfig 設定を取得する
	GetConfig() domain.FiftyOneConfig

	// ExchangeOne 手札1枚と場札1枚を交換する
	ExchangeOne(handIdx, tableIdx int) error
	// ExchangeAll 手札5枚と場札5枚を全交換する
	ExchangeAll() error
	// Stop ストップ宣言する
	Stop() error
	// CpuPlay CPUのターンを実行する
	CpuPlay() error

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FiftyOnePhase
	// IsHumanTurn 人間のターンかどうか
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FiftyOnePlayer
	// GetCurrentTurn 現在のターンプレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetTableCards 場札を取得する
	GetTableCards() []*domain.Card
	// GetStopCallerIdx ストップ宣言者インデックスを取得する
	GetStopCallerIdx() int
	// GetTurnNumber ターン番号を取得する
	GetTurnNumber() int
	// GetLastAction 直前のアクション種別を取得する
	GetLastAction() string
	// GetLastHandIdx 直前の手札インデックスを取得する
	GetLastHandIdx() int
	// GetLastTableIdx 直前の場札インデックスを取得する
	GetLastTableIdx() int
}
