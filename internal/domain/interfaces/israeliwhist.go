//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// IsraeliWhistGame イスラエリホイスト (Israeli Whist) ゲームインタフェース
type IsraeliWhistGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerAuctionBid 1 段階目のオークションで入札する
	PlayerAuctionBid(bid, suit int) error
	// PlayerAuctionPass オークションを降りる
	PlayerAuctionPass() error
	// CpuAuction 手番の CPU がオークションの判断をする
	CpuAuction()
	// PlayerBid 2 段階目で目標トリック数を宣言する
	PlayerBid(bid int) error
	// CpuBid 手番の CPU が宣言する
	CpuBid()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.IsraeliWhistConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.IsraeliWhistConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.IsraeliWhistPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanAuctionTurn 人間がオークションで判断する番かを返す
	IsHumanAuctionTurn() bool
	// IsHumanBidTurn 人間が宣言する番かを返す
	IsHumanBidTurn() bool
	// MinimumBidFor そのプレイヤーが宣言できる下限を返す（落札者のノルマ）
	MinimumBidFor(idx int) int
	// GetRestrictedBid 最後の宣言者が選べない宣言値を返す (-1: 制限なし)
	GetRestrictedBid() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetRoundDoubled 直前のラウンドで得点が 2 倍になったかを取得する
	GetRoundDoubled() bool
	// GetRoundAllExact 2 倍の理由が全員的中かを取得する (false なら全員外し)
	GetRoundAllExact() bool
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (オークション中は 0)
	GetTrumpSuit() int
	// GetDeclarerIdx オークションの落札者を取得する (-1: 未決定)
	GetDeclarerIdx() int
	// GetHighBid 現在の最高入札のトリック数を取得する
	GetHighBid() int
	// GetHighSuit 現在の最高入札のスートを取得する
	GetHighSuit() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetAuctionPlayerIdx オークションの手番を取得する
	GetAuctionPlayerIdx() int
	// GetBidPlayerIdx 宣言の手番を取得する
	GetBidPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.IsraeliWhistPlayer
	// GetWinnerIdx 勝利プレイヤーを取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.IsraeliWhistHint
}
