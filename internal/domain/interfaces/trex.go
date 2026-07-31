//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TrexGame トリックスゲームインタフェース
type TrexGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ChooseContract 王が契約を選ぶ
	ChooseContract(player int, contract domain.TrexContract) error
	// PlayCard 手札の札を出す
	PlayCard(player, handIdx int) error
	// Pass ドミノで出せないときに手番を渡す
	Pass(player int) error
	// NextDeal 次のディールを配る
	NextDeal() error
	// TrexCpuDecide CPU が取る手を決める
	TrexCpuDecide(idx int) domain.TrexCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TrexConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TrexConfig)

	// GetGameEndFlag 20 ディールを終えたかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TrexPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetKingIdx 王の添字を取得する
	GetKingIdx() int
	// GetContract 現在の契約を取得する
	GetContract() domain.TrexContract
	// AvailableContracts 王がまだ選んでいない契約を取得する
	AvailableContracts() []domain.TrexContract
	// IsContractUsed 指定の王が契約を消化済みかを取得する
	IsContractUsed(king int, contract domain.TrexContract) bool
	// IsTrix 現在の契約がドミノかを取得する
	IsTrix() bool
	// GetDealNumber 完了したディール数を取得する
	GetDealNumber() int
	// GetTrick 現在のトリックを取得する
	GetTrick() []domain.TrexTrickCard
	// GetTrickNumber 完了したトリック数を取得する
	GetTrickNumber() int
	// GetTricksWon 指定プレイヤーのトリック数を取得する
	GetTricksWon(idx int) int
	// GetValidPlayIndices 出せる手札の添字を取得する
	GetValidPlayIndices(player int) []int
	// GetSuitRun 指定スートのドミノの伸びを取得する
	GetSuitRun(suit int) (bool, int, int)
	// GetFinishOrder ドミノの上がり順を取得する
	GetFinishOrder() []int
	// GetScore 累計得点を取得する
	GetScore(idx int) int
	// GetDealScore 今ディールの得点を取得する
	GetDealScore(idx int) int
	// GetWinnerIdx 最終得点が最も高い席を取得する (-1: 未終局)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.TrexPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.TrexPlayer
}
