//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// Rummy500Game Rummy 500ゲームインタフェース
type Rummy500Game interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock 山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard 捨て札の指定インデックスから上のカードをすべて引く
	PlayerDrawFromDiscard(idx int) error
	// PlayerMeld 手札のカードを場に出してメルドを作る
	PlayerMeld(cardIndices []int) error
	// PlayerLayoff 既存メルドにカードを追加する
	PlayerLayoff(meldOwner, meldIdx, cardIndex int) error
	// PlayerDiscard カードを捨ててターンを終える
	PlayerDiscard(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.Rummy500Config
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.Rummy500Config)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.Rummy500Phase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardPile 捨て札の山を取得する
	GetDiscardPile() []*domain.Card
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.Rummy500Player
	// GetRoundEnderIdx ラウンドを終わらせたプレイヤーのインデックスを取得する
	GetRoundEnderIdx() int
}
