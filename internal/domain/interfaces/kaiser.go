//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KaiserGame カイザー (Kaiser) ゲームインタフェース
type KaiserGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// Bid 点数を宣言する
	Bid(player, value int, contract domain.KaiserContract) error
	// PassBid ビッドを見送る
	PassBid(player int) error
	// SetTrump 落札者が切札を指定する
	SetTrump(player, suit int) error
	// Discard 落札者がキティを取り込んだあと2枚捨てる
	Discard(player int, idxs []int) error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// KaiserValidPlays 出せる手札インデックスを返す
	KaiserValidPlays(player int) []int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KaiserConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KaiserConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KaiserPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx ビッド中の手番を取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーを取得する
	GetDealerIdx() int
	// GetBids この局のビッド履歴を取得する
	GetBids() []*domain.KaiserBid
	// GetHighBid 現在の最高ビッドを取得する
	GetHighBid() *domain.KaiserBid
	// GetDeclarerIdx 落札者を取得する
	GetDeclarerIdx() int
	// GetTrumpSuit 切札を取得する (0 ならノートランプ系)
	GetTrumpSuit() int
	// GetContract 契約種別を取得する
	GetContract() domain.KaiserContract
	// GetKittySize キティの残り枚数を取得する
	GetKittySize() int
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetHandPoints チームがこの局で取った点を取得する
	GetHandPoints(team int) int
	// GetHeartFiveBy ♥5 を取った席を取得する
	GetHeartFiveBy() int
	// GetSpadeThreeBy ♠3 を取った席を取得する
	GetSpadeThreeBy() int
	// IsBidMade 直前の局で落札側が達成したかを取得する
	IsBidMade() bool
	// GetScore チームの通算点を取得する
	GetScore(team int) int
	// GetTargetScore 現在の目標点を取得する
	GetTargetScore() int
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.KaiserPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.KaiserPlayer
}
