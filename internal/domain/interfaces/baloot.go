//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BalootGame バルート (Baloot) ゲームインタフェース
type BalootGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// DeclareSun 切り札なし (Sun) を宣言する
	DeclareSun() error
	// DeclareHokom 指定スートを切り札として Hokom を宣言する
	DeclareHokom(suit int) error
	// PassDeclaration 宣言を見送る（親は見送れない）
	PassDeclaration() error
	// CpuDeclare 手番の CPU がモードを宣言する
	CpuDeclare()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BalootConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BalootConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BalootPhase
	// GetMode 現在のモード (Sun / Hokom) を取得する
	GetMode() domain.BalootMode
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanDeclareTurn 人間がモードを宣言する番かを返す
	IsHumanDeclareTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する (Sun では 0)
	GetTrumpSuit() int
	// GetDeclarerIdx モードを宣言したプレイヤーを取得する (-1: 未決定)
	GetDeclarerIdx() int
	// GetScore チームの累計得点を取得する
	GetScore(team int) int
	// GetRoundPoints チームの現ラウンド点を取得する
	GetRoundPoints(team int) int
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
	GetPlayer(i int) *domain.BalootPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.BalootHint
}
