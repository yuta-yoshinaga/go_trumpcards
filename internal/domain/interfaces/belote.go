//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BeloteGame ベロートゲームインタフェース
type BeloteGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPickUp 人間プレイヤーがピックアップ判断する
	PlayerPickUp(orderUp bool) error
	// CpuPickUp CPUプレイヤーがピックアップ判断する
	CpuPickUp()
	// PlayerCallTrump 人間プレイヤーがスートを指名する
	PlayerCallTrump(suit int) error
	// PlayerPassCall 人間プレイヤーがコールフェーズでパスする
	PlayerPassCall() error
	// CpuCallTrump CPUプレイヤーがコールトランプ判断する
	CpuCallTrump()
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
	GetConfig() domain.BeloteConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BeloteConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BelotePhase
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
	// GetFaceUpCard 表向きカードを取得する
	GetFaceUpCard() *domain.Card
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
	GetPlayer(i int) *domain.BelotePlayer
	// GetHint ヒントを取得する
	GetHint() *domain.BeloteHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
