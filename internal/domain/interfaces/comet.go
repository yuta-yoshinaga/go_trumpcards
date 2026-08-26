//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CometGame はコメットのゲームインタフェース。
type CometGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次の局を開始する
	NextRound()
	// PlayerPlay 人間が 1 枚出す
	PlayerPlay(handIdx int) error
	// PlayerPass 人間がパスする
	PlayerPass() error
	// CpuPlay CPU が 1 手打つ
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CometConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CometConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetPile 今の連なりに出た札を取得する
	GetPile() []*domain.Card
	// GetNeed 次に要るランクを取得する (0 = 連なりの先頭)
	GetNeed() int
	// GetDeadCount 死に手の枚数を取得する
	GetDeadCount() int
	// GetRoundNumber 現在の局を取得する
	GetRoundNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetLastPlayerIdx 最後に札を出した席を取得する (-1 = なし)
	GetLastPlayerIdx() int
	// PlayableIdxs 指定席が出せる手札の位置を取得する
	PlayableIdxs(seat int) []int
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.CometPlayer
	// GetLastResult 直前の局の集計を取得する
	GetLastResult() *domain.CometRoundResult
	// GetWinnerIdx 勝者の席を取得する (-1 = 未決)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.CometHint
}
