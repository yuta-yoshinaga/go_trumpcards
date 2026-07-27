//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TuteGame トゥーテのゲームインタフェース
type TuteGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerDeclareMarriage 人間が結婚宣言する
	PlayerDeclareMarriage(suit int) error
	// PlayerDeclareTute 人間が Tute を宣言する
	PlayerDeclareTute() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TuteConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TuteConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TutePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// IsSuitDeclared 指定スートが結婚宣言済みかを返す
	IsSuitDeclared(suit int) bool
	// GetTeamScores チーム別累積点を取得する
	GetTeamScores() [domain.TuteTeamCnt]int
	// GetRoundTeamPoints 現ラウンドのチーム別得点を取得する
	GetRoundTeamPoints() [domain.TuteTeamCnt]int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TutePlayer
	// CanHumanDeclareMarriage 人間が結婚宣言できるかを返す
	CanHumanDeclareMarriage() bool
	// GetHumanDeclarableMarriageSuits 人間が結婚宣言できる未宣言スート一覧を返す
	GetHumanDeclarableMarriageSuits() []int
	// CanHumanDeclareTute 人間が Tute を宣言できるかを返す
	CanHumanDeclareTute() bool
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.TuteHint
}
