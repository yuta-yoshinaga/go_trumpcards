//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ThirtyOneGame ThirtyOne (サーティワン / Scat) ゲームインタフェース
type ThirtyOneGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札からカードを引く
	PlayerDrawFromDiscard() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerKnock プレイヤーがノックする
	PlayerKnock() error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ThirtyOneConfig
	// GetCpuKnockThreshold 現在の難易度で CPU がノックを検討する合計値を取得する
	GetCpuKnockThreshold() int
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ThirtyOneConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ThirtyOnePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ThirtyOnePlayer
	// GetKnockerIdx ノックしたプレイヤーインデックスを取得する
	GetKnockerIdx() int
	// GetHint 現在の局面での推奨手を取得する (人間の手番でなければ nil)
	GetHint() *domain.ThirtyOneHint
	// GetThirtyOneIdx 31を達成したプレイヤーインデックスを取得する
	GetThirtyOneIdx() int
	// GetRoundWinnerIdx 直近ラウンドの勝者インデックスを取得する
	GetRoundWinnerIdx() int
	// GetRoundLosers 直近ラウンドでライフを失ったプレイヤーを取得する
	GetRoundLosers() []int
}
