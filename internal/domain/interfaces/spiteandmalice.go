package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpiteAndMaliceGame Spite & Malice ゲームインタフェース
type SpiteAndMaliceGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// PlayFromHand 手札のカードをファウンデーションに出す
	PlayFromHand(handIdx, foundationIdx int) error
	// PlayFromGoal ゴールパイルのトップをファウンデーションに出す
	PlayFromGoal(foundationIdx int) error
	// PlayFromSide サイドパイルのトップをファウンデーションに出す
	PlayFromSide(sideIdx, foundationIdx int) error
	// Discard 手札 1 枚をサイドパイルに捨ててターンを終了する
	Discard(handIdx, sideIdx int) error
	// CpuStep CPU の手番を 1 ステップ進める
	CpuStep() error
	// AutoComplete 自明な (foundation に出せる) 手を連続で適用する
	AutoComplete() error
	// CanAutoComplete AutoComplete が有効に働く状態か
	CanAutoComplete() bool
	// IsCpuTurn 現在のターンが CPU か
	IsCpuTurn() bool
	// GetHint 現在ターンの推奨手を返す
	GetHint() *domain.SpiteAndMaliceHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SpiteAndMalicePhase
	// GetCurrent 現在ターンのプレイヤーインデックスを取得する
	GetCurrent() int
	// GetMoveCount 操作回数を取得する
	GetMoveCount() int
	// GetWinner 勝者インデックス (-1 なら未決着) を取得する
	GetWinner() int
	// GetStockSize 山札残り枚数を取得する
	GetStockSize() int
	// GetCompletedSize 完成済み山の枚数を取得する
	GetCompletedSize() int
	// GetFoundations ファウンデーション一覧を取得する
	GetFoundations() [domain.SpiteAndMaliceFoundationCnt][]*domain.Card
	// GetFoundationTopValue ファウンデーションのトップ値を取得する
	GetFoundationTopValue(foundationIdx int) int
	// IsGoalTopPlayable ゴール札の一番上が今どれかの基礎札に出せるかを返す
	IsGoalTopPlayable(playerIdx int) bool
	// GetPlayer プレイヤー状態を取得する
	GetPlayer(idx int) *domain.SpiteAndMalicePlayer
	// GetConfig 設定を取得する
	GetConfig() domain.SpiteAndMaliceConfig
	// SetConfig 設定を更新する
	SetConfig(cfg domain.SpiteAndMaliceConfig)
}
