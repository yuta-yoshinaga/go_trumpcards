//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FiveHundredGame 500 (Five Hundred) ゲームインタフェース
type FiveHundredGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーがビッドする
	PlayerBid(kind domain.FiveHundredContractKind, tricks, suit int) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPUプレイヤーが1ビッド実行する
	CpuBid()
	// PlayerExchangeKitty 人間(落札者)が3枚捨てる
	PlayerExchangeKitty(discardIndices []int) error
	// CpuExchange CPU(落札者)が3枚捨てる
	CpuExchange()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex, jokerSuit int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.FiveHundredConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.FiveHundredConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FiveHundredPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
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
	// GetTrumpSuit 切り札スートを取得する (-1 = なし)
	GetTrumpSuit() int
	// GetContractKind 契約種別を取得する
	GetContractKind() int
	// GetContractTricks 契約トリック数を取得する
	GetContractTricks() int
	// GetContractValue 契約の得点を取得する
	GetContractValue() int
	// GetDeclarerIdx 落札者インデックスを取得する (-1 = 未確定)
	GetDeclarerIdx() int
	// GetRoundResult ラウンド終了時の得点内訳 (それ以外は nil)
	GetRoundResult() *domain.FiveHundredRoundResult
	// GetHighestBid 現在の最高ビッドを取得する (nil = なし)
	GetHighestBid() *domain.FiveHundredBid
	// GetHighestBidder 最高ビッダーのインデックスを取得する (-1 = なし)
	GetHighestBidder() int
	// GetJokerLeadSuit ジョーカーリードの指名スートを取得する (-1 = なし)
	GetJokerLeadSuit() int
	// GetKitty キティを取得する
	GetKitty() []*domain.Card
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetWinnerTeam 勝利チームを取得する (-1 = 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FiveHundredPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.FiveHundredHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
