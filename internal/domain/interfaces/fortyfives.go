//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FortyFivesGame オークション・フォーティファイブズのゲームインタフェース
type FortyFivesGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが入札する
	PlayerBid(bid domain.FortyFivesBid) error
	// CpuBid CPUプレイヤーが1件入札する
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
	GetConfig() domain.FortyFivesConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.FortyFivesConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FortyFivesPhase
	// IsHumanTurn 現在の手番が人間かを返す (プレイフェーズ)
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の入札手番が人間かを返す
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
	// GetDeclarerIdx 落札者インデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetContract 確定した契約を取得する
	GetContract() domain.FortyFivesBid
	// GetBids 各プレイヤーの入札を取得する
	GetBids() [domain.FortyFivesPlayerCnt]domain.FortyFivesBid
	// GetBidDone 各プレイヤーが入札を済ませたかを取得する (未入札とパスの区別用)
	GetBidDone() [domain.FortyFivesPlayerCnt]bool
	// GetTrumpSuit 切り札スートを取得する (0=なし)
	GetTrumpSuit() int
	// GetTeamScores チーム別累積点を取得する
	GetTeamScores() [domain.FortyFivesTeamCnt]int
	// GetContractProgress 落札チームの契約達成見込みを取得する
	GetContractProgress() *domain.FortyFivesContractProgress
	// GetRoundTeamPoints 現ラウンドのチーム別得点を取得する
	GetRoundTeamPoints() [domain.FortyFivesTeamCnt]int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FortyFivesPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.FortyFivesHint
}
