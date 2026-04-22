package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CassinoGame カシノのゲームインタフェース。
type CassinoGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム開始)
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerTake 場札 / ビルドを捕獲する
	PlayerTake(handIdx int, tableIdxs []int, buildIdxs []int) error
	// PlayerBuild ビルドを作成する
	PlayerBuild(handIdx int, tableIdxs []int, declaredValue int) error
	// PlayerTrail 場に置くだけ (捕獲しない)
	PlayerTrail(handIdx int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.CassinoConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CassinoPlayer
	// GetTableCards 場の単独カード一覧を取得する
	GetTableCards() []*domain.Card
	// GetBuilds 場のビルド一覧を取得する
	GetBuilds() []*domain.CassinoBuild
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを返す (-1 = なし)
	GetLastCaptureIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.CassinoAction
	// GetCpuActions CPU 行動記録一覧を取得する
	GetCpuActions() []*domain.CassinoAction
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CassinoConfig
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// GetLastRoundDetail 直前ラウンドの得点詳細を取得する
	GetLastRoundDetail() *domain.CassinoScoreDetail
	// GetRoundWinners 勝者インデックス一覧を取得する
	GetRoundWinners() []int
	// GetRemainingDeck 山札残り枚数を取得する
	GetRemainingDeck() int
	// GetPacksDealt 既に配布されたパック数を取得する
	GetPacksDealt() int
}
