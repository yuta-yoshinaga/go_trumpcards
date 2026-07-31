//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BidEuchreGame ビッド・ユーカー (Bid Euchre) ゲームインタフェース
type BidEuchreGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// Bid 宣言する
	Bid(player, value int) error
	// PassBid 宣言を見送る
	PassBid(player int) error
	// ChooseTrump 切札を宣言する
	ChooseTrump(player int, t domain.BidEuchreTrump) error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// BidEuchreValidPlays 出せる手札インデックスを返す
	BidEuchreValidPlays(player int) []int
	// BidEuchreTeamTricks チームが取ったトリック数を返す
	BidEuchreTeamTricks(team int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BidEuchreConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BidEuchreConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BidEuchrePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx 宣言中の手番を取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーを取得する
	GetDealerIdx() int
	// GetBids この局の宣言履歴を取得する
	GetBids() []*domain.BidEuchreBid
	// GetHighBid 現在の最高宣言を取得する
	GetHighBid() *domain.BidEuchreBid
	// GetDeclarerIdx 落札者を取得する
	GetDeclarerIdx() int
	// GetTrump 切札の宣言内容を取得する
	GetTrump() domain.BidEuchreTrump
	// GetTrumpSuit 切札スートを取得する (0 ならノートランプ)
	GetTrumpSuit() int
	// IsTrumpChosen 切札が宣言済みかを返す
	IsTrumpChosen() bool
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetTricksWon 席が取ったトリック数を取得する
	GetTricksWon(idx int) int
	// GetScore チームの得点を取得する
	GetScore(team int) int
	// GetLastResult 直前の局の精算を取得する
	GetLastResult() *domain.BidEuchreHandResult
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.BidEuchrePlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.BidEuchrePlayer
}
