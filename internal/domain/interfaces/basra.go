//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BasraGame はバスラ (Basra / Bastra) のゲームインタフェース。
type BasraGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のゲームを開始する (Reset と同義)
	NextRound()
	// PlayerPlay 人間が手札を出す (tableIdxs で捕獲対象を指定; 空ならトレイル/ジャック一掃)
	PlayerPlay(handIdx int, tableIdxs []int) error
	// CpuPlay CPUプレイヤーが1ステップ実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BasraConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BasraConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BasraPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetTableCards 場の札を取得する
	GetTableCards() []*domain.Card
	// GetLastCaptureIdx 最後に捕獲したプレイヤーを取得する (-1=なし)
	GetLastCaptureIdx() int
	// GetRemainingDeck 山札の残り枚数を取得する
	GetRemainingDeck() int
	// GetRoundNumber 配布パック数 (配り直し回数) を取得する
	GetRoundNumber() int
	// GetWinners 勝者リストを取得する
	GetWinners() []int
	// GetLastDealDetail 直前ゲームの得点内訳を取得する
	GetLastDealDetail() *domain.BasraScoreDetail
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BasraPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetCaptureOptions 各手札が捕獲できる場札インデックスを取得する
	GetCaptureOptions(playerIdx int) map[int][]int
	// GetHint ヒントを取得する
	GetHint() *domain.BasraHint
}
