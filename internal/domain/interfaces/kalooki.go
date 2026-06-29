//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KalookiGame カルーキゲームインタフェース
type KalookiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドへ進む（Kalooki は 1 ラウンド完結のためゲーム終了を確定）
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札トップからカードを引く
	PlayerDrawFromDiscard() error
	// PlayerMeld プレイヤーがメルド群を場に出す（オープニング要件チェックを含む）
	PlayerMeld(meldGroups [][]int) error
	// PlayerLayoff プレイヤーが既存メルドに 1 枚追加する
	PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KalookiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KalookiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KalookiPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetOpeningThreshold オープニング要件のしきい値
	GetOpeningThreshold() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカード
	GetDiscardTop() *domain.Card
	// GetDiscardPile 捨て札の山
	GetDiscardPile() []*domain.Card
	// GetDrawPileCount 山札の残り枚数
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックス（-1 未確定）
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.KalookiPlayer
	// GetRoundWinnerIdx 直近ラウンドの勝者
	GetRoundWinnerIdx() int
}
