//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ShelemGame シェレム (Shelem) ゲームインタフェース
type ShelemGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerBid 点数で入札する
	PlayerBid(bid int) error
	// PlayerBidShelem Shelem（全トリック独占）を宣言する
	PlayerBidShelem() error
	// PlayerPass 競りを降りる
	PlayerPass() error
	// CpuBid 手番の CPU が競りの判断をする
	CpuBid()
	// PlayerDiscard 落札者が 4 枚捨てて切り札を決める
	PlayerDiscard(indices []int, suit int) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ShelemConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ShelemConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ShelemPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 人間が入札する番かを返す
	IsHumanBidTurn() bool
	// IsHumanDiscardTurn 人間が捨て札と切り札を決める番かを返す
	IsHumanDiscardTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (未決定は 0)
	GetTrumpSuit() int
	// GetDeclarerIdx 落札者を取得する (-1: 未決定)
	GetDeclarerIdx() int
	// GetContract 落札した点数を取得する
	GetContract() int
	// GetShelemBid Shelem 宣言で落札したかを取得する
	GetShelemBid() bool
	// GetWidowSize 伏せられているウィドウの枚数を取得する
	GetWidowSize() int
	// GetScore チームの累計得点を取得する
	GetScore(team int) int
	// GetRoundPoints チームの現ラウンドのカード点を取得する
	GetRoundPoints(team int) int
	// TeamTricks チームの獲得トリック数を取得する
	TeamTricks(team int) int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx 競りの手番を取得する
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
	GetPlayer(i int) *domain.ShelemPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.ShelemHint
}
