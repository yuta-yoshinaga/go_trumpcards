//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// QuodlibetGame はクオドリベットのゲームインタフェース。
type QuodlibetGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextDeal 次のディールを開始する
	NextDeal()
	// SelectContract ディーラーがコントラクトを選択する
	SelectContract(contract int) error
	// CpuSelectContract CPU のディーラーがコントラクトを選ぶ
	CpuSelectContract()
	// PlayerPlay 人間が 1 手指す (シェディング系では handIdx == -1 でパス)
	PlayerPlay(handIdx int) error
	// CpuPlay CPU が 1 手指す
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.QuodlibetConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.QuodlibetConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間が決める番かを取得する
	IsHumanTurn() bool
	// GetDealNumber 現在のディール index (0 始まり) を取得する
	GetDealNumber() int
	// GetRoundNumber 現在の輪 (1-3) を取得する
	GetRoundNumber() int
	// GetDealerIdx ディーラー (Bierkönig) を取得する
	GetDealerIdx() int
	// GetCurrentContract 現在のコントラクト (-1 = 未選択) を取得する
	GetCurrentContract() int
	// GetAvailableContracts この輪でまだ打たれていないコントラクトを取得する
	GetAvailableContracts() []int
	// GetUsedContracts 消化済みコントラクトを取得する
	GetUsedContracts() [domain.QuodlibetContractCnt]bool
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentTurn 手番の席を取得する
	GetCurrentTurn() int
	// GetLeadPlayerIdx リードの席を取得する
	GetLeadPlayerIdx() int
	// GetCurrentTrick 進行中のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する (-1 = なし)
	GetLastTrickWinner() int
	// GetPlayableIndices いま出せる手札の位置を取得する
	GetPlayableIndices(playerIdx int) []int
	// GetSheddingPlayableIndices シェディング系で出せる手札の位置を取得する
	GetSheddingPlayableIndices(playerIdx int) []int
	// GetTablePlaced 小食いの場を取得する (index 1-4 = スート)
	GetTablePlaced() [5]uint16
	// GetStack 四分の現在の重ねを取得する
	GetStack() []*domain.Card
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.QuodlibetPlayer
	// GetLastDealDetail 直前ディールの罰点内訳を取得する
	GetLastDealDetail() *domain.QuodlibetDealDetail
	// GetDealHistory 完了した各ディールの罰点内訳を古い順に取得する
	GetDealHistory() []*domain.QuodlibetDealDetail
	// GetWinners 罰点が最も少ない席を取得する
	GetWinners() []int
	// GetHint ヒントを取得する
	GetHint() *domain.QuodlibetHint
}
