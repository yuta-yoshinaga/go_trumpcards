//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MendikotGame メンディコット (Mendikot) ゲームインタフェース
type MendikotGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextHand 次のハンドを開始する
	NextHand()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MendikotConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MendikotConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MendikotPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (0: まだ決まっていない)
	GetTrumpSuit() int
	// GetTrumpChooserIdx 切り札を決めたプレイヤーを取得する (-1: 未決定)
	GetTrumpChooserIdx() int
	// GetScore チームのハンド勝ち点を取得する
	GetScore(team int) int
	// TeamTens チームが獲得した 10 の枚数を取得する
	TeamTens(team int) int
	// TeamTricks チームの獲得トリック数を取得する
	TeamTricks(team int) int
	// GetLastHandWinner 直前のハンドを制したチームを取得する (-1: まだ無い)
	GetLastHandWinner() int
	// GetLastHandKind 直前のハンドの結末の種類を取得する
	GetLastHandKind() string
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
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MendikotPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.MendikotHint
}
