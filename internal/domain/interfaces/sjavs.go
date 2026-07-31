//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SjavsGame シャウスゲームインタフェース
type SjavsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bid 切札スート長を申告する (0 はパス)
	Bid(player, length int) error
	// PlayCard 手札の札を出す
	PlayCard(player, handIdx int) error
	// NextHand 次のハンドを配る
	NextHand() error
	// SjavsCpuDecide CPU が取る手を決める
	SjavsCpuDecide(idx int) domain.SjavsCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SjavsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SjavsConfig)

	// GetGameEndFlag ラバーが決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SjavsPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx 親の添字を取得する
	GetDealerIdx() int
	// GetTrumpSuit 切札スートを取得する (-1: 未確定)
	GetTrumpSuit() int
	// GetBidderIdx 切札を宣言した席を取得する (-1: 未確定)
	GetBidderIdx() int
	// GetBidLength 確定したビッドの枚数を取得する
	GetBidLength() int
	// GetBids 席ごとの申告枚数を取得する
	GetBids() []int
	// LongestTrumpLength 指定プレイヤーの最長切札スート長を取得する
	LongestTrumpLength(player int) int
	// GetTrick 現在のトリックを取得する
	GetTrick() []domain.SjavsTrickCard
	// GetTrickNumber 完了したトリック数を取得する
	GetTrickNumber() int
	// GetValidPlayIndices 出せる手札の添字を取得する
	GetValidPlayIndices(player int) []int
	// GetTeamPoints チームの獲得点を取得する
	GetTeamPoints(team int) int
	// GetRemaining チームの 24 からの残りを取得する
	GetRemaining(team int) int
	// GetCrosses チームのラバー勝利数を取得する
	GetCrosses(team int) int
	// GetCarryOver 60-60 で持ち越された上乗せ点を取得する
	GetCarryOver() int
	// GetHandResult 直近のハンド精算を取得する (nil: 未精算)
	GetHandResult() *domain.SjavsHandResult
	// GetWinnerTeam 勝ったチームを取得する (-1: 未決着)
	GetWinnerTeam() int
	// IsDoubleVictory 相手が 24 のままの勝利かを取得する
	IsDoubleVictory() bool
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.SjavsPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.SjavsPlayer
}
