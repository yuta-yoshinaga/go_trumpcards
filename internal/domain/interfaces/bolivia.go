//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BoliviaGame ボリビアゲームインタフェース
type BoliviaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札の山を取る
	PlayerDrawFromDiscard(naturalPairIndices []int) error
	// PlayerMeld プレイヤーがメルドを出す
	PlayerMeld(meldGroups [][]int) error
	// PlayerSkipMeld メルドフェーズをスキップする
	PlayerSkipMeld() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerGoOut プレイヤーが上がる
	PlayerGoOut() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BoliviaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BoliviaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BoliviaPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetDiscardPileCount 捨て札の枚数を取得する
	GetDiscardPileCount() int
	// GetIsFrozen 捨て札の山がフリーズ状態かを返す
	GetIsFrozen() bool
	// GetWinnerIdx 勝利チームインデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BoliviaPlayer
	// GetTeamCount チーム数を取得する
	GetTeamCount() int
	// GetTeamScore チームの累積スコアを取得する
	GetTeamScore(team int) int
	// GetMinimumMeldValue 初回メルドに要する最低点を取得する
	GetMinimumMeldValue(playerIdx int) int
	// GetDrewFromDiscard 捨て札から引いたかを返す
	GetDrewFromDiscard() bool
}
