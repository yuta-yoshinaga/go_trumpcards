//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// EstimationGame エスティメーション (Estimation) ゲームインタフェース
type EstimationGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SelectTrump 親が切り札スートを決める
	SelectTrump(suit int) error
	// CpuSelectTrump CPU の親が切り札スートを決める
	CpuSelectTrump()
	// PlayerBid プレイヤーが獲得予定トリック数を宣言する
	PlayerBid(bid int) error
	// CpuBid 手番の CPU が宣言する
	CpuBid()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.EstimationConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.EstimationConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.EstimationPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 人間が宣言する番かを返す
	IsHumanBidTurn() bool
	// IsHumanTrumpTurn 人間が切り札を決める番かを返す
	IsHumanTrumpTurn() bool
	// GetRestrictedBid 最後の宣言者が選べない宣言値を返す (-1: 制限なし)
	GetRestrictedBid() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (未決定は 0)
	GetTrumpSuit() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx 宣言の手番を取得する
	GetBidPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.EstimationPlayer
	// GetWinnerIdx 勝利プレイヤーを取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.EstimationHint
}
