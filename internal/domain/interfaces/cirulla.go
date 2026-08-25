//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CirullaGame はチルッラのゲームインタフェース。
type CirullaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay 人間が 1 枚出す (captureIdxs が空なら場に置く)
	PlayerPlay(handIdx int, captureIdxs []int) error
	// CpuPlay CPU が 1 枚出す
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CirullaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CirullaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetTable 場札を取得する
	GetTable() []*domain.Card
	// GetRoundNumber 現在のラウンドを取得する
	GetRoundNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetLastCapturer 最後に捕獲した席を取得する (-1 = なし)
	GetLastCapturer() int
	// GetLastBonus 席ごとの直近の配札ボーナス識別子を取得する
	GetLastBonus() []string
	// GetDeckRemaining 山の残り枚数を取得する
	GetDeckRemaining() int
	// GetCaptureOptions 手札を出したときの捕獲候補を取得する
	GetCaptureOptions(playerIdx, handIdx int) [][]int
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.CirullaPlayer
	// GetLastResult 直前ラウンドの集計を取得する
	GetLastResult() *domain.CirullaRoundResult
	// GetWinnerIdx 勝者の席を取得する (-1 = 未決)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.CirullaHint
}
