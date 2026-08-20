//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TrenteEtQuaranteGame はトラント・エ・カラント (Trente et Quarante / Rouge et Noir) の
// ゲームインタフェース。
type TrenteEtQuaranteGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 同じシュー・チップで次のラウンドを始める
	NextRound()
	// PlaceBet ベット種別とステークを賭け、両列を配って解決する
	PlaceBet(bet domain.TrenteEtQuaranteBet, stake int) error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TrenteEtQuaranteConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TrenteEtQuaranteConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TrenteEtQuarantePhase
	// GetGameEndFlag ラウンド終了フラグを取得する
	GetGameEndFlag() bool
	// GetCurrentBet 現在のベット種別を取得する
	GetCurrentBet() domain.TrenteEtQuaranteBet
	// GetStake 現在のステークを取得する
	GetStake() int
	// GetNoirRow 黒 (Noir) 列の札を取得する
	GetNoirRow() []*domain.Card
	// GetRougeRow 赤 (Rouge) 列の札を取得する
	GetRougeRow() []*domain.Card
	// GetNoirTotal 黒 (Noir) 列の合計を取得する
	GetNoirTotal() int
	// GetRougeTotal 赤 (Rouge) 列の合計を取得する
	GetRougeTotal() int
	// GetWinningRow 勝ち列 (None/Noir/Rouge) を取得する
	GetWinningRow() int
	// GetFirstCardRed 最初に配られた札が赤かどうかを取得する
	GetFirstCardRed() bool
	// GetRefait Refait (31 の同点) だったかどうかを取得する
	GetRefait() bool
	// GetResult 現在のベットに対する勝敗結果を取得する
	GetResult() domain.TrenteEtQuaranteResult
	// GetPayout このラウンドでチップに戻された総額を取得する
	GetPayout() int
	// GetChips 保有チップ数を取得する
	GetChips() int
	// GetRoundNumber 解決したラウンド数を取得する
	GetRoundNumber() int
	// GetRemainingDeck シューの残り枚数を取得する
	GetRemainingDeck() int
	// GetPlayer プレイヤーを取得する
	GetPlayer() *domain.TrenteEtQuarantePlayer
	// GetHint ヒントを取得する
	GetHint() *domain.TrenteEtQuaranteHint
}
