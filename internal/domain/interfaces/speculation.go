//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpeculationGame はスペキュレーションのドメインが満たす操作。
type SpeculationGame interface {
	// Reset は次のラウンドを始める。
	Reset()
	// Flip は手番の席の伏せ札を 1 枚めくる。
	Flip() error
	// Accept は競りの申し出を受ける。
	Accept() error
	// Decline は競りの申し出を断る。
	Decline() error
	// Bid は提示額に上乗せして買う。
	Bid(amount int) error
	// NextRound は決着後に次のラウンドへ進む。
	NextRound() error

	// GetPhase は現在のフェーズを返す。
	GetPhase() domain.SpeculationPhase
	// GetPlayers は全プレイヤーを返す。
	GetPlayers() []*domain.SpeculationPlayer
	// GetConfig は卓設定を返す。
	GetConfig() domain.SpeculationConfig
	// GetTrumpSuit はこのラウンドの切り札スートを返す。
	GetTrumpSuit() int
	// GetTrumpCard は切り札を決めた札を返す。
	GetTrumpCard() *domain.Card
	// GetPot は現在のポットを返す。
	GetPot() int
	// GetTurnSeat は次にめくる席を返す。
	GetTurnSeat() int
	// GetBestSeat は最高切り札を持つ席を返す。無ければ -1。
	GetBestSeat() int
	// GetOfferFrom は買い取りを申し出ている席を返す。無ければ -1。
	GetOfferFrom() int
	// GetOfferTo は申し出を受けている席を返す。無ければ -1。
	GetOfferTo() int
	// GetOfferAmount は提示額を返す。
	GetOfferAmount() int
	// GetRoundNo は消化済みのラウンド数を返す。
	GetRoundNo() int
	// GetWinnerSeat は直前のラウンドの勝者席を返す。決着前・流局なら -1。
	GetWinnerSeat() int
	// GetGameEndFlag はゲームが終わったかを返す。
	GetGameEndFlag() bool
	// GetActionLog は棋譜を返す。
	GetActionLog() []*domain.ActionLogEntry
}
