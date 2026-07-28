//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MightyGame マイティゲームインタフェース
type MightyGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする (0 = パス、noTrump = ノートランプ宣言)
	PlayerBid(bid int, noTrump bool) error
	// CpuBid CPUプレイヤーがビッドする
	CpuBid()
	// PlayerDeclareTrumpAndFriend 切り札とパートナーを宣言する (suit は MightyTrumpNone でノートランプ)
	PlayerDeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) error
	// CpuDeclareTrumpAndFriend CPU宣言者が切り札とパートナーを宣言する
	CpuDeclareTrumpAndFriend()
	// PlayerExchangeKitty 場札を交換する (捨て札を3枚指定)
	PlayerExchangeKitty(discardIndices []int) error
	// CpuExchangeKitty CPU宣言者が場札を交換する
	CpuExchangeKitty()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerPlayJokerLead プレイヤーがジョーカーをリードする (要求スート指定)
	PlayerPlayJokerLead(cardIndex int, demandSuit int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MightyConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MightyConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MightyPhase
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
	GetCurrentTrick() []*domain.MightyTrickCard
	// GetTrumpSuit 切り札スートを取得する (MightyTrumpNone = ノートランプ)
	GetTrumpSuit() int
	// GetPartnerCard パートナー指名カードを取得する
	GetPartnerCard() *domain.Card
	// GetDeclarerIdx 宣言者のプレイヤーインデックスを取得する
	GetDeclarerIdx() int
	// GetPartnerIdx パートナーのプレイヤーインデックスを取得する (宣言者と同じなら単独宣言)
	GetPartnerIdx() int
	// GetPartnerRevealed パートナー公開状態を取得する
	GetPartnerRevealed() bool
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
	// GetWinningBidNoTrump 落札ビッドがノートランプかどうか
	GetWinningBidNoTrump() bool
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MightyPlayer
	// GetValidPlayIndices プレイ可能なカードインデックスを取得する
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.MightyHint
}
