//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// YanivGame Yaniv (ヤニブ) ゲームインタフェース
type YanivGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDiscard プレイヤーがカードの組を捨てる
	PlayerDiscard(cardIndices []int) error
	// PlayerDeclareYaniv プレイヤーが Yaniv を宣言する
	PlayerDeclareYaniv() error
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromPickup プレイヤーが直前の捨て札の端から引く
	PlayerDrawFromPickup(end int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.YanivConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.YanivConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.YanivPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 引ける捨て札の末尾を取得する
	GetDiscardTop() *domain.Card
	// GetPickupCards 引ける捨て札の束を取得する
	GetPickupCards() []*domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.YanivPlayer
	// GetCallerIdx Yaniv を宣言したプレイヤーインデックスを取得する
	GetCallerIdx() int
	// GetAsafWinnerIdx アサフで宣言者を下回ったプレイヤーインデックスを取得する
	GetAsafWinnerIdx() int
	// GetIsAsaf 直近の宣言がアサフだったかを取得する
	GetIsAsaf() bool
	// GetRoundScores 直近ラウンドで各プレイヤーが加算された失点を取得する
	GetRoundScores() []int
}
