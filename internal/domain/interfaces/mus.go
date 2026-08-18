//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MusGame ムスのゲームインタフェース
type MusGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerMus Mus(true)/Corte(false)を宣言する
	PlayerMus(mus bool) error
	// PlayerDiscard 交換する札を選ぶ
	PlayerDiscard(indices []int) error
	// PlayerBet 賭けアクションを実行する
	PlayerBet(action, amount int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// Showdown 受理・流局ラウンドを精算する
	Showdown()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MusConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MusConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MusPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetManoIdx 親インデックスを取得する
	GetManoIdx() int
	// GetMusTurn Mus宣言中のプレイヤーインデックスを取得する
	GetMusTurn() int
	// GetDiscardTurn 交換中のプレイヤーインデックスを取得する
	GetDiscardTurn() int
	// GetBetTeam 現在賭けアクションするチームを取得する
	GetBetTeam() int
	// GetPendingStake 応答待ちの賭け額を取得する (-1=ordago, 0=なし)
	GetPendingStake() int
	// GetLastBettorTeam 直近に賭けたチームを取得する
	GetLastBettorTeam() int
	// GetAmarrakos チーム別累積点を取得する
	GetAmarrakos() [domain.MusTeamCnt]int
	// GetResult ラウンド ri の結果を取得する
	GetResult(ri int) domain.MusRoundResult
	// GetMusCycle Mus交換の繰り返し回数を取得する
	GetMusCycle() int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MusPlayer
	// GetHandSummary 指定プレイヤーの手役評価 (Grande/Chica/Pares/Juego) を取得する
	GetHandSummary(i int) *domain.MusHandSummary
	// GetHint ヒントを取得する
	GetHint() *domain.MusHint
}
