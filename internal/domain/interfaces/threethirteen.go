//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ThreeThirteenGame スリー・サーティーンゲームインタフェース
type ThreeThirteenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドへ進む（11 ラウンド消化後はゲーム終了を確定）
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札トップからカードを引く
	PlayerDrawFromDiscard() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerKnock プレイヤーがカードを捨ててノックする
	PlayerKnock(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ThreeThirteenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ThreeThirteenConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ThreeThirteenPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRound 現在のラウンド番号（1..11）
	GetRound() int
	// WildRank そのラウンドのワイルドランク
	WildRank() int
	// GetDealCount そのラウンドの 1 人あたり配布枚数
	GetDealCount() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetKnockerIdx ノックしたプレイヤー（-1 = 未ノック）
	GetKnockerIdx() int
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
	GetPlayer(i int) *domain.ThreeThirteenPlayer
	// GetPlayerDeadwoodValue プレイヤーの最善メルド分割でのデッドウッド点
	GetPlayerDeadwoodValue(i int) int
	// GetDeadwoodAfterDiscard その札を捨てた後のデッドウッド点 (範囲外は -1)
	GetDeadwoodAfterDiscard(playerIdx, cardIndex int) int
}
