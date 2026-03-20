package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoubtGame ダウトゲームインタフェース
type DoubtGame interface {
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndices []int, claimedValue int, humanPlayMs int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveDoubt ダウト判定を実行する
	ResolveDoubt(doubterIndices []int)
	// SkipDoubt ダウトをスキップする
	SkipDoubt()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DoubtConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DoubtConfig)

	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.DoubtHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DoubtPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DoubtPlayer
	// GetTableCardCount 場のカード枚数を取得する
	GetTableCardCount() int
	// GetTableCards 場のカード一覧を取得する
	GetTableCards() []*domain.Card
	// GetLastAction 最後のアクション情報を取得する
	GetLastAction() *domain.DoubtAction
	// GetCpuDoubters CPUダウター一覧を取得する
	GetCpuDoubters() []int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.DoubtCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.DoubtCpuAction
	// GetLastDoubtResult 最後のダウト結果を取得する
	GetLastDoubtResult() *domain.DoubtDoubtResult
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
}
