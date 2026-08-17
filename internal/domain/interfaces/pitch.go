package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PitchGame ピッチゲームインタフェース
type PitchGame interface {
	// GetRoundBreakdown 直近ラウンドの High/Low/Jack/Game をそれぞれ誰が取ったか
	GetRoundBreakdown() domain.PitchRoundBreakdown
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする (0=pass, 2..4)
	PlayerBid(bid int) error
	// CpuBid CPUプレイヤーがビッドする
	CpuBid()
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
	GetConfig() domain.PitchConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PitchConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PitchPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetDealerIdx ディーラー (親) のインデックスを取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx 現在ビッド中のプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetCurrentBid 現在の最高ビッド値を取得する
	GetCurrentBid() int
	// GetBidWinnerIdx 最高ビッダーのインデックスを取得する
	GetBidWinnerIdx() int
	// GetTrumpSuit 切り札スートを取得する (PitchTrumpUnset=未確定)
	GetTrumpSuit() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PitchPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.PitchHint
	// GetValidPlayIndices プレイ可能なカードのインデックスを返す
	GetValidPlayIndices(playerIdx int) []int
}
