//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HachiHachiGame は八八 (Hachi-Hachi) のゲームインタフェース。
type HachiHachiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay 人間が手札を出す (fieldIdx で 2 枚一致時の捕獲対象を指定; 不要なら -1)
	PlayerPlay(handIdx, fieldIdx int) error
	// CpuPlay CPU のプレイ手番を 1 ステップ実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HachiHachiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HachiHachiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HachiHachiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetFieldCards 場札を取得する
	GetFieldCards() []*domain.Card
	// GetRemainingDeck 山札の残り枚数を取得する
	GetRemainingDeck() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetLastRoundResult 直近ラウンド結果を取得する
	GetLastRoundResult() *domain.HachiHachiRoundResult
	// GetWinner 終局時の勝者を取得する (-1=引き分け/未決)
	GetWinner() int
	// GetResult 人間視点のゲーム結果を取得する
	GetResult() domain.HachiHachiResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HachiHachiPlayer
	// GetYaku 指定プレイヤーの現在の成立出来役と素点合計を取得する
	GetYaku(playerIdx int) ([]domain.HachiHachiYaku, int)
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetCaptureOptions 各手札が捕獲できる場札インデックスを取得する
	GetCaptureOptions(playerIdx int) map[int][]int
	// GetHint ヒントを取得する
	GetHint() *domain.HachiHachiHint
}
