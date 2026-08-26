//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ShengJiGame 升级 (Sheng Ji) ゲームインタフェース
type ShengJiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// Declare 亮牌する (ShengJiNoTrump はパス)
	Declare(seat, suit int) error
	// BuryKitty 底牌に8枚埋め戻す
	BuryKitty(seat int, idxs []int) error
	// Play 手を出す
	Play(seat int, idxs []int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// ShengJiDeclareStrength 席がそのスートで出せる亮牌の強さを返す
	ShengJiDeclareStrength(seat, suit int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ShengJiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ShengJiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ShengJiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetLevel この局の基準レベルを取得する
	GetLevel() int
	// GetTeamLevel チームの現在レベルを取得する
	GetTeamLevel(team int) int
	// GetDeclarerTeam この局の宣言側を取得する
	GetDeclarerTeam() int
	// GetTrumpSuit 切札スートを取得する
	GetTrumpSuit() int
	// GetDeclaration 成立している亮牌を取得する
	GetDeclaration() *domain.ShengJiDeclaration
	// GetKittySize 底牌の枚数を取得する
	GetKittySize() int
	// GetKitty 底牌を取得する (終局まで nil)
	GetKitty() []*domain.Card
	// GetTrick いまのトリックに出された手を取得する
	GetTrick() [][]*domain.Card
	// GetTrickLeader いまのトリックのリード席を取得する
	GetTrickLeader() int
	// GetLeadCombo リードされた手の形を取得する
	GetLeadCombo() *domain.ShengJiCombo
	// GetTeamPoints チームが集めた点を取得する
	GetTeamPoints(team int) int
	// GetTrickCount 消化したトリック数を取得する
	GetTrickCount() int
	// GetLastTrickWinner 直前のトリックの勝者席を取得する
	GetLastTrickWinner() int
	// GetLastResult 直前の局の結果を取得する
	GetLastResult() *domain.ShengJiHandResult
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.ShengJiPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.ShengJiPlayer
}
