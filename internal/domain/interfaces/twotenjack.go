package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TwoTenJackGame ツーテンジャックゲームインタフェース
type TwoTenJackGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDeclareTrump プレイヤーがトランプスートを宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPUプレイヤーがトランプスートを宣言する
	CpuDeclareTrump()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TwoTenJackConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TwoTenJackConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TwoTenJackPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanDeclareTurn 現在の宣言手番が人間かを返す
	IsHumanDeclareTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDeclarerIdx 宣言者インデックスを取得する
	GetDeclarerIdx() int
	// GetTrumpSuit トランプスートを取得する (-1 = 未宣言)
	GetTrumpSuit() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TwoTenJackPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.TwoTenJackHint
}
