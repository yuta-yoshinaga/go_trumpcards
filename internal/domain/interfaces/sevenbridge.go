//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SevenBridgeGame セブンブリッジゲームインタフェース
type SevenBridgeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerClaimPon プレイヤーがポンで捨て札を取得する
	PlayerClaimPon(cardIndices []int) error
	// PlayerClaimChi プレイヤーがチーで捨て札を取得する
	PlayerClaimChi(cardIndices []int) error
	// PlayerMeld プレイヤーがメルドを場に出す
	PlayerMeld(cardIndices []int) error
	// PlayerLayoff プレイヤーが既存メルドに 1 枚追加する
	PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// SuggestMeld playerIdx の最善メルド (手札インデックス) を返す。無ければ nil
	SuggestMeld(playerIdx int) []int
	// SuggestDiscard playerIdx の推奨ディスカード手札インデックスを返す。無ければ -1
	SuggestDiscard(playerIdx int) int
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SevenBridgeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SevenBridgeConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SevenBridgePhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
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
	GetPlayer(i int) *domain.SevenBridgePlayer
	// GetRoundWinnerIdx 直近ラウンドの勝者
	GetRoundWinnerIdx() int
	// GetClaimedThisTurn 直前ターンで claim されたか
	GetClaimedThisTurn() bool
}
