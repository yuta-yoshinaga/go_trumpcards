//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PrimeroGame はプリメロ (Primero) のゲームインタフェース。
type PrimeroGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 次のラウンドを配る
	NextRound()
	// PlayerCall 人間 (現在の手番) がコールする
	PlayerCall() error
	// PlayerRaise 人間 (現在の手番) がレイズ (ヴィ) する
	PlayerRaise() error
	// PlayerFold 人間 (現在の手番) が降りる
	PlayerFold() error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PrimeroConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PrimeroConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PrimeroPhase
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetDealerIdx ディーラーの座席番号を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 現在の手番プレイヤーの座席番号を取得する
	GetCurrentPlayerIdx() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetCurrentBet 現在の必要総拠出額を取得する
	GetCurrentBet() int
	// GetRaiseCount このラウンドのレイズ回数を取得する
	GetRaiseCount() int
	// GetMaxRaises 1 ラウンドあたりの最大レイズ回数を取得する
	GetMaxRaises() int
	// GetAnte アンティ額を取得する
	GetAnte() int
	// GetWinnerIdx 直近ラウンドの勝者を取得する (-1 = なし)
	GetWinnerIdx() int
	// GetMatchWinnerIdx ゲーム全体の勝者を取得する (-1 = 未確定)
	GetMatchWinnerIdx() int
	// GetResult 人間から見たラウンド結果を取得する
	GetResult() domain.PrimeroResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PrimeroPlayer
	// GetChips 人間 (seat 0) の保有チップを取得する
	GetChips() int
	// IsHumanTurn 現在の手番が人間かどうかを返す
	IsHumanTurn() bool
	// CanRaise 現在の手番プレイヤーがレイズ可能かを返す
	CanRaise() bool
	// GetHint ヒントを取得する
	GetHint() *domain.PrimeroHint
}
