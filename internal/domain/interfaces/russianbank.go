//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RussianBankGame ロシアンバンク (クラペット) のゲームインタフェース。
type RussianBankGame interface {
	BaseGame
	// Reset ゲームを初期化する。
	Reset()
	// SetConfig ゲーム設定をセットする。
	SetConfig(cfg domain.RussianBankConfig)
	// GetConfig ゲーム設定を取得する。
	GetConfig() domain.RussianBankConfig

	// MoveToFoundation 移動元トップを任意のファウンデーションへ移す。
	MoveToFoundation(src domain.RussianBankSource) error
	// MoveToTableau 移動元トップを共有タブロー col へ移す。
	MoveToTableau(src domain.RussianBankSource, col int) error
	// Discard 手札を 1 枚廃札に送り手番を終える。
	Discard() error
	// CallStop CPU の取りこぼしを咎める。
	CallStop() error
	// Undo 直近の人間の 1 手を取り消す。
	Undo() error
	// CanUndo Undo 可能か。
	CanUndo() bool
	// RunCpuTurn CPU の手番を自動進行する。
	RunCpuTurn()

	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.RussianBankPhase
	// GetGameEndFlag 決着済みかを返す。
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す。
	IsHumanTurn() bool
	// CanCallStop 今 stop を宣言できるかを返す。
	CanCallStop() bool
	// GetCurrentPlayer 手番プレイヤー seat を取得する。
	GetCurrentPlayer() int
	// GetWinner 勝者 seat を取得する (-1=未確定/引き分け)。
	GetWinner() int
	// GetMoveCount 累計手数を取得する。
	GetMoveCount() int
	// GetPlayers プレイヤー列を取得する。
	GetPlayers() []*domain.RussianBankPlayer
	// GetPlayer seat のプレイヤーを取得する。
	GetPlayer(seat int) *domain.RussianBankPlayer
	// GetTableau 共有タブローを取得する。
	GetTableau() [domain.RussianBankTableauCnt][]*domain.Card
	// GetFoundations 共有ファウンデーションを取得する。
	GetFoundations() [domain.RussianBankFoundationCnt][]*domain.Card
	// GetStopPoints seat の stop 得点を取得する。
	GetStopPoints(seat int) int
	// GetHint ヒントを取得する。
	GetHint() *domain.RussianBankHint
}
