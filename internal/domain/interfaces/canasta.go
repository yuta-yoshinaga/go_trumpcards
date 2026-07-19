//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CanastaGame カナスタゲームインタフェース
type CanastaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札の山を取る
	PlayerDrawFromDiscard(naturalPairIndices []int) error
	// PlayerMeld プレイヤーがメルドを出す
	PlayerMeld(meldGroups [][]int) error
	// PlayerSkipMeld メルドフェーズをスキップする
	PlayerSkipMeld() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerGoOut プレイヤーが上がる
	PlayerGoOut() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CanastaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CanastaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CanastaPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDiscardPile 捨て札の山全体を古い順（下から上）に取得する
	GetDiscardPile() []*domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetDiscardPileCount 捨て札の枚数を取得する
	GetDiscardPileCount() int
	// GetPozzettoCount 残っているポゼット（予備手札）の山の数を取得する (Burraco モード)
	GetPozzettoCount() int
	// GetIsFrozen 捨て札の山がフリーズ状態かを返す
	GetIsFrozen() bool
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CanastaPlayer
	// GetDrewFromDiscard 捨て札から引いたかを返す
	GetDrewFromDiscard() bool
}
