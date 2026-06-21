//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ScopaGame スコパのゲームインタフェース。
type ScopaGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム開始)
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay 手札を出す (tableIdxs が空なら場に置く、それ以外は捕獲)
	PlayerPlay(handIdx int, tableIdxs []int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.ScopaConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ScopaPlayer
	// GetTableCards 場札一覧を取得する
	GetTableCards() []*domain.Card
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを返す (-1 = なし)
	GetLastCaptureIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.ScopaAction
	// GetCpuActions CPU 行動記録一覧を取得する
	GetCpuActions() []*domain.ScopaAction
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ScopaConfig
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// GetLastRoundDetail 直前ラウンドの得点詳細を取得する
	GetLastRoundDetail() *domain.ScopaScoreDetail
	// GetRoundWinners 勝者インデックス一覧を取得する
	GetRoundWinners() []int
	// GetRemainingDeck 山札残り枚数を取得する
	GetRemainingDeck() int
	// GetPacksDealt 既に配布されたパック数を取得する
	GetPacksDealt() int
}
