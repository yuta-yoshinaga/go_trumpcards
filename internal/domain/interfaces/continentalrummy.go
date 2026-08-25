//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ContinentalRummyGame はコンチネンタル・ラミーのゲームインタフェース。
type ContinentalRummyGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// DrawStock 山札から 1 枚引く
	DrawStock() error
	// DrawDiscard 捨て札の一番上を取る
	DrawDiscard() error
	// Discard 手札の i 番を捨てる
	Discard(i int) error
	// GoOut 15 枚を並べて上がる (i は捨てる 1 枚)
	GoOut(i int) error
	// NextRound 次のラウンドへ進む
	NextRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ContinentalRummyConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ContinentalRummyConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の番かを取得する
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetRoundNumber 何ラウンド目かを取得する
	GetRoundNumber() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDiscardTop 捨て札の一番上を取得する
	GetDiscardTop() *domain.Card
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席を取得する
	GetPlayer(i int) *domain.ContinentalRummyPlayer
	// GetLastResult 直前のラウンドの結果を取得する
	GetLastResult() *domain.ContinentalRummyRoundResult
	// GetWinnerIdx 勝者の席を取得する (-1 = 引き分け)
	GetWinnerIdx() int
	// CanGoOut いま上がれるかと、そのとき捨てる札を取得する
	CanGoOut() (int, bool)
	// CanGoOutOnTheDeal 引かずに、配られた 15 枚のまま上がれるかを取得する
	CanGoOutOnTheDeal() bool
	// GetHint ヒントを取得する
	GetHint() *domain.ContinentalRummyHint
}
