//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KingGame はキング (King) のゲームインタフェース。
type KingGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextDeal 次のディールを開始する
	NextDeal()
	// SelectContract 親がコントラクトを選択する (trumpSuit は King Trump のみ)
	SelectContract(contract, trumpSuit int) error
	// PlayerPlay 人間プレイヤーがカードを出す
	PlayerPlay(handIdx int) error
	// CpuPlay CPUプレイヤーが1ステップ実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KingConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KingConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetDealNumber 現在のディール番号 (0-indexed) を取得する
	GetDealNumber() int
	// GetDealerIdx 親インデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetCurrentContract 現在のコントラクトを取得する (-1=未選択)
	GetCurrentContract() int
	// GetTrumpSuit 切り札スートを取得する (-1=なし)
	GetTrumpSuit() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する (-1=なし)
	GetLastTrickWinner() int
	// GetUsedContracts 使用済みコントラクトを取得する
	GetUsedContracts() [domain.KingContractCnt]bool
	// GetLastDealDetail 直前ディールの得点内訳を取得する
	GetLastDealDetail() *domain.KingDealDetail
	// GetRoundWinners ゲーム終了時の最高得点プレイヤーを取得する
	GetRoundWinners() []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.KingPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.KingHint
}
