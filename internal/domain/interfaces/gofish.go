package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GoFishGame Go Fishゲームインタフェース
type GoFishGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.GoFishConfig)
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerAsk 人間プレイヤーが相手にランクを要求する
	PlayerAsk(targetIdx, rank int) error
	// CpuAsk CPUプレイヤーが1回要求を実行する
	CpuAsk() error

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GoFishPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetPhase ゲームフェーズを取得する
	GetPhase() domain.GoFishPhase
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GoFishConfig
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する
	GetWinnerIdx() int
	// GetTurnNumber 現在のターン番号を取得する
	GetTurnNumber() int
	// GetDeckRemaining 山札の残り枚数を取得する
	GetDeckRemaining() int

	// GetLastAskPlayerIdx 最後に要求したプレイヤーのインデックスを取得する
	GetLastAskPlayerIdx() int
	// GetLastAskTargetIdx 最後に要求された相手のインデックスを取得する
	GetLastAskTargetIdx() int
	// GetLastAskRank 最後に要求されたランクを取得する
	GetLastAskRank() int
	// GetLastAskSuccess 最後の要求が成功したかを返す
	GetLastAskSuccess() bool
	// GetLastCardsReceived 最後に受け取ったカードを取得する
	GetLastCardsReceived() []*domain.Card
	// GetLastDrawnCard 最後にGo Fishで引いたカードを取得する
	GetLastDrawnCard() *domain.Card
	// GetLastBookFormed 最後のアクションでブックが完成したかを返す
	GetLastBookFormed() bool
	// GetLastBookRank 最後に完成したブックのランクを取得する
	GetLastBookRank() int
	// GetCpuActions CPUターンの行動履歴を取得する
	GetCpuActions() []*domain.GoFishCpuAction
	// GetKnownRanks は各プレイヤーが保有を公開済みのランク一覧を返す
	GetKnownRanks() map[int][]int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.GoFishCpuAction
}
