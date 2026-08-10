//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RookGame ルーク(Rook)ゲームインタフェース
type RookGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーがビッドする
	PlayerBid(bid int) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPUプレイヤーが1ビッド実行する
	CpuBid()
	// PlayerExchangeNest 人間(落札者)が5枚捨て切り札色を宣言する
	PlayerExchangeNest(discardIndices []int, trumpColor int) error
	// CpuExchange CPU(落札者)が5枚捨て切り札色を宣言する
	CpuExchange()
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
	GetConfig() domain.RookConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.RookConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.RookPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
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
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッド手番インデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpColor 切り札色を取得する (-1 = なし)
	GetTrumpColor() int
	// GetContractBid 落札ビッドを取得する
	GetContractBid() int
	// GetDeclarerIdx 落札者インデックスを取得する (-1 = 未確定)
	GetDeclarerIdx() int
	// GetHighestBid 現在の最高ビッドを取得する (0 = なし)
	GetHighestBid() int
	// GetHighestBidder 最高ビッダーのインデックスを取得する (-1 = なし)
	GetHighestBidder() int
	// GetNest ネストを取得する
	GetNest() []*domain.Card
	// GetNestPoints ネストの得点を取得する
	GetNestPoints() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetTeamPoints チームがこのラウンドで獲得した得点札の合計を取得する
	GetTeamPoints(team int) int
	// GetWinnerTeam 勝利チームを取得する (-1 = 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.RookPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.RookHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
