package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ContractRummyGame コントラクトラミーゲームインタフェース
type ContractRummyGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札トップからカードを引く
	PlayerDrawFromDiscard() error
	// PlayerMeldContract プレイヤーがコントラクトを達成するメルドを場に出す
	PlayerMeldContract(indicesPerSlot [][]int) error
	// PlayerMeldExtra プレイヤーが追加メルドを場に出す
	PlayerMeldExtra(indices []int) error
	// PlayerLayoff プレイヤーが既存メルドに 1 枚追加する
	PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ContractRummyConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ContractRummyConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ContractRummyPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
	// GetCurrentContract 現在のコントラクトを取得
	GetCurrentContract() domain.Contract
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
	GetPlayer(i int) *domain.ContractRummyPlayer
	// GetRoundWinnerIdx 直近ラウンドの勝者
	GetRoundWinnerIdx() int
}
