//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BuraGame ブラゲームインタフェース
type BuraGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCards プレイヤーが手札の indices を出す
	PlayCards(idx int, indices []int) error
	// Claim 31点到達を宣言する (足りなければ相手の勝ちになる)
	Claim(idx int) error
	// DeclareCombination 手札の即勝ち役を宣言する
	DeclareCombination(idx int) error
	// BuraCpuDecide CPU が取る手を決める
	BuraCpuDecide(idx int) domain.BuraCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BuraConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BuraConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BuraPhase
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetCurrentLead リードされているカードを取得する (解決済みなら空)
	GetCurrentLead() []*domain.Card
	// GetTrickNumber 完了したトリック数を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切札スートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 表向きの切札指示カードを取得する (引かれた後は nil)
	GetTrumpCard() *domain.Card
	// GetStock 山札を取得する
	GetStock() []*domain.Card
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.BuraPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BuraPlayer
	// GetPlayerPoints プレイヤーの累積得点を取得する
	GetPlayerPoints(i int) int
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定または引き分け)
	GetWinnerIdx() int
	// IsDraw 宣言のないまま流局したかを取得する
	IsDraw() bool
}
