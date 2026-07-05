//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MachiavelliGame マキャヴェッリゲームインタフェース
type MachiavelliGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDraw 人間プレイヤーが山札から 1 枚引く（ターン終了）
	PlayerDraw() error
	// PlayerPlay 新しい場（メルド群）と追加する手札インデックスを提出する
	PlayerPlay(refs [][]domain.MachiavelliCardRef, handIndices []int) error
	// PlayerNewMeld 手札インデックスから新しいメルドを 1 つ場に出す
	PlayerNewMeld(handIndices []int) error
	// PlayerLayoff 手札 1 枚を既存メルドに追加する
	PlayerLayoff(meldIdx, handIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MachiavelliConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MachiavelliConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MachiavelliPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
	// GetTargetRounds ゲーム終了までのラウンド数
	GetTargetRounds() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックス
	GetDealerIdx() int
	// GetTable 共有テーブル上のメルド群
	GetTable() [][]*domain.Card
	// GetDrawPileCount 山札の残り枚数
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックス（-1 未確定）
	GetWinnerIdx() int
	// GetRoundWinnerIdx 直近ラウンドの勝者（-1 = 山切れ）
	GetRoundWinnerIdx() int
	// GetPlayerCnt プレイヤー数
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MachiavelliPlayer
	// PlayerDeadwoodValue プレイヤー i のデッドウッド点
	PlayerDeadwoodValue(i int) int
}
