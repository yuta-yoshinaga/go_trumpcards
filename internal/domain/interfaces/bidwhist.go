//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BidWhistGame Bid Whist ゲームインタフェース
type BidWhistGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーがビッドする (tricks=ブックを超える目標数, direction=Uptown/Downtown/NoTrump)
	PlayerBid(tricks, direction int) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPUプレイヤーが1ビッド実行する
	CpuBid()
	// PlayerDeclareTrump 人間(落札者)が切り札スートを宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPU(落札者)が切り札スートを宣言する
	CpuDeclareTrump()
	// PlayerExchangeKitty 人間(落札者)が6枚捨てる
	PlayerExchangeKitty(discardIndices []int) error
	// CpuExchange CPU(落札者)が6枚捨てる
	CpuExchange()
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
	GetConfig() domain.BidWhistConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BidWhistConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BidWhistPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanDeclarerTurn 現在の落札者手番 (切り札宣言/キティ交換) が人間かを返す
	IsHumanDeclarerTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッド手番インデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する (-1 = NT/未宣言)
	GetTrumpSuit() int
	// GetContractTricks 契約トリック数を取得する (ブックを超える目標数)
	GetContractTricks() int
	// GetContractDirection 契約方向を取得する (Uptown/Downtown/NoTrump)
	GetContractDirection() int
	// GetDeclarerIdx 落札者インデックスを取得する (-1 = 未確定)
	GetDeclarerIdx() int
	// GetHighestBid 現在の最高ビッドを取得する (nil = なし)
	GetHighestBid() *domain.BidWhistBid
	// GetHighestBidder 最高ビッダーのインデックスを取得する (-1 = なし)
	GetHighestBidder() int
	// GetKitty キティを取得する
	GetKitty() []*domain.Card
	// GetKittyIndices キティ交換フェーズ中、落札者の手札のうちキティ由来カードのインデックスを取得する
	GetKittyIndices() []int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetWinnerTeam 勝利チームを取得する (-1 = 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BidWhistPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.BidWhistHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
