//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SixBidSoloGame シックスビッド・ソロ (Six-Bid Solo) ゲームインタフェース
type SixBidSoloGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// Bid 宣言する
	Bid(player int, kind domain.SixBidSoloBidKind) error
	// PassBid 宣言を見送る
	PassBid(player int) error
	// Declare 切札 (とコール・ソロの指名札) を決める
	Declare(player, suit int, called *domain.Card) error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// SixBidSoloValidPlays 出せる手札インデックスを返す
	SixBidSoloValidPlays(player int) []int
	// SixBidSoloWidowPoints ウィドウのカード点を返す
	SixBidSoloWidowPoints() int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SixBidSoloConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SixBidSoloConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SixBidSoloPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx 宣言中の手番を取得する
	GetBidPlayerIdx() int
	// GetDealerIdx 親を取得する
	GetDealerIdx() int
	// GetBids この局の宣言履歴を取得する
	GetBids() []*domain.SixBidSoloBid
	// GetHighBid 現在の最高宣言を取得する
	GetHighBid() *domain.SixBidSoloBid
	// GetDeclarerIdx 落札者を取得する
	GetDeclarerIdx() int
	// GetTrumpSuit 切札を取得する (ミゼール系は 0)
	GetTrumpSuit() int
	// IsDeclared 切札が確定済みかを返す
	IsDeclared() bool
	// GetCalledCard コール・ソロで指名された札を取得する
	GetCalledCard() *domain.Card
	// IsSpreadOpen スプレッド・ミゼールで手札が公開済みかを返す
	IsSpreadOpen() bool
	// GetWidow ウィドウを取得する
	GetWidow() []*domain.Card
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetPoints 席が取ったカード点を取得する
	GetPoints(idx int) int
	// GetTricksWon 席が取ったトリック数を取得する
	GetTricksWon(idx int) int
	// GetScore 席の通算得点を取得する
	GetScore(idx int) int
	// GetLastResult 直前の局の精算を取得する
	GetLastResult() *domain.SixBidSoloHandResult
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerIdx 勝者の席を取得する
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.SixBidSoloPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.SixBidSoloPlayer
}
