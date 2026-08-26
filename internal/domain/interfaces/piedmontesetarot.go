//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PiedmonteseTarotGame はピエモンテ・タロッコ (Tarocco Piemontese) の
// ゲームインタフェース。
type PiedmonteseTarotGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerScarto 人間の親がタロンぶんを捨てる
	PlayerScarto(cardIndices []int) error
	// CpuScarto CPU の親が捨てる
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
	GetConfig() domain.PiedmonteseTarotConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PiedmonteseTarotConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PiedmonteseTarotPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanScartoTurn 現在のスカルト手番が人間かを返す
	IsHumanScartoTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
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
	// GetScartoCount 親が捨てた札の枚数を取得する
	GetScartoCount() int
	// TalonSize タロン (親が捨てる枚数) を取得する
	TalonSize() int
	// HandSize 1 人の手札枚数を取得する
	HandSize() int
	// GetPlayerScores 累積得点を取得する
	GetPlayerScores() []int
	// GetDealScores 直近ディールの精算値を取得する
	GetDealScores() []int
	// GetCardThirds 指定席の獲得点を 1/3 単位で取得する
	GetCardThirds(i int) int
	// GetLastTrickWinner 最後のトリックを取った席を取得する
	GetLastTrickWinner() int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.PiedmonteseTarotOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.PiedmonteseTarotResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PiedmonteseTarotPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetDiscardableIndices 親がいまスカルトに出せる手札のインデックスを取得する
	GetDiscardableIndices() []int
	// GetHint ヒントを取得する
	GetHint() *domain.PiedmonteseTarotHint
}
