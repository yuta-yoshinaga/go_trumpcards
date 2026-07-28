//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// WattenGame ヴァッテンゲームインタフェース
type WattenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerDeclare 人間ディーラーが Schlag ランク + 切り札スートを宣言する
	PlayerDeclare(rank, suit int) error
	// CpuDeclare CPUディーラーが宣言する
	CpuDeclare()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// PlayerRaise 人間プレイヤーがステークを引き上げる
	PlayerRaise() error
	// PlayerRespond 人間プレイヤーがレイズに応答する (true=hold / false=fold)
	PlayerRespond(hold bool) error
	// CpuRespond CPUが応答フェーズで hold/fold を判断する
	CpuRespond()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.WattenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.WattenConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.WattenPhase
	// IsHumanTurn 現在のプレイ手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanDeclareTurn 現在の宣言手番が人間かを返す
	IsHumanDeclareTurn() bool
	// IsHumanRespondTurn 現在の応答手番が人間かを返す
	IsHumanRespondTurn() bool
	// CanHumanRaise 現在の手番の人間がレイズ可能かを返す
	CanHumanRaise() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetSchlagRank Schlag ランクを取得する
	GetSchlagRank() int
	// GetCriticalSuit 切り札スートを取得する
	GetCriticalSuit() int
	// GetStake 現在の確定ステークを取得する
	GetStake() int
	// GetPendingStake 応答待ちで提示中のステークを取得する
	GetPendingStake() int
	// GetRaiseCount 確定済みレイズ回数を取得する
	GetRaiseCount() int
	// GetRaiserTeam 応答待ちのレイズ実施チームを取得する
	GetRaiserTeam() int
	// GetResponderIdx 応答すべきプレイヤーインデックスを取得する
	GetResponderIdx() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetTeamTricks 当ディールのチーム別トリック数を取得する
	GetTeamTricks(team int) int
	// GetDealWinnerTeam 直近ディールの勝者チームを取得する
	GetDealWinnerTeam() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.WattenResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.WattenPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.WattenHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
