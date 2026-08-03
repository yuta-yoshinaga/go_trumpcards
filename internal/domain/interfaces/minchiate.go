//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MinchiateGame ミンキアーテのゲームインタフェース
type MinchiateGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerScarto 人間のディーラーが余剰札を捨てる
	PlayerScarto(cardIndices []int) error
	// CpuScarto CPU のディーラーが自動で捨てる
	CpuScarto()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドを締め、規定局数ならマッチを終える
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MinchiateConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MinchiateConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MinchiatePhase
	// IsHumanTurn 現在の手番が人間かを返す (プレイフェーズ)
	IsHumanTurn() bool
	// IsHumanScartoTurn 現在のスカルト手番が人間かを返す
	IsHumanScartoTurn() bool
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
	// GetLastTrickWinner 直前のトリック勝者を取得する (-1=未確定)
	GetLastTrickWinner() int
	// GetScartoSize ディーラーが捨てた枚数を取得する
	GetScartoSize() int
	// GetTeamScores チーム別の累計得点を取得する
	GetTeamScores() [2]int
	// GetRoundTricks 現ラウンドの席別獲得トリック数を取得する
	GetRoundTricks() [domain.MinchiatePlayerCnt]int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定/同点)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MinchiatePlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.MinchiateHint
}
