//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TarneebGame Tarneeb ゲームインタフェース
type TarneebGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする
	PlayerBid(bid int) error
	// CpuBid CPU プレイヤーがビッドする
	CpuBid()
	// PlayerDeclareTrump プレイヤーがトランプを宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPU プレイヤーがトランプを宣言する
	CpuDeclareTrump()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TarneebConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TarneebConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TarneebPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanTrumpTurn 現在のトランプ宣言手番が人間かを返す
	IsHumanTrumpTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する (0 = 未宣言)
	GetTrumpSuit() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetBidWinnerIdx ビッド勝者インデックスを取得する (-1 = 未確定)
	GetBidWinnerIdx() int
	// GetHighestBid 現在の最高ビッド値を取得する
	GetHighestBid() int
	// GetRedealCount 現ラウンドの再配布回数を取得する
	GetRedealCount() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetWinnerTeam 勝利チームを取得する (-1 = 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TarneebPlayer
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを取得する
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.TarneebHint
}
