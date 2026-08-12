//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BotifarraGame ボティファラゲームインタフェース
type BotifarraGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Declare 切り札を宣言する (BotifarraNoTrump で切り札なし)
	Declare(suit int) error
	// Delegate 宣言を相方に委ねる
	Delegate() error
	// Double 倍付けを宣言する
	Double() error
	// PassDouble 倍付けを見送る
	PassDouble() error
	// PlayCard 札を出す
	PlayCard(cardIndex int) error
	// NextRound 次のラウンドを配る
	NextRound() error
	// GiveUp 投了する
	GiveUp()
	// CpuPlay CPUの手番を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BotifarraConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BotifarraConfig)

	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// IsHumanTurn 人間の入力を待っているか
	IsHumanTurn() bool
	// GetValidPlayIndices 出せる手札の位置
	GetValidPlayIndices(playerIdx int) []int
	// GetDealerIdx 親の席
	GetDealerIdx() int
	// GetDeclarerIdx 切り札を決めた席 (-1: 未決定)
	GetDeclarerIdx() int
	// GetTrumpSuit 切り札 (BotifarraNoTrump: 切り札なし)
	GetTrumpSuit() int
	// GetMultiplier 倍付けの倍率
	GetMultiplier() int
	// GetCurrentTurn いま出す番の席
	GetCurrentTurn() int
	// GetTrick 進行中のトリック
	GetTrick() []*domain.TrickCard
	// GetLastTrick 直前に完成したトリック
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前のトリックを取った席 (-1: まだ無い)
	GetLastTrickWinner() int
	// GetTrickCount このラウンドで完成したトリック数
	GetTrickCount() int
	// GetRoundPoints チームがこのラウンドで取った点
	GetRoundPoints(team int) int
	// GetScore チームの通算得点
	GetScore(team int) int
	// GetWinnerTeam 勝ったチーム (-1: 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt 人数
	GetPlayerCnt() int
	// GetPlayer 席 i のプレイヤー
	GetPlayer(i int) *domain.BotifarraPlayer
	// GetHint 助言
	GetHint() *domain.BotifarraHint
}
