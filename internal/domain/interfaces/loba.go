//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LobaGame ロバゲームインタフェース
type LobaGame interface {
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
	// Discard 手札1枚を捨てて手番を終える
	Discard(player, handIdx int) error
	// NextRound 次のラウンドを配る
	NextRound() error
	// LobaCpuDecide CPU が取る手を決める
	LobaCpuDecide(idx int) domain.LobaCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.LobaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.LobaConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.LobaPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDiscardTop 捨て札の一番上を取得する
	GetDiscardTop() *domain.Card
	// GetMelds 場のメルドを取得する
	GetMelds() []*domain.LobaMeld
	// HasMelded このラウンドで既に出したかを取得する
	HasMelded(idx int) bool
	// GetScore 累計失点を取得する
	GetScore(idx int) int
	// IsEliminated 脱落しているかを取得する
	IsEliminated(idx int) bool
	// GetRoundNumber 完了したラウンド数を取得する
	GetRoundNumber() int
	// GetRoundWinner 直近ラウンドで上がった人を取得する (-1: なし)
	GetRoundWinner() int
	// IsRoundClean 直近の上がりが「一気に」だったかを取得する
	IsRoundClean() bool
	// GetWinnerIdx 勝者の添字を取得する (-1: 未決着)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.LobaPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.LobaPlayer
}
