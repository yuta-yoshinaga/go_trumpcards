//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CuarentaGame クアレンタのゲームインタフェース。
type CuarentaGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム開始)
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay 手札を出す (同ランク捕獲、なければ場に置く)
	PlayerPlay(handIdx int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.CuarentaConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CuarentaPlayer
	// GetTeamScore 指定チームの累計点を取得する
	GetTeamScore(team int) int
	// GetTableCards 場札一覧を取得する
	GetTableCards() []*domain.Card
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを返す (-1 = なし)
	GetLastCaptureIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.CuarentaAction
	// GetCpuActions CPU 行動記録一覧を取得する
	GetCpuActions() []*domain.CuarentaAction
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CuarentaConfig
	// GetPhase 現在のフェーズ (数値) を取得する
	GetPhase() int
	// GetLastRoundDetail 直前ラウンドの得点詳細を取得する
	GetLastRoundDetail() *domain.CuarentaRoundDetail
	// GetRoundWinners 勝者チームインデックス一覧を取得する
	GetRoundWinners() []int
	// GetRemainingDeck 山札残り枚数を取得する
	GetRemainingDeck() int
}
