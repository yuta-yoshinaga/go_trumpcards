//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BlackJackSwitchGame ブラックジャック・スイッチゲームインタフェース。
//
// 標準BJの全機能（スプリット/サレンダー/インシュランス/サイドベット/CPU）
// は本バリアントには無く、最小限のヒット/スタンド/ダブルダウンに加えて
// 「Switch（2枚目交換）」アクションを公開する。
type BlackJackSwitchGame interface {
	// SwitchPreviewScores 2枚目を入れ替えた場合の両ハンドの得点 (ok=false なら不可)
	SwitchPreviewScores() (int, int, bool)
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerBet 1ハンドあたりのベットを置きカードを配る
	PlayerBet(amount int) error
	// PlayerSwitch 2ハンドの2枚目を交換する
	PlayerSwitch() error
	// PlayerKeep スイッチを行わずアクションフェーズへ進む
	PlayerKeep() error
	// PlayerHit 現在のハンドをヒット
	PlayerHit() error
	// PlayerStand 現在のハンドをスタンド
	PlayerStand() error
	// PlayerDoubleDown 現在のハンドをダブルダウン
	PlayerDoubleDown() error

	// GetPlayer プレイヤー
	GetPlayer() *domain.BlackJackPlayer
	// GetDealer ディーラー
	GetDealer() *domain.BlackJackPlayer
	// GetHands ハンド一覧（常に2件）
	GetHands() []*domain.BlackJackHand
	// GetCurrentHandIdx 現在操作中のハンドインデックス
	GetCurrentHandIdx() int
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// IsSwitched 直近のラウンドでスイッチを実行したか
	IsSwitched() bool
	// IsDealerPushed22 ディーラー22プッシュが発生したか
	IsDealerPushed22() bool
	// GetHandResults 各ハンドの勝敗結果
	GetHandResults() []domain.GameResult
	// GetHandPayouts 各ハンドの配当（ベット返却込み）
	GetHandPayouts() []int
	// GetTotalPayout 全ハンドの合計配当
	GetTotalPayout() int
	// GetOverallResult 総合勝敗
	GetOverallResult() domain.GameResult
}
