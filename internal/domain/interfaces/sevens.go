package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SevensGame 7並べゲームインタフェース
type SevensGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.SevensConfig)
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// HasAnyOption 指定プレイヤーが出せるカードまたはパスが可能かを返す
	HasAnyOption(playerIdx int) bool
	// AutoHandleNoOption 出せるカードもパスも不可の場合の自動処理
	AutoHandleNoOption()
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(idx int) error
	// PlayerPlayJoker プレイヤーがジョーカーを使って指定位置にカードを配置する
	PlayerPlayJoker(cardIdx, targetSuit, targetValue int) error
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SevensPlayer
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SevensConfig
	// GetTableMinVals 各スートの最小配置値を取得する
	GetTableMinVals() [5]int
	// GetTableMaxVals 各スートの最大配置値を取得する
	GetTableMaxVals() [5]int
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.SevensCpuAction
	// GetCpuActions CPU行動記録一覧を取得する
	GetCpuActions() []*domain.SevensCpuAction
	// GetTablePlaced 各スートの配置済みビットマスクを取得する
	GetTablePlaced() [5]uint16
	// GetPlayableCardIndices いま出せる人間の手札インデックスを取得する。
	// nil は「判定していない」、空スライスは「1枚も出せない」で意味が違う。
	GetPlayableCardIndices() []int
}
