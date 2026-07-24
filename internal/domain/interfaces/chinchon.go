//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ChinchonGame チンチョンゲームインタフェース
type ChinchonGame interface {
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
	// PlayerKnock プレイヤーがノックする (1枚捨ててノック)
	PlayerKnock(cardIndex int) error
	// PlayerLayoff プレイヤーがノッカーのメルドにカードをレイオフする
	PlayerLayoff(cardIndices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ChinchonConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ChinchonConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ChinchonPhase
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
	// GetWinnerIdx マッチ勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ChinchonPlayer
	// GetPlayerDeadwoodValue プレイヤーの最善メルド分割でのデッドウッド点を取得する
	GetPlayerDeadwoodValue(i int) int
	// GetKnockThreshold ノック可能なデッドウッド点の上限を取得する
	GetKnockThreshold() int
	// GetKnockerIdx ノッカーのインデックスを取得する (-1 = ノックなし)
	GetKnockerIdx() int
	// GetKnockerMelds ノッカーのメルドを取得する
	GetKnockerMelds() [][]*domain.Card
}
