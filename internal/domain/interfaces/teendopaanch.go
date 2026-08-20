//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TeenDoPaanchGame 3-2-5 (ティーン・ドー・パーンチ) ゲームインタフェース
type TeenDoPaanchGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerDeclareTrump 人間が切り札を宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPUが切り札を宣言する
	CpuDeclareTrump()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TeenDoPaanchConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TeenDoPaanchConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TeenDoPaanchPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanTrumpTurn 人間が切り札を宣言する番かを返す
	IsHumanTrumpTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (0: 未宣言)
	GetTrumpSuit() int
	// GetFivePlayerIdx ノルマ5を負う席 (切り札を決める席) を取得する
	GetFivePlayerIdx() int
	// GetLastExchange 直前のラウンド間で動いた札の枚数を取得する
	GetLastExchange() int
	// GetLastExchangePairs 直前のやり取りの内訳（誰から誰へ何枚）を取得する
	GetLastExchangePairs() []domain.TeenDoPaanchExchange
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TeenDoPaanchPlayer
	// GetWinnerIdx 勝者を取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.TeenDoPaanchHint
}
