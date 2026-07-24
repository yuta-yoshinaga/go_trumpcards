//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HandAndFootGame ハンドアンドフットゲームインタフェース
type HandAndFootGame interface {
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
	// SuggestMelds playerIdx が作れるメルド候補 (カード群) を返す。無ければ nil
	SuggestMelds(playerIdx int) [][]*domain.Card
	// PlayerSkipMeld メルドフェーズをスキップする
	PlayerSkipMeld() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerGoOut プレイヤーが上がる
	PlayerGoOut() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HandAndFootConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HandAndFootConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HandAndFootPhase
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
	// GetWinnerTeam 勝利チームインデックスを取得する
	GetWinnerTeam() int
	// GetWinnerIdx 勝利チームの代表プレイヤーインデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HandAndFootPlayer
	// GetTeamMelds 指定チームのメルドを取得する
	GetTeamMelds(team int) []*domain.CanastaMeld
	// GetTeamRed3s 指定チームの赤3を取得する
	GetTeamRed3s(team int) []*domain.Card
	// GetDrewFromDiscard 捨て札から引いたかを返す
	GetDrewFromDiscard() bool
}
