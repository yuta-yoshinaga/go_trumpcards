//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GinRummyGame ジンラミーゲームインタフェース
type GinRummyGame interface {
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
	PlayerKnock(cardIndex int) error
	// PlayerLayoff プレイヤーがレイオフする
	PlayerLayoff(cardIndices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GinRummyConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GinRummyConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GinRummyPhase
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
	GetPlayer(i int) *domain.GinRummyPlayer
	// GetKnockerIdx ノックしたプレイヤーインデックスを取得する
	GetKnockerIdx() int
	// GetKnockerMelds ノッカーのメルド一覧を取得する
	GetKnockerMelds() [][]*domain.Card
	// LayoffTargets その札を足せるノッカーのメルド番号を返す
	LayoffTargets(card *domain.Card) []int
	// GetKnockerDeadwood ノッカーのデッドウッドを取得する
	GetKnockerDeadwood() []*domain.Card
	// GetIsGin ジンかを返す
	GetIsGin() bool
}
