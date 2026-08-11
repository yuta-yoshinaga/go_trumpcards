//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HokmGame ホクム (Hokm) ゲームインタフェース
type HokmGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerDeclareTrump 親が切り札スートを宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPU の親が切り札スートを宣言する
	CpuDeclareTrump()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextHand 次のハンドを開始する
	NextHand()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HokmConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HokmConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HokmPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanTrumpTurn 人間が切り札を宣言する番かを返す
	IsHumanTrumpTurn() bool
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (宣言前は 0)
	GetTrumpSuit() int
	// GetHakemIdx 親 (Hakem) のインデックスを取得する
	GetHakemIdx() int
	// GetScore チームのハンド勝ち点を取得する
	GetScore(team int) int
	// TeamTricks チームの獲得トリック数を取得する
	TeamTricks(team int) int
	// GetLastHandKot 直前のハンドが Kot だったかを取得する
	GetLastHandKot() bool
	// GetLastHandWinner 直前のハンドを制したチームを取得する (-1: まだ無い)
	GetLastHandWinner() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HokmPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.HokmHint
}
