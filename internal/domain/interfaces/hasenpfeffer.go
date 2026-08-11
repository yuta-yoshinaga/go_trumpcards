//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HasenpfefferGame ハーゼンプフェファー ゲームインタフェース
type HasenpfefferGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerBid 人間が宣言する (0: 降りる)
	PlayerBid(bid int) error
	// CpuBid CPUが1人分宣言する
	CpuBid()
	// PlayerDiscard 人間が切り札を宣言して1枚捨てる
	PlayerDiscard(cardIndex, suit int) error
	// CpuDiscard CPUの落札者が切り札を宣言して1枚捨てる
	CpuDiscard()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextHand 次のハンドを開始する
	NextHand()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HasenpfefferConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HasenpfefferConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HasenpfefferPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 人間が宣言する番かを返す
	IsHumanBidTurn() bool
	// IsHumanDiscardTurn 人間が捨て札をする番かを返す
	IsHumanDiscardTurn() bool
	// MustBid 指定席が降りられない (義務競り) かを返す
	MustBid(playerIdx int) bool
	// NextBid 次に出せる最小の宣言額を返す (0: もう宣言できない)
	NextBid() int
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (0: 未宣言)
	GetTrumpSuit() int
	// GetDeclarerIdx 落札者を取得する (-1: 競り中)
	GetDeclarerIdx() int
	// GetContract 落札額を取得する
	GetContract() int
	// GetBlindSize 伏せ札の枚数を取得する
	GetBlindSize() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetScore チームの得点を取得する
	GetScore(team int) int
	// TeamTricks チームの獲得トリック数を取得する
	TeamTricks(team int) int
	// GetLastHandEuchred 直前のハンドで落札側が失敗したかを取得する
	GetLastHandEuchred() bool
	// GetLastHandTricks 直前のハンドで落札側が取ったトリック数を取得する
	GetLastHandTricks() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HasenpfefferPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.HasenpfefferHint
}
