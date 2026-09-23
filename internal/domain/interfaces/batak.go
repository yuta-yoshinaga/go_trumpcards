package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BatakGame Batak ゲームインタフェース
type BatakGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする
	PlayerBid(bid int) error
	// CpuBid CPU プレイヤーがビッドする
	CpuBid()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BatakConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BatakConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BatakPhase
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
	// GetSpadesBroken スペードブレイク済みかを返す
	GetSpadesBroken() bool
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetDeclarerIdx 親インデックスを取得する (-1 = 未確定)
	GetDeclarerIdx() int
	// GetHighBid 現在の最高ビッドを取得する (0 = 未宣言)
	GetHighBid() int
	// GetBidStartIdx ビッド開始席インデックスを取得する
	GetBidStartIdx() int
	// MinLegalBid 現在発言可能な最小ビッドを返す (0 = パスのみ可能)
	MinLegalBid() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BatakPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.BatakHint
	// GetValidPlayIndices プレイ可能なカードのインデックスを返す
	GetValidPlayIndices(playerIdx int) []int
}
