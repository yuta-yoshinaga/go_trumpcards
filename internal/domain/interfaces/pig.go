//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PigGame ピッグゲームインタフェース
type PigGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPass 人間が渡す札を選ぶ
	PlayerPass(cardIndex int) error
	// PlayerSignal 人間が合図に気づいたことを伝える
	PlayerSignal() error
	// NextRound ラウンド終了状態から次を配る
	NextRound() error
	// CpuPlay CPUの手を1つ進める
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PigConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PigConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PigPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetValidPassIndices 渡せる手札のインデックスを返す
	GetValidPassIndices(playerIdx int) []int
	// HasChosenPass 席iが渡す札を選び終えたかを返す
	HasChosenPass(i int) bool
	// GetCurrentPlayerIdx 渡す札を選ぶ番の席を返す
	GetCurrentPlayerIdx() int
	// GetSignallerIdx 最初に合図した席を返す (-1: 合図なし)
	GetSignallerIdx() int
	// GetNoticedCnt 合図に気づいた人数を返す
	GetNoticedCnt() int
	// GetRoundLoserIdx 直近ラウンドで文字が付いた席を返す (-1: 未)
	GetRoundLoserIdx() int
	// GetRoundNumber ラウンド数を返す
	GetRoundNumber() int
	// GetPassCount 当該ラウンドのパス回数を返す
	GetPassCount() int
	// GetDeckSize この卓で使うデッキ枚数を返す
	GetDeckSize() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PigPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.PigHint
}
