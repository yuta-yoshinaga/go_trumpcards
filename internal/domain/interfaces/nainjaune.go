//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NainJauneGame ル・ナン・ジョーヌゲームインタフェース
type NainJauneGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Play 手札1枚を出す
	Play(player, handIdx int) error
	// NextDeal 次のディールを配る
	NextDeal() error
	// NainJauneCpuDecide CPU が出す手札の添字を返す (-1: 出せない)
	NainJauneCpuDecide(idx int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.NainJauneConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.NainJauneConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.NainJaunePhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetBoard 5区画の残高を取得する
	GetBoard() domain.NainJauneBoard
	// GetTalonCount talon の枚数を取得する
	GetTalonCount() int
	// GetAwards このディールで区画が動いた記録を取得する
	GetAwards() []*domain.NainJauneAward
	// GetPlayedPile 場に出た札を取得する
	GetPlayedPile() []*domain.Card
	// GetRunRank 今の並びの最高ランクを取得する (0: 好きな札で始められる)
	GetRunRank() int
	// GetDealNumber 完了したディール数を取得する
	GetDealNumber() int
	// GetDealWinner 直近ディールで出し切った席を取得する (-1: なし)
	GetDealWinner() int
	// GetWinnerIdx 勝者の添字を取得する (-1: 未決着)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.NainJaunePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.NainJaunePlayer
}
