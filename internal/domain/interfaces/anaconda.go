//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AnacondaGame はアナコンダ (Anaconda / Pass the Trash) のゲームインタフェース。
type AnacondaGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 次のラウンドを配る
	NextRound()
	// Pass 人間 (seat 0) が選んだ札を左隣へ渡す
	Pass(indices []int) error
	// Keep 人間 (seat 0) が残す 5 枚を選ぶ
	Keep(indices []int) error
	// PlayerCall 人間がコール (チェック含む) する
	PlayerCall() error
	// PlayerRaise 人間がレイズする
	PlayerRaise() error
	// PlayerFold 人間がフォールドする
	PlayerFold() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.AnacondaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.AnacondaConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AnacondaPhase
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetDealerIdx ディーラーの座席番号を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 現在の手番プレイヤーの座席番号を取得する
	GetCurrentPlayerIdx() int
	// GetPassCount 現在のパスサブラウンドで渡す枚数を取得する
	GetPassCount() int
	// GetRollIndex ロールフェーズで公開済みの枚数を取得する
	GetRollIndex() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetCurrentBet 現在のストリートで必要な拠出額を取得する
	GetCurrentBet() int
	// GetRaiseCount このストリートのレイズ回数を取得する
	GetRaiseCount() int
	// GetMaxRaises 1 ストリートあたりの最大レイズ回数を取得する
	GetMaxRaises() int
	// GetAnte アンティ額を取得する
	GetAnte() int
	// GetWinnerIdx 直近ラウンドの勝者を取得する (-1 = なし)
	GetWinnerIdx() int
	// GetMatchWinnerIdx ゲーム全体の勝者を取得する (-1 = 未確定)
	GetMatchWinnerIdx() int
	// GetResult 人間から見たラウンド結果を取得する
	GetResult() domain.AnacondaResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.AnacondaPlayer
	// GetChips 人間 (seat 0) の保有チップを取得する
	GetChips() int
	// GetRevealedCards 指定プレイヤーの公開カードを取得する
	GetRevealedCards(idx int) []*domain.Card
	// IsHandFullyRevealed 指定プレイヤーの手札が完全公開されているかを返す
	IsHandFullyRevealed(idx int) bool
	// IsHumanTurn 現在がロールフェーズの人間手番かを返す
	IsHumanTurn() bool
	// CanRaise 現在の手番プレイヤーがレイズ可能かを返す
	CanRaise() bool
	// GetHint ヒントを取得する
	GetHint() *domain.AnacondaHint
}
