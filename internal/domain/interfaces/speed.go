package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpeedGame スピードゲームインタフェース
type SpeedGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay 人間プレイヤーがカードを出す
	PlayerPlay(cardIndex, pileIndex int) error
	// CpuPlay CPUがカードを出す
	CpuPlay() []*domain.SpeedCpuAction
	// Flip 膠着時に台札をめくる
	Flip() error
	// UpdatePhase フェーズを再計算する
	UpdatePhase()

	// GetHint ヒントを返す
	GetHint() (cardIdx, pileIdx int, found bool)

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SpeedConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SpeedConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SpeedPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsStuck 膠着状態かを返す
	IsStuck() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SpeedPlayer
	// GetCenterPile 指定インデックスの台札を取得する
	GetCenterPile(i int) *domain.Card
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// CanPlay 指定プレイヤーの手札カードが指定台札に出せるか
	CanPlay(playerIdx, cardIdx, pileIdx int) bool
	// PlayerHasAnyPlay 指定プレイヤーに出せる手があるか
	PlayerHasAnyPlay(playerIdx int) bool
}
