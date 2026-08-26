//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CoincheGame コワンシュゲームインタフェース
type CoincheGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが目標点と切り札スートを宣言する
	PlayerBid(points, suit int) error
	// PlayerPassBid 人間プレイヤーが競りでパスする
	PlayerPassBid() error
	// CpuBid CPUプレイヤーが競りで宣言またはパスする
	CpuBid()
	// PlayerCoinche 守備側が倍化する
	PlayerCoinche() error
	// PlayerSurcoinche 宣言側が再倍化する
	PlayerSurcoinche() error
	// PlayerDeclineDouble 倍化せずに進める
	PlayerDeclineDouble() error
	// CpuDouble CPUプレイヤーが倍化判断する
	CpuDouble()
	// GetContractPoints 落札された目標点を取得する (0 = 未落札)
	GetContractPoints() int
	// GetDouble 倍率の状態を取得する
	GetDouble() domain.CoincheDouble
	// GetMultiplier 現在の得点倍率を取得する
	GetMultiplier() int
	// GetBiddablePoints 今この席が宣言できる目標点を取得する
	GetBiddablePoints() []int
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CoincheConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CoincheConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CoinchePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
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
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int

	// GetMakerTeam メイカーチームを取得する
	GetMakerTeam() int
	// GetMakerPlayerIdx メイカープレイヤーを取得する
	GetMakerPlayerIdx() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetRoundPoints 当ラウンドのチーム別カード点数を取得する
	GetRoundPoints(team int) int
	// GetRoundBeloteBonus 当ラウンドの Belote/Rebelote ボーナスを取得する
	GetRoundBeloteBonus(team int) int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CoinchePlayer
	// GetHint ヒントを取得する
	GetHint() *domain.CoincheHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
