//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PishtiGame は Pişti (ピシュティ) のゲームインタフェース。
type PishtiGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム開始)
	Reset()
	// NextRound 次のゲームを開始する (Reset と同義)
	NextRound()
	// PlayerPlay 人間プレイヤーが手札を場へ出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.PishtiConfig)

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PishtiConfig
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PishtiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetPile 場の山を取得する (末尾が一番上)
	GetPile() []*domain.Card
	// GetPileTop 場の一番上の札を取得する (なければ nil)
	GetPileTop() *domain.Card
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを取得する (-1 = なし)
	GetLastCaptureIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PishtiPlayer
	// GetRemainingDeck 山札残り枚数を取得する
	GetRemainingDeck() int
	// GetWinners 勝者シートのリストを取得する (同点なら複数)
	GetWinners() []int
	// GetFinalScores 各プレイヤーの最終得点を取得する
	GetFinalScores() []int
}
