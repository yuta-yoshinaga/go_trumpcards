//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TarabishGame タラビッシュ (Tarabish) ゲームインタフェース
type TarabishGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// TakeTrump 表向きの札のスートを切り札として引き受ける
	TakeTrump() error
	// PassTrump 切り札を見送る（親は見送れない）
	PassTrump() error
	// CpuBid 手番の CPU が切り札を引き受けるか決める
	CpuBid()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TarabishConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TarabishConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TarabishPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 人間が切り札の選択をする番かを返す
	IsHumanBidTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetTrumpSuit 切り札のスートを取得する
	GetTrumpSuit() int
	// GetUpCard 切り札候補として表向きにした1枚を取得する
	GetUpCard() *domain.Card
	// GetTrumpTakerIdx 切り札を引き受けたプレイヤーを取得する (-1: 未決定)
	GetTrumpTakerIdx() int
	// GetScore チームの累計得点を取得する
	GetScore(team int) int
	// GetRoundPoints チームの現ラウンド点を取得する
	GetRoundPoints(team int) int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TarabishPlayer
	// GetWinnerTeam 勝利チームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.TarabishHint
}
