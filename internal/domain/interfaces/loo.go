//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// LooGame はルー (Loo / Lanterloo) のゲームインタフェース。
type LooGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerDecide 人間が参加 (play) / 降り (pass) を決める
	PlayerDecide(play bool) error
	// CpuDecide CPUプレイヤーが1回 play/pass を決める
	CpuDecide()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ステップ実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールのポットを精算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.LooConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.LooConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.LooPhase
	// IsHumanTurn 現在の意思決定者が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetDealerIdx 親インデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetDecidePlayerIdx 現在 play/pass を決めるプレイヤーインデックスを取得する
	GetDecidePlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する (-1=なし)
	GetLastTrickWinner() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetTrumpSuit 切り札スートを取得する (0=未確定)
	GetTrumpSuit() int
	// GetTurnUp めくり札を取得する (nil の場合もある)
	GetTurnUp() *domain.Card
	// GetPot 現在のポット額を取得する
	GetPot() int
	// GetPotStart 現ディール開始時のポット額を取得する
	GetPotStart() int
	// GetLastDealDetail 直前ディールの精算内訳を取得する
	GetLastDealDetail() *domain.LooDealDetail
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.LooPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.LooHint
}
