//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GuandanGame 掼蛋 (Guandan) ゲームインタフェース
type GuandanGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// PlayCards 手札から役を出す
	PlayCards(player int, idxs []int) error
	// Pass パスする
	Pass(player int) error
	// ReturnTribute 還貢する
	ReturnTribute(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GuandanConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GuandanConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GuandanPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetLevel この局の基準レベルを取得する
	GetLevel() int
	// GetTeamLevel チームの現在レベルを取得する
	GetTeamLevel(team int) int
	// GetDeclarerTeam この局のレベルを提供しているチームを取得する
	GetDeclarerTeam() int
	// GetLastCombo 場に出ている最後の役を取得する
	GetLastCombo() *domain.GuandanCombo
	// GetLastPlayerIdx 最後に出した席を取得する
	GetLastPlayerIdx() int
	// GetFinished 上がった順の席を取得する
	GetFinished() []int
	// GetTributes この局の進貢を取得する
	GetTributes() []*domain.GuandanTribute
	// IsTributeCancelled 赤ジョーカー保持で貢が取り消されたかを返す
	IsTributeCancelled() bool
	// GetLastResult 直前の局の結果を取得する
	GetLastResult() *domain.GuandanHandResult
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.GuandanPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.GuandanPlayer
}
