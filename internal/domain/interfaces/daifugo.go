package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DaifugoGame 大富豪ゲームインタフェース
type DaifugoGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(indices []int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// HasPendingAction ペンディングアクションがあるかを返す
	HasPendingAction() bool
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.DaifugoConfig)
	// SortHumanHand 人間の手札を指定モードでソートする
	SortHumanHand(mode domain.DaifugoSortMode) error

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DaifugoPlayer
	// GetRevolutionActive 革命が有効かを返す
	GetRevolutionActive() bool
	// GetElevenBackActive 11バックが有効かを返す
	GetElevenBackActive() bool
	// GetSuitLocked スート縛りが有効かを返す
	GetSuitLocked() bool
	// GetLockedSuit 縛りスートを取得する
	GetLockedSuit() int
	// GetTableIsSequence 場が階段かを返す
	GetTableIsSequence() bool
	// GetExchangeActions カード交換アクション一覧を取得する
	GetExchangeActions() []*domain.DaifugoExchangeAction
	// GetTableCards 場のカード一覧を取得する
	GetTableCards() []*domain.Card
	// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックスを取得する
	GetLastPlayPlayerIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.DaifugoCpuAction
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.DaifugoCpuAction
	// GetPlayableCardIndices いま出せる組み合わせに含まれる手札インデックスを取得する
	GetPlayableCardIndices() []int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DaifugoConfig
	// GetPassCount パス回数を取得する
	GetPassCount() int
	// GetPendingActionType ペンディングアクションの種類を取得する
	GetPendingActionType() domain.DaifugoPendingAction
	// GetPendingActionTarget ペンディングアクション対象プレイヤーを取得する
	GetPendingActionTarget() int
	// GetReverseDirection 逆回りかを返す
	GetReverseDirection() bool
	// GetNumberLocked 数字縛りが有効かを返す
	GetNumberLocked() bool
	// GetSequenceLocked 階段縛りが有効かを返す
	GetSequenceLocked() bool
	// GetSortMode 現在のソートモードを取得する
	GetSortMode() domain.DaifugoSortMode
}
