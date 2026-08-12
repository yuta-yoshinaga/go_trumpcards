//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LingerLongerGame リンガーロンガーゲームインタフェース
type LingerLongerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.LingerLongerConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.LingerLongerConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.LingerLongerPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetStockSize 山札の残り枚数を取得する
	GetStockSize() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リード席を取得する
	GetLeadPlayerIdx() int
	// GetTrickNumber 解決済みのトリック数を取得する
	GetTrickNumber() int
	// GetLastDrawIdx 直前に補充した席を取得する (-1: 無し)
	GetLastDrawIdx() int
	// GetEliminatedCnt 脱落した人数を取得する
	GetEliminatedCnt() int
	// GetDiscarded 場から抜けた札の枚数を取得する
	GetDiscarded() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.LingerLongerPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.LingerLongerHint
}
