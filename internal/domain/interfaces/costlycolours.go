//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CostlyColoursGame はコストリー・カラーズのゲームインタフェース。
type CostlyColoursGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextDeal 次のディールを開始する
	NextDeal()
	// PlayerMog 人間が交換に応じるかを決める
	PlayerMog(accept bool) error
	// PlayerPlay 人間が 1 枚出す
	PlayerPlay(handIdx int) error
	// CpuAct CPU が 1 手打つ (交換の可否も含む)
	CpuAct()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CostlyColoursConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CostlyColoursConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetTurnUp 表に返した 1 枚を取得する
	GetTurnUp() *domain.Card
	// GetPile 今の数え上げに出た札を取得する
	GetPile() []*domain.Card
	// GetTotal 今の数え上げの累計を取得する
	GetTotal() int
	// GetWentOut 「ゴー」を宣言した席を取得する (-1 = なし)
	GetWentOut() int
	// GetDealNumber 現在のディールを取得する
	GetDealNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// PlayableIdxs 指定席が出せる手札の位置を取得する
	PlayableIdxs(seat int) []int
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.CostlyColoursPlayer
	// GetLastResult 直前のディールの集計を取得する
	GetLastResult() *domain.CostlyColoursDealResult
	// GetWinnerIdx 勝者の席を取得する (-1 = 未決)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.CostlyColoursHint
}
