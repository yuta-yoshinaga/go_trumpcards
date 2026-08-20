//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TysiacGame サウザンド (Tysiąc) のゲームインタフェース
type TysiacGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間がビッドする (raise=true で +10、false でパス)
	PlayerBid(raise bool) error
	// CpuBid CPUプレイヤーが1回ビッドする
	CpuBid()
	// PlayerDiscard 人間 declarer が talon 交換で1枚を相手へ渡す
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
	GetConfig() domain.TysiacConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TysiacConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TysiacPhase
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
	// GetDeclarerIdx declarer インデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetContract 確定 contract を取得する
	GetContract() int
	// GetCurrentBid 現在の最高ビッドを取得する
	GetCurrentBid() int
	// GetTrumpSuit 切り札スートを取得する (0=未設定)
	GetTrumpSuit() int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.TysiacPlayerCnt]int
	// GetRoundCardPoints 現ラウンドのプレイヤー別カード得点を取得する
	GetRoundCardPoints() [domain.TysiacPlayerCnt]int
	// GetRoundMarriage 現ラウンドのプレイヤー別結婚点を取得する
	GetRoundMarriage() [domain.TysiacPlayerCnt]int
	// GetMarriageOptions いま結婚を宣言できるスートとその点を取得する
	GetMarriageOptions(playerIdx int) []domain.TysiacMarriageOption
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TysiacPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.TysiacHint
}
