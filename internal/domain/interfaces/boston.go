//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BostonGame ボストン (Boston) ゲームインタフェース
type BostonGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// Bid 宣言する
	Bid(player int, level domain.BostonBidLevel, suit int) error
	// PassBid 宣言を見送る
	PassBid(player int) error
	// CallPartner パートナーを指名する (-1 なら単独)
	CallPartner(player, partner int) error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// BostonValidPlays 出せる手札インデックスを返す
	BostonValidPlays(player int) []int
	// BostonIsDeclarerSide 席が落札側かを返す
	BostonIsDeclarerSide(seat int) bool
	// BostonDeclarerTricks 落札側が取ったトリック数を返す
	BostonDeclarerTricks() int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BostonConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BostonConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BostonPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx 宣言中の手番を取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーを取得する
	GetDealerIdx() int
	// GetBids この局の宣言履歴を取得する
	GetBids() []*domain.BostonBidRecord
	// GetHighBid 現在の最高宣言を取得する
	GetHighBid() *domain.BostonBidRecord
	// GetDeclarerIdx 落札者を取得する
	GetDeclarerIdx() int
	// GetPartnerIdx 指名されたパートナーを取得する
	GetPartnerIdx() int
	// GetTrumpSuit 切札を取得する (0 なら切札なし)
	GetTrumpSuit() int
	// IsExposed 落札者の手札を公開しているかを取得する
	IsExposed() bool
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetTricksWon 席が取ったトリック数を取得する
	GetTricksWon(idx int) int
	// IsBidMade 直前の局で落札側が達成したかを取得する
	IsBidMade() bool
	// GetChips 席の通算を取得する
	GetChips(idx int) int
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetTargetHands 規定局数を取得する
	GetTargetHands() int
	// GetWinnerIdx 勝者を取得する
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.BostonPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.BostonPlayer
}
