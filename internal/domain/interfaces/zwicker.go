//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ZwickerGame ツヴィッカーゲームインタフェース
type ZwickerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Take 手札1枚を指定の値として使い、場札とビルドを取る
	Take(player, handIdx, playedValue int, tableIdxs, buildIdxs []int) error
	// Build 手札1枚と場札を積んで宣言値のビルドを作る
	Build(player, handIdx int, tableIdxs []int, declaredValue int) error
	// Trail 手札1枚を場に置いて手番を終える
	Trail(player, handIdx int) error
	// NextRound 次のディールを配る
	NextRound() error
	// ZwickerCpuDecide CPU が取る手を決める
	ZwickerCpuDecide(idx int) domain.ZwickerCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ZwickerConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ZwickerConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ZwickerPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetTableCards 場の単札を取得する
	GetTableCards() []*domain.Card
	// GetBuilds 場のビルドを取得する
	GetBuilds() []*domain.ZwickerBuild
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetTeamScore チームの累計得点を取得する
	GetTeamScore(team int) int
	// GetLastRoundScore 直近ディールの内訳を取得する (未精算なら nil)
	GetLastRoundScore() *domain.ZwickerRoundScore
	// GetWinnerTeam 勝ったチームを取得する (-1: 未決着)
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.ZwickerPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.ZwickerPlayer
}
