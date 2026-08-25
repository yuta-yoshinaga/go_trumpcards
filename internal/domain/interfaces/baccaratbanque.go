//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BaccaratBanqueGame はバカラ・バンクのゲームインタフェース。
type BaccaratBanqueGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextCoup 次のクーを開始する
	NextCoup()
	// BankerDraw バンカー (人間) が 3 枚目を引くかを決める
	BankerDraw(draw bool) error
	// Retire バンカーが自分からバンクを降りる
	Retire() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BaccaratBanqueConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BaccaratBanqueConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の判断待ちかを取得する
	IsHumanTurn() bool
	// GetCoupNumber 何回目のクーかを取得する
	GetCoupNumber() int
	// GetBankHeld このバンクで続けたクー数を取得する
	GetBankHeld() int
	// GetShoeRemaining シューの残り枚数を取得する
	GetShoeRemaining() int
	// IsRetired バンカーが自分から降りたかを取得する
	IsRetired() bool
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席を取得する
	GetPlayer(i int) *domain.BaccaratBanquePlayer
	// GetLastResult 直前のクーの結果を取得する
	GetLastResult() *domain.BaccaratBanqueCoupResult
	// GetWinnerIdx 勝者の席を取得する (-1 = バンカーの負け越し)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.BaccaratBanqueHint
}
