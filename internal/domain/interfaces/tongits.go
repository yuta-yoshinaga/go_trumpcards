//go:build !js || !wasm || extra5

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TongitsGame Tongitsゲームインタフェース
type TongitsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札からカードを引く
	PlayerDrawFromDiscard() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerMeld 手札のカードでメルドを作り場に公開する
	PlayerMeld(indices []int) error
	// PlayerSapaw 公開済みメルド (他家のものを含む) に手札を1枚付け足す
	PlayerSapaw(targetPlayerIdx, meldIdx, cardIndex int) error
	// PlayerChallenge ドロー (challenge) を宣言する。他家全員が応じたら残り点で決着する
	PlayerChallenge(agreed []bool) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TongitsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TongitsConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TongitsPhase
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
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TongitsPlayer
	// GetIsTongits 配牌Tongitsかを返す
	GetIsTongits() bool
}
