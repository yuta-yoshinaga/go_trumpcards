//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CucumberGame キューカンバーゲームインタフェース
type CucumberGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay 人間が1枚出す
	PlayerPlay(cardIndex int) error
	// NextRound ラウンド終了状態から次を配る
	NextRound() error
	// CpuPlay CPUが1枚出す
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CucumberConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CucumberConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CucumberPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetValidPlayIndices 出せる手札の位置を返す
	GetValidPlayIndices(playerIdx int) []int
	// HighestInTrick いまトリックに出ている最高ランクを返す (0: まだ無い)
	HighestInTrick() int
	// IsForcedLowest 更新できないので低い札に決まっているかを返す
	//
	// **各層で判定し直さないこと。** 「合法手が 1 つ = 更新できない」は偽です。
	IsForcedLowest(playerIdx int) bool
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リード席を取得する
	GetLeadPlayerIdx() int
	// GetTrickNumber 解決済みのトリック数を取得する
	GetTrickNumber() int
	// GetRoundNumber ラウンド数を取得する
	GetRoundNumber() int
	// GetLastTrickWinnerIdx 直前ラウンドで最終トリックを取った席 (-1: 未)
	GetLastTrickWinnerIdx() int
	// GetLastPenalty 直前ラウンドで付いた失点を取得する
	GetLastPenalty() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CucumberPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.CucumberHint
}
