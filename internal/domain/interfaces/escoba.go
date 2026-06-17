//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// EscobaGame エスコバのゲームインタフェース (4 人フリーフォーオール、チームなし)。
type EscobaGame interface {
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
	SetConfig(config domain.EscobaConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ScopaPlayer
	// GetTableCards 場札一覧を取得する
	GetTableCards() []*domain.Card
	// GetStockRemaining 山札の残り枚数を取得する
	GetStockRemaining() int
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを返す (-1 = なし)
	GetLastCaptureIdx() int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetWinnerIdx 勝利プレイヤーを取得する (-1 = 未確定)
	GetWinnerIdx() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.EscobaConfig
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// GetLastRoundDetail 直前ラウンドの得点詳細を取得する
	GetLastRoundDetail() *domain.EscobaScoreDetail
	// GetValidCaptures handIdx のカードで可能な捕獲パターンを取得する
	GetValidCaptures(handIdx int) [][]int
}
