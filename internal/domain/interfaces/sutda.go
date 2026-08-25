//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SutdaGame はソッタのゲームインタフェース。
type SutdaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次のハンドを開始する
	NextHand()
	// PlayerAction 人間が 1 手打つ (call / raise / fold)
	PlayerAction(action string) error
	// CpuAction CPU が 1 手打つ
	CpuAction()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SutdaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SutdaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetPot 場のチップを取得する
	GetPot() int
	// GetCurrentBet 現在のベット額を取得する
	GetCurrentBet() int
	// GetRaiseCount このハンドのレイズ回数を取得する
	GetRaiseCount() int
	// GetCallAmount コールに必要な額を取得する (0 ならチェック)
	GetCallAmount(playerIdx int) int
	// CanRaise その席がいまレイズできるかを取得する
	CanRaise(playerIdx int) bool
	// GetHandOf 席の役を取得する
	GetHandOf(playerIdx int) domain.SutdaHand
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.SutdaPlayer
	// GetLastResult 直前ハンドの結果を取得する
	GetLastResult() *domain.SutdaHandResult
	// GetWinnerIdx 最終的な勝者の席を取得する (-1 = 未決)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.SutdaHint
}
