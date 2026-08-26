//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DilotiGame はディロティのゲームインタフェース。
type DilotiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次の局を開始する
	NextRound()
	// PlayerPlay 人間が 1 手打つ
	PlayerPlay(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) error
	// CpuPlay CPU が 1 手打つ
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DilotiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DilotiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetTable 場の緩い札を取得する
	GetTable() []*domain.Card
	// GetDeclarations 場に積まれた宣言を取得する
	GetDeclarations() []*domain.DilotiDeclaration
	// GetRoundNumber 現在の局を取得する
	GetRoundNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetLastCapturer 最後に取った席を取得する (-1 = なし)
	GetLastCapturer() int
	// GetDeckRemaining 山の残り枚数を取得する
	GetDeckRemaining() int
	// GetTakeOptions 手札を出したときの取り手を取得する
	GetTakeOptions(seat, handIdx int) []domain.DilotiTake
	// GetDeclareOptions 手札で作れる新しい宣言を取得する
	GetDeclareOptions(seat, handIdx int) []domain.DilotiDeclCandidate
	// CanTrail その手札を場に置けるかを取得する
	CanTrail(seat, handIdx int) bool
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.DilotiPlayer
	// GetLastResult 直前の局の集計を取得する
	GetLastResult() *domain.DilotiRoundResult
	// GetWinnerIdx 勝者の席を取得する (-1 = 未決)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.DilotiHint
}
