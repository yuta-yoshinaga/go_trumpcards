//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PopeJoanGame ポープ・ジョーンゲームインタフェース
type PopeJoanGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Play 手札1枚を出す
	Play(player, handIdx int) error
	// NextDeal 次のディールを配る
	NextDeal() error
	// PopeJoanCpuDecide CPU が出す手札の添字を返す (-1: 出せない)
	PopeJoanCpuDecide(idx int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PopeJoanConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PopeJoanConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PopeJoanPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetBoard 8区画の残高を取得する
	GetBoard() domain.PopeJoanBoard
	// GetTrumpSuit トランプを取得する
	GetTrumpSuit() int
	// GetTurnUp dead hand の最後の1枚を取得する
	GetTurnUp() *domain.Card
	// GetAwards このディールで区画が動いた記録を取得する
	GetAwards() []*domain.PopeJoanAward
	// GetPlayedPile 場に出た札を取得する
	GetPlayedPile() []*domain.Card
	// GetRunSuit 今の並びのスートを取得する (-1: 好きな札で開始できる)
	GetRunSuit() int
	// GetRunRank 今の並びの最高ランクを取得する
	GetRunRank() int
	// GetDealNumber 完了したディール数を取得する
	GetDealNumber() int
	// GetDealWinner 直近ディールで出し切った席を取得する (-1: なし)
	GetDealWinner() int
	// GetWinnerIdx 勝者の添字を取得する (-1: 未決着)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.PopeJoanPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.PopeJoanPlayer
}
