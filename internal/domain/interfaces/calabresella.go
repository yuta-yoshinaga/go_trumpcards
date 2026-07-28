//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CalabresellaGame カラブレセッラ (Calabresella) のゲームインタフェース
type CalabresellaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間がビッドする (pass/chiamo/solo)
	PlayerBid(bid domain.CalabresellaBid) error
	// CpuBid CPUプレイヤーが1回ビッドする
	CpuBid()
	// PlayerDiscard 人間ソリストが monte 交換で1枚を捨てる
	PlayerDiscard(cardIndex int) error
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
	GetConfig() domain.CalabresellaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CalabresellaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CalabresellaPhase
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
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetForehandIdx forehand インデックスを取得する
	GetForehandIdx() int
	// GetSoloistIdx ソリストインデックスを取得する (-1=未確定)
	GetSoloistIdx() int
	// GetWinningBid 確定ビッドを取得する
	GetWinningBid() domain.CalabresellaBid
	// GetCurrentBidderIdx 現在のビッド手番インデックスを取得する
	GetCurrentBidderIdx() int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.CalabresellaPlayerCnt]int
	// GetRoundThirds 現ラウンドのプレイヤー別 1/3 点を取得する
	GetRoundThirds() [domain.CalabresellaPlayerCnt]int
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CalabresellaPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.CalabresellaHint
}
