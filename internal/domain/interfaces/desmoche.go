//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DesmocheGame デスモチェゲームインタフェース
type DesmocheGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// DrawFromStock 山札から1枚引く
	DrawFromStock(player int) error
	// DrawFromDiscard 捨て札の一番上を取る
	DrawFromDiscard(player int) error
	// Meld 手札の添字集合をメルドとして出す
	Meld(player int, handIdxs []int) error
	// LayOff 手札1枚を既存のメルドに付ける
	LayOff(player, handIdx, meldIdx int) error
	// Desmoche 自分の場のメルドから1枚を抜いて別のメルドへ移す
	Desmoche(player, fromMeldIdx, cardIdx, toMeldIdx int) error
	// Discard 手札1枚を捨てて手番を終える
	Discard(player, handIdx int) error
	// NextRound 次のラウンドを配る
	NextRound() error
	// DesmocheCpuDecide CPU が取る手を決める
	DesmocheCpuDecide(idx int) domain.DesmocheCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DesmocheConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DesmocheConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DesmochePhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDiscardTop 捨て札の一番上を取得する
	GetDiscardTop() *domain.Card
	// GetMelds 場のメルドを取得する
	GetMelds() []*domain.DesmocheMeld
	// MeldedCount 指定プレイヤーが場に出している総枚数を取得する
	MeldedCount(player int) int
	// GetPot 場の掛け金を取得する
	GetPot() int
	// GetScore 収支を取得する
	GetScore(idx int) int
	// GetRoundNumber 完了したラウンド数を取得する
	GetRoundNumber() int
	// GetRoundWinner 直近ラウンドの勝者を取得する (-1: 勝者なし)
	GetRoundWinner() int
	// IsRoundExhausted 山札切れで勝者なしに終わったかを取得する
	IsRoundExhausted() bool
	// GetWinnerIdx 勝者の添字を取得する (-1: 未決着)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.DesmochePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.DesmochePlayer
}
