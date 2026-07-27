//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SchnapsenGame シュナプセン (Schnapsen / Sixty-Six) ゲームインタフェース
type SchnapsenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerDeclareMarriage プレイヤーがマリアージュを宣言してその K/Q をリードする
	PlayerDeclareMarriage(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する (補充ドロー + ゲーム終了検出を含む)
	NextTrick()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SchnapsenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SchnapsenConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SchnapsenPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsEndgame 第2フェーズ (マストフォロー) かを返す
	IsEndgame() bool
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 場に表向きで置かれている切り札表示カードを取得する (山札に残っていなければ nil)
	GetTrumpCard() *domain.Card
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetPlayerPoints プレイヤーの累積得点を取得する
	GetPlayerPoints(i int) int
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SchnapsenPlayer
	// GetStockRemaining 山札の残り枚数を取得する
	GetStockRemaining() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetMarriageIndices マリアージュ宣言を開始できるカードのインデックスを返す
	GetMarriageIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.SchnapsenHint
}
