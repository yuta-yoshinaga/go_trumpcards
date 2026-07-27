//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ScartoGame スカルト (Scarto / Piedmontese Tarot) のゲームインタフェース
type ScartoGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerScarto 人間の親がスカルトで 3 枚を捨てる
	PlayerScarto(cardIndices []int) error
	// CpuScarto CPU の親がスカルトで 3 枚を捨てる
	CpuScarto()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ScartoConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ScartoConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ScartoPhase
	// IsHumanTurn 現在のプレイ手番が人間かを返す
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
	// GetDealerIdx 親インデックスを取得する
	GetDealerIdx() int
	// GetScartoCount スカルト札の枚数を取得する
	GetScartoCount() int
	// GetPlayerScores プレイヤー別累積得点を取得する
	GetPlayerScores() [domain.ScartoPlayerCnt]int
	// GetDealScores 直近ディールの精算値を取得する
	GetDealScores() [domain.ScartoPlayerCnt]int
	// GetCardPoints プレイヤー i の獲得ハーフポイントを取得する
	GetCardPoints(i int) int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.ScartoOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.ScartoResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.ScartoPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.ScartoHint
}
