//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GermanWhistGame ジャーマンホイスト (German Whist) ゲームインタフェース
type GermanWhistGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GermanWhistPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit 切り札のスートを取得する
	GetTrumpSuit() int
	// GetUpCard 場に表向きで置かれている札を取得する (前半が終われば nil)
	GetUpCard() *domain.Card
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GermanWhistPlayer
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.GermanWhistHint
}
