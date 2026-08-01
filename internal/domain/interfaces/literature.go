//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LiteratureGame リテラチャー (Literature) ゲームインタフェース
type LiteratureGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Ask 札を要求する
	Ask(from, to int, c *domain.Card) error
	// Claim ハーフスートを宣言する
	Claim(player, half int, holders []int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// LiteratureCanAsk 要求が成立するかを返す
	LiteratureCanAsk(from, to int, c *domain.Card) error
	// LiteratureHoldsHalfSuit 席がそのハーフスートを1枚以上持つかを返す
	LiteratureHoldsHalfSuit(seat, half int) bool
	// LiteratureTeamHalfSuits チームが取ったハーフスート数を返す
	LiteratureTeamHalfSuits(team int) int
	// LiteratureCancelledCount 無効になったハーフスート数を返す
	LiteratureCancelledCount() int
	// LiteratureOpenCount まだ決着していないハーフスート数を返す
	LiteratureOpenCount() int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.LiteratureConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.LiteratureConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.LiteraturePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetHalfSuitState ハーフスートの帰属を取得する
	GetHalfSuitState(half int) domain.LiteratureHalfSuitState
	// GetAsks 要求の履歴を取得する (公開情報)
	GetAsks() []*domain.LiteratureAsk
	// GetClaims 宣言の履歴を取得する
	GetClaims() []*domain.LiteratureClaimResult
	// GetLastAsk 直前の要求を取得する
	GetLastAsk() *domain.LiteratureAsk
	// GetLastClaim 直前の宣言を取得する
	GetLastClaim() *domain.LiteratureClaimResult
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.LiteraturePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.LiteraturePlayer
}
