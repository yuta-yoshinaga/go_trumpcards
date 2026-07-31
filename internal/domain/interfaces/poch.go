//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PochGame ポッホゲームインタフェース
type PochGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet pochen で1単位賭ける
	Bet(player int) error
	// Fold pochen で降りる
	Fold(player int) error
	// Play ストップスで手札1枚を出す
	Play(player, handIdx int) error
	// NextDeal 次のディールを配る
	NextDeal() error
	// PochCpuDecide CPU が取る手を決める
	PochCpuDecide(idx int) domain.PochCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PochConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PochConfig)

	// GetGameEndFlag 決着しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PochPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetBoard 9プールの残高を取得する
	GetBoard() domain.PochBoard
	// GetPaySuit pay suit を取得する
	GetPaySuit() int
	// GetTurnUp 表向きにした余り札を取得する
	GetTurnUp() *domain.Card
	// GetStakingAwards 直近ディールの第1段階の結果を取得する
	GetStakingAwards() []*domain.PochStakingAward
	// GetBetTarget 現在の賭け額を取得する
	GetBetTarget() int
	// GetPochenWinner pochen を取った席を取得する (-1: 未決着)
	GetPochenWinner() int
	// GetPochenPot pochen で動いたチップを取得する
	GetPochenPot() int
	// GetPlayedPile ストップスで出た札を取得する
	GetPlayedPile() []*domain.Card
	// GetStopsSuit 今の並びのスートを取得する (-1: 好きな札で開始できる)
	GetStopsSuit() int
	// GetStopsRank 今の並びの最高ランクを取得する
	GetStopsRank() int
	// GetDealNumber 完了したディール数を取得する
	GetDealNumber() int
	// GetDealWinner 直近ディールで出し切った席を取得する (-1: なし)
	GetDealWinner() int
	// GetWinnerIdx 勝者の添字を取得する (-1: 未決着)
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.PochPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.PochPlayer
}
