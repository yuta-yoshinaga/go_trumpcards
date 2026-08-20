//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PasurGame パスールゲームインタフェース
type PasurGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す (tableIndices が空ならトレール)
	PlayerPlay(cardIndex int, tableIndices []int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PasurConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PasurConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PasurPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetTableCards 場の札を取得する
	GetTableCards() []*domain.Card
	// GetCaptureOptions 指定手札で取れる場札の組み合わせを返す
	GetCaptureOptions(playerIdx, cardIndex int) [][]int
	// GetDeckRemaining 山札の残り枚数を取得する
	GetDeckRemaining() int
	// GetPacksDealt 配ったパック数を取得する
	GetPacksDealt() int
	// GetLastCaptureIdx 最後に捕獲した席を取得する (-1: なし)
	GetLastCaptureIdx() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PasurPlayer
	// GetScore 席の得点を取得する
	GetScore(i int) int
	// GetWinners 勝った席を取得する (同点なら複数、終局前は空)
	GetWinners() []int
	// GetHint ヒントを取得する
	GetHint() *domain.PasurHint
}
