package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PresidentGame プレジデントゲームインタフェース
type PresidentGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(indices []int) error
	// SuggestWeakestPlay playerIdx の最弱の合法手 (手札インデックス) を返す。合法手が無ければ nil
	SuggestWeakestPlay(playerIdx int) []int
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.PresidentConfig)

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PresidentPlayer
	// GetRevolutionActive 革命が有効かを返す
	GetRevolutionActive() bool
	// GetExchangeActions カード交換アクション一覧を取得する
	GetExchangeActions() []*domain.PresidentExchangeAction
	// GetTableCards 場のカード一覧を取得する
	GetTableCards() []*domain.Card
	// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックスを取得する
	GetLastPlayPlayerIdx() int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.PresidentCpuAction
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.PresidentCpuAction
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PresidentConfig
	// GetPassCount パス回数を取得する
	GetPassCount() int
}
