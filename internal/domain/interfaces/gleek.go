//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GleekGame グリーク (Gleek) のゲームインタフェース
type GleekGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間が競る (0=降りる)
	PlayerBid(bid int) error
	// CpuBid CPUプレイヤーが1回競る
	CpuBid()
	// NextBidAmount 次に置ける額を取得する (0=これ以上競り上げられない)
	NextBidAmount() int
	// HighestBid 現在の最高額を取得する
	HighestBid() int
	// PlayerDiscard 落札者が捨てる札をまとめて指定する
	PlayerDiscard(indices []int) error
	// CpuDiscard CPU の落札者に捨て札を選ばせる
	CpuDiscard()
	// GetDiscardHint 捨てるべき札の索引を取得する
	GetDiscardHint() []int
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GleekConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GleekConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GleekPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の競り手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanDiscardTurn 人間の捨て札待ちかを返す
	IsHumanDiscardTurn() bool
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
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetElderIdx エルダーインデックスを取得する
	GetElderIdx() int
	// GetTrumpSuit 切り札スートを取得する (1..4)
	GetTrumpSuit() int
	// GetTurnUp 表向きになった切り札の札を取得する
	GetTurnUp() *domain.Card
	// GetCurrentBidderIdx 現在の競り手番インデックスを取得する
	GetCurrentBidderIdx() int
	// GetBids 各席の競り額を取得する
	GetBids() [domain.GleekPlayerCnt]int
	// GetPassed 各席が降りたかを取得する
	GetPassed() [domain.GleekPlayerCnt]bool
	// GetBuyerIdx ストックを買った席を取得する (-1=未確定)
	GetBuyerIdx() int
	// GetWinningBid 落札額を取得する
	GetWinningBid() int
	// GetRuffs 各席のラフを取得する
	GetRuffs() []*domain.GleekRuff
	// GetRuffWinnerIdx ラフを取った席を取得する (-1=未確定)
	GetRuffWinnerIdx() int
	// GetMelds 申告されたグリーク / マーニヴァルを取得する
	GetMelds() []*domain.GleekMeld
	// GetTrickPoints 各席のトリック点を取得する
	GetTrickPoints() [domain.GleekPlayerCnt]int
	// GetLastTrickWinner 直前のトリックを取った席を取得する (-1=まだ無い)
	GetLastTrickWinner() int
	// DealPoints このディールで配られた点の合計を取得する
	DealPoints() int
	// Par 精算の基準点を取得する
	Par() int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.GleekPlayerCnt]int
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.GleekResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GleekPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.GleekHint
}
