//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpoonsGame はスプーン (Spoons) のゲームインタフェース。
type SpoonsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 設定を更新してゲームを初期化する
	ResetWithConfig(cfg domain.SpoonsConfig)
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPass 人間プレイヤーが手札の 1 枚を次へ渡す
	PlayerPass(cardIndex int) error
	// PlayerGrabSpoon 人間プレイヤーがスプーンを掴む
	PlayerGrabSpoon() error
	// CpuPlay CPU の手番を 1 ステップ進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SpoonsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SpoonsConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SpoonsPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SpoonsPlayer
	// GetWinnerIdx 勝者インデックスを取得する (-1=未確定)
	GetWinnerIdx() int
	// GetSpoonsRemaining テーブル上に残るスプーン本数を取得する
	GetSpoonsRemaining() int
	// GetCurrentPlayerIdx 現在パスする番のプレイヤーを取得する
	GetCurrentPlayerIdx() int
	// GetFeederIdx 配り手プレイヤーを取得する
	GetFeederIdx() int
	// GetDrawPileSize ドローパイル残枚数を取得する
	GetDrawPileSize() int
	// GetPassedCard 現在流れているカードを取得する
	GetPassedCard() *domain.Card
	// IsGrabWindowOpen グラブウィンドウが開いているかを取得する
	IsGrabWindowOpen() bool
	// GetFirstGrabberIdx 最初にスプーンを掴んだプレイヤーを取得する (-1=未)
	GetFirstGrabberIdx() int
	// GetRoundLoserIdx 直近ラウンドで文字が付いたプレイヤーを取得する (-1=未)
	GetRoundLoserIdx() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
}
