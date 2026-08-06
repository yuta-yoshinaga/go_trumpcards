//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BarbuGame はバルブのゲームインタフェース。
type BarbuGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム開始)
	Reset()
	// NextDeal 次のディールを開始する
	NextDeal()
	// SelectContract ディーラーがコントラクトを選択する (trumpSuit は Trumps のみ)
	SelectContract(contract, trumpSuit int) error
	// PlayerPlay 手札を出す (Dominoes では handIdx == -1 でパス)
	PlayerPlay(handIdx int, tableIdxs []int) error
	// CpuPlay CPU プレイヤーが 1 ステップ実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.BarbuConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の意思決定者が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BarbuPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetDealerIdx 現在のディーラーを取得する
	GetDealerIdx() int
	// GetDealNumber 現在のディール番号 (0-indexed) を取得する
	GetDealNumber() int
	// GetCurrentContract 現在のコントラクト (-1 = 未選択) を取得する
	GetCurrentContract() int
	// GetTrumpSuit 切り札スート (-1 = なし) を取得する
	GetTrumpSuit() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentTrick 進行中のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetPlayableIndices いま出せる手札の位置 (フォロー義務を反映)
	GetPlayableIndices(playerIdx int) []int
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する (-1 = なし)
	GetLastTrickWinner() int
	// GetTablePlaced Dominoes の場の状態を取得する (index 1-4 = スート)
	GetTablePlaced() [5]uint16
	// GetUsedContracts 指定ディーラーが使用済みのコントラクトを取得する
	GetUsedContracts(dealerIdx int) [domain.BarbuContractCnt]bool
	// GetDominoPlayableIndices Dominoes でプレイ可能な手札インデックスを取得する
	GetDominoPlayableIndices(playerIdx int) []int
	// GetLastDealDetail 直前ディールの得点内訳を取得する
	GetLastDealDetail() *domain.BarbuDealDetail
	// GetDealHistory 完了した各ディールの得点内訳を古い順に取得する
	GetDealHistory() []*domain.BarbuDealDetail
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BarbuConfig
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// GetRoundWinners 勝者インデックス一覧を取得する
	GetRoundWinners() []int
}
