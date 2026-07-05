//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PanGame パングインゲ（Pan）ゲームインタフェース
type PanGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札トップからカードを引く
	PlayerDrawFromDiscard() error
	// PlayerMeld プレイヤーが手札のカードで新しいメルドを場に出す
	PlayerMeld(cardIndices []int) error
	// PlayerLayoff プレイヤーが既存メルドにカードをレイオフする
	PlayerLayoff(meldOwner, meldIdx, cardIndex int) error
	// PlayerDiscard プレイヤーがカードを捨ててターンを終える
	PlayerDiscard(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PanConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PanConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PanPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
	// GetTargetRounds ゲーム終了までのラウンド数
	GetTargetRounds() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックス
	GetDealerIdx() int
	// GetDiscardPile 捨て札の山
	GetDiscardPile() []*domain.Card
	// GetDiscardTop 捨て札の一番上のカード
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックス（-1 未確定）
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PanPlayer
	// GetPanDeclarerIdx 「パン」を宣言したプレイヤー（-1 = 宣言なし）
	GetPanDeclarerIdx() int
	// PlayerHandPoints プレイヤー i の手札ピップ点
	PlayerHandPoints(i int) int
	// PlayerMeldedCount プレイヤー i が場に出したカード枚数
	PlayerMeldedCount(i int) int
}
