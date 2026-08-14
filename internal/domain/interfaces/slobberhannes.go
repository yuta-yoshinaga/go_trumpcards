//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SlobberhannesGame スロバーハンネス (Slobberhannes) ゲームインタフェース
type SlobberhannesGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SlobberhannesConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SlobberhannesConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SlobberhannesPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SlobberhannesPlayer
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.SlobberhannesHint
}
