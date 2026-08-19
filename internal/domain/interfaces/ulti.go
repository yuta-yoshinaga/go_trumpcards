//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// UltiGame ウルティ (Ulti) のゲームインタフェース
type UltiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間がコントラクトを宣言する (party は切り札スートも)
	PlayerBid(contract domain.UltiContract, trumpSuit int) error
	// CpuBid CPU の宣言 (非競争のため no-op)
	CpuBid()
	// PlayerDiscard デクレアラーがタロン受け取り後に 2 枚を捨てる
	PlayerDiscard(cardIndices []int) error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.UltiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.UltiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.UltiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の宣言手番が人間かを返す
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
	// GetDeclarerIdx デクレアラーインデックスを取得する
	GetDeclarerIdx() int
	// GetContract コントラクトを取得する
	GetContract() domain.UltiContract
	// GetTrumpSuit 切り札スートを取得する (-1=なし, 1..4)
	GetTrumpSuit() int
	// GetTalonCount タロンの残り枚数を取得する
	GetTalonCount() int
	// GetTalonTaken タロンを拾ったかを取得する
	GetTalonTaken() bool
	// GetDiscardCount 捨札枚数を取得する
	GetDiscardCount() int
	// GetPlayerCoins プレイヤー別累積コインを取得する
	GetPlayerCoins() [domain.UltiPlayerCnt]int
	// GetLastDealCoins 直近ディールの精算による符号付きコイン増減を取得する
	GetLastDealCoins() [domain.UltiPlayerCnt]int
	// GetCardPoints プレイヤー i の獲得カードポイントを取得する
	GetCardPoints(i int) int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.UltiOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.UltiResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.UltiPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.UltiHint
}
