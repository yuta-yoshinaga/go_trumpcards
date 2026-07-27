//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CinchGame はチンチ (Cinch / Double Pedro / High Five) のゲームインタフェース。
type CinchGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間がビッドする (0=pass, 1..CinchMaxBid)
	PlayerBid(bid int) error
	// CpuBid CPUプレイヤーが1回ビッドする
	CpuBid()
	// NameTrump 人間のビッド勝者が切り札スートを宣言する
	NameTrump(suit int) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ステップ実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CinchConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CinchConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CinchPhase
	// IsHumanTurn 現在の意思決定者が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetDealerIdx 親インデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する (-1=なし)
	GetLastTrickWinner() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx 現在ビッド中のプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetCurrentBid 現在の最高ビッド値を取得する
	GetCurrentBid() int
	// GetBidWinnerIdx 最高ビッダーのインデックスを取得する (-1=未確定)
	GetBidWinnerIdx() int
	// GetTrumpSuit 切り札スートを取得する (0=未確定)
	GetTrumpSuit() int
	// GetWinnerIdx 勝者インデックスを取得する (-1=未確定)
	GetWinnerIdx() int
	// GetLastDealDetail 直前ディールの得点内訳を取得する
	GetLastDealDetail() *domain.CinchDealDetail
	// GetRoundWinners ゲーム終了時の最高得点プレイヤーを取得する
	GetRoundWinners() []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CinchPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.CinchHint
}
