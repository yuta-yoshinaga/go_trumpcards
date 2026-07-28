//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MichiganGame はミシガン (Michigan / Newmarket) のゲームインタフェース。
type MichiganGame interface {
	BaseGame
	// Reset ゲームを初期化する (新規ゲーム)
	Reset()
	// NextRound 次のラウンドを配る
	NextRound()
	// PlaceHumanBet 人間 (seat 0) のブードル賭けを適用する
	PlaceHumanBet(bets []int) error
	// PlayCard 人間 (現在の手番) が手札インデックスのカードを出す
	PlayCard(idx int) error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MichiganConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MichiganConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MichiganPhase
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetDealerIdx ディーラーの座席番号を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 現在の手番プレイヤーの座席番号を取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx このラウンドの最初のリード席を取得する
	GetLeadPlayerIdx() int
	// GetAnte アンティ額を取得する
	GetAnte() int
	// GetBetBudget 人間が今ラウンド賭けるべき額を取得する
	GetBetBudget() int
	// GetHumanBetPlaced 人間が今ラウンドの賭けを済ませたかを取得する
	GetHumanBetPlaced() bool
	// GetBoodleCnt ブードルの数を取得する (常に 4)
	GetBoodleCnt() int
	// GetBoodle 指定インデックスのブードルを取得する
	GetBoodle(i int) *domain.MichiganBoodle
	// GetSeqSuit 現在のシーケンスのスートを取得する (0 = 新シーケンス待ち)
	GetSeqSuit() int
	// GetSeqHighValue 現在のシーケンスの最大値を取得する (0 = なし)
	GetSeqHighValue() int
	// GetDeadHandCount デッドハンドの枚数を取得する
	GetDeadHandCount() int
	// GetWinnerIdx 直近ラウンドで手札を出し切ったプレイヤーを取得する (-1 = なし)
	GetWinnerIdx() int
	// GetMatchWinnerIdx ゲーム全体の勝者を取得する (-1 = 未確定)
	GetMatchWinnerIdx() int
	// GetResult 人間から見たラウンド結果を取得する
	GetResult() domain.MichiganResult
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MichiganPlayer
	// GetChips 人間 (seat 0) の保有チップを取得する
	GetChips() int
	// IsHumanTurn 現在人間の入力待ちかどうかを返す
	IsHumanTurn() bool
	// GetPlayableIndices 人間が今出せる手札インデックス列を返す
	GetPlayableIndices() []int
	// GetHint ヒントを取得する
	GetHint() *domain.MichiganHint
}
