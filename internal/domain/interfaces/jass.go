//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// JassGame ヤス(シーバー)ゲームインタフェース
type JassGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerChooseTrump 人間プレイヤーが切り札スートを指名する
	PlayerChooseTrump(suit int) error
	// PlayerSchieben 人間プレイヤー(フォアハンド)がパートナーへ Schieben する
	PlayerSchieben() error
	// CpuBid CPUプレイヤーがビッド判断する
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
	GetConfig() domain.JassConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.JassConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.JassPhase
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
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetForehandIdx フォアハンドインデックスを取得する
	GetForehandIdx() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetSchieben Schieben 状態を取得する
	GetSchieben() bool
	// GetMakerTeam メイカーチームを取得する
	GetMakerTeam() int
	// GetMakerPlayerIdx メイカープレイヤーを取得する
	GetMakerPlayerIdx() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetRoundPoints 当ラウンドのチーム別カード点数を取得する
	GetRoundPoints(team int) int
	// GetRoundWeisPoints 当ラウンドの Weis 得点を取得する
	GetRoundWeisPoints(team int) int
	// GetRoundStockPoints 当ラウンドの Stöck 得点を取得する
	GetRoundStockPoints(team int) int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.JassPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.JassHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
