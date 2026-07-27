//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// EcarteGame エカルテ (Écarté) ゲームインタフェース
type EcarteGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerPropose elder が交換を提案する
	PlayerPropose() error
	// PlayerStand elder が交換せずに勝負する
	PlayerStand() error
	// PlayerRespond 親が提案に承諾 (accept=true) か拒否 (accept=false) する
	PlayerRespond(accept bool) error
	// PlayerDiscard 現在の手番プレイヤーが捨て札を選んで引き直す
	PlayerDiscard(indices []int) error
	// CpuExchange 現在の交換手番が CPU の場合に 1 アクション実行する
	CpuExchange()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する (プレイフェーズ)
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.EcarteConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.EcarteConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.EcartePhase
	// GetNegStep 現在の交換ステップを取得する
	GetNegStep() domain.EcarteNegStep
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 場に表向きで置かれている切り札表示カードを取得する (なければ nil)
	GetTrumpCard() *domain.Card
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetElderIdx 非親 (先手) インデックスを取得する
	GetElderIdx() int
	// IsRefusalByDealer 親が交換を拒否したかを返す
	IsRefusalByDealer() bool
	// GetDealPoints プレイヤーの当ディール得点を取得する
	GetDealPoints(i int) int
	// GetMatchScore プレイヤーの試合累積得点を取得する
	GetMatchScore(i int) int
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.EcartePlayer
	// GetStockRemaining 山札の残り枚数を取得する
	GetStockRemaining() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.EcarteHint
}
