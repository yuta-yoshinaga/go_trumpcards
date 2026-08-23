//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// JulepeGame フレペ (Julepe) ゲームインタフェース
type JulepeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Decide 参加 (true) するか降りる (false) かを選ぶ
	Decide(play bool) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.JulepeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.JulepeConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.JulepePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsDecidePhase 参加するかどうかの選択中かを返す
	IsDecidePhase() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetPot 場に積まれているチップを取得する
	GetPot() int
	// GetRequiredTricks 現在の参加人数に対する規定トリック数を取得する
	GetRequiredTricks() int
	// GetBeast 次ラウンドのアンティが倍になる席を取得する
	GetBeast() []bool
	// GetTrumpSuit 切り札のスートを取得する
	GetTrumpSuit() int
	// GetUpCard 切り札を決めた表向きの1枚を取得する
	GetUpCard() *domain.Card
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetActiveCount このラウンドの参加者数を取得する
	GetActiveCount() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.JulepePlayer
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.JulepeHint
}
