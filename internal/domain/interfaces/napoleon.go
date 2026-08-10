//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NapoleonGame ナポレオンゲームインタフェース
type NapoleonGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする (0 = パス)
	PlayerBid(bid int) error
	// CpuBid CPUプレイヤーがビッドする
	CpuBid()
	// PlayerDeclareTrump 切り札と副官を宣言する
	PlayerDeclareTrump(suit int, adjSuit int, adjVal int) error
	// CpuDeclareTrump CPUナポレオンが切り札と副官を宣言する
	CpuDeclareTrump()
	// PlayerExchangeKitty 場札を交換する
	PlayerExchangeKitty(discardIndex int) error
	// CpuExchangeKitty CPUナポレオンが場札を交換する
	CpuExchangeKitty()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.NapoleonConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.NapoleonConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.NapoleonPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanDeclareTurn 切り札宣言が人間の番かを返す
	IsHumanDeclareTurn() bool
	// IsHumanExchangeTurn 場札交換が人間の番かを返す
	IsHumanExchangeTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.NapoleonTrickCard
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetAdjutantCard 副官カードを取得する
	GetAdjutantCard() *domain.Card
	// GetNapoleonIdx ナポレオンインデックスを取得する
	GetNapoleonIdx() int
	// GetAdjutantIdx 副官インデックスを取得する
	GetAdjutantIdx() int
	// GetAdjutantRevealed 副官公開状態を取得する
	GetAdjutantRevealed() bool
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetKitty 場札を取得する
	GetKitty() []*domain.Card
	// GetHighestBid 最高ビッドを取得する
	GetHighestBid() int
	// GetHighestBidder 最高ビッドプレイヤーを取得する
	GetHighestBidder() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.NapoleonPlayer
	// GetValidPlayIndices プレイ可能なカードインデックスを取得する
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.NapoleonHint
	// GetHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
	GetHumanIdx() int
}
