//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// UnsunKarutaGame はうんすんカルタ (八人メリ) のゲームインタフェース。
type UnsunKarutaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerPlay 人間がカードを出す (リードなら declare でメリ/モンチを宣言)
	PlayerPlay(cardIndex int, declare bool) error
	// CpuPlay CPU が 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールを集計する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.UnsunKarutaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.UnsunKarutaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.UnsunKarutaPhase
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// CanDeclare 人間がいま宣言できるかを取得する
	CanDeclare() bool
	// IsMustFollow いまのトリックにフォロー義務があるかを取得する
	IsMustFollow() bool
	// IsDeclaredThisTrick いまのトリックで宣言が行われたかを取得する
	IsDeclaredThisTrick() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 場の札を取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードした席を取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// TrumpCard 表に返した切り札札を取得する
	TrumpCard() *domain.Card
	// GetTeamTricks ディール中のチーム別「コ」を取得する
	GetTeamTricks() []int
	// GetTeamScores マッチ累計を取得する
	GetTeamScores() []int
	// GetLastTrickWinner 直前のトリックを取った席を取得する
	GetLastTrickWinner() int
	// GetResult 人間チームから見たマッチ結果を取得する
	GetResult() domain.UnsunKarutaResult
	// GetWinnerTeam 勝ったチームを取得する (-1=引き分け)
	GetWinnerTeam() int
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.UnsunKarutaPlayer
	// GetPlayableIndices 出せる札のインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.UnsunKarutaHint
}
