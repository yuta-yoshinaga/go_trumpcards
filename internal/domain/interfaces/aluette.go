//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AluetteGame アリュエットのゲームインタフェース
//
// **スカルトも入札も無い。**タロー系から写すと PlayerScarto / IsHumanScartoTurn
// が付いてくるが、アリュエットには余剰札を伏せる工程が無い。
type AluetteGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のメーヌを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound メーヌを締め、規定点に達していればマッチを終える
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.AluetteConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.AluetteConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AluettePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のメーヌ番号を取得する
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
	// GetTeamScores チーム別の累計メーヌ数を取得する
	GetTeamScores() [2]int
	// GetRoundTricks 現メーヌの席別獲得トリック数を取得する
	GetRoundTricks() [domain.AluettePlayerCnt]int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定/同点)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.AluettePlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.AluetteHint
}
