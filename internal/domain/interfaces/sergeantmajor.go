//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SergeantMajorGame サージェントメジャー (8-5-3) ゲームインタフェース
type SergeantMajorGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerDeclareTrump 人間（親）が切り札を宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPUの親が切り札を宣言する
	CpuDeclareTrump()
	// PlayerDiscard 人間（親）がキティのぶんを捨てる
	PlayerDiscard(indices []int) error
	// CpuDiscard CPUの親がキティのぶんを捨てる
	CpuDiscard()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SergeantMajorConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SergeantMajorConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SergeantMajorPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanTrumpTurn 人間が切り札を宣言する番かを返す
	IsHumanTrumpTurn() bool
	// IsHumanDiscardTurn 人間がキティを捨てる番かを返す
	IsHumanDiscardTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (0: 未宣言)
	GetTrumpSuit() int
	// GetKittySize キティの枚数を取得する
	GetKittySize() int
	// GetDiscardCount 親が捨てる枚数を取得する
	GetDiscardCount() int
	// GetLastExchange 直前のラウンド間で動いた札の枚数を取得する
	GetLastExchange() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する (この席がノルマ8)
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SergeantMajorPlayer
	// GetWinnerIdx 勝者を取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.SergeantMajorHint
}
