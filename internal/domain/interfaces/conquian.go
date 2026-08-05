//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ConquianGame コンキャンゲームインタフェース
type ConquianGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札からカードを引く (強制使用)
	PlayerDrawFromDiscard() error
	// PlayerMeld プレイヤーが手札からメルドを並べる/付ける
	PlayerMeld(meldGroups [][]int) error
	// PlayerMeldWithTargets 延長先メルドの指定つきでメルドする
	PlayerMeldWithTargets(meldGroups [][]int, extendTargets []int) error
	// GetExtendableMeldIndices その札を足せるメルド番号を返す
	GetExtendableMeldIndices(playerIdx int, card *domain.Card) []int
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ConquianConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ConquianConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ConquianPhase
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
	// GetWinnerIdx マッチ勝者インデックスを取得する
	GetWinnerIdx() int
	// GetRoundWinnerIdx 直近ラウンド勝者インデックスを取得する
	GetRoundWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ConquianPlayer
	// GetTookDiscard 今ターンに捨て札を取ったか (強制使用待ち) を返す
	GetTookDiscard() bool
}
