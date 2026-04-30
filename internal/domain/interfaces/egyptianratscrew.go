package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// EgyptianRatscrewGame エジプシャン・ラットスクリューゲームインタフェース
type EgyptianRatscrewGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 設定を更新してゲームを初期化する
	ResetWithConfig(cfg domain.EgyptianRatscrewConfig)
	// Step 現手番プレイヤーがストック先頭1枚を場に出す
	Step() error
	// Slap 指定プレイヤーがスラップを試みる
	Slap(playerIdx int) error
	// Tick 保留中の CPU アクションを進行させる
	Tick() domain.EgyptianRatscrewPendingKind

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.EgyptianRatscrewConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.EgyptianRatscrewConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.EgyptianRatscrewPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.EgyptianRatscrewPlayer
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetCenterPileSize 場の総枚数を取得する
	GetCenterPileSize() int
	// GetTopCard 場のトップカードを取得する
	GetTopCard() *domain.Card
	// GetCurrentTurnIdx 現在の手番プレイヤーを取得する
	GetCurrentTurnIdx() int
	// IsTopFaceCard 場のトップが絵札 (J/Q/K/A) かを返す
	IsTopFaceCard() bool
	// IsSlappable 場の上 2 枚がペアまたは上 3 枚がサンドイッチかを返す
	IsSlappable() bool
	// GetChanceRemaining チャンスバトル中の残 flip 回数 (0 ならチャンスバトル外)
	GetChanceRemaining() int
	// GetChanceFromIdx チャンスを課したプレイヤーインデックス (チャンスバトル外では -1)
	GetChanceFromIdx() int
	// GetPending 保留中の CPU アクションを取得する
	GetPending() domain.EgyptianRatscrewPending
	// GetLastEvent 直近イベントを取得する
	GetLastEvent() domain.EgyptianRatscrewLastEvent
}
