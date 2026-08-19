//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ChemindeFerGame シュマン・ド・フェールゲームインタフェース
type ChemindeFerGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// SetStake 親がバンク額を張る
	SetStake(amount int) error
	// PlaceBet 子が賭ける (0 で降りる)
	PlaceBet(seatIdx, amount int) error
	// PunterDraw 子側が 3 枚目を引く
	PunterDraw() error
	// PunterStand 子側が立つ
	PunterStand() error
	// BankerDraw 親が 3 枚目を引く
	BankerDraw() error
	// BankerStand 親が立つ
	BankerStand() error
	// PassBank 親が自分から親を降りる
	PassBank() error
	// NextRound 次のラウンドを始める
	NextRound() error
	// GiveUp 投了する
	GiveUp()
	// CpuPlay CPUの手番を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ChemindeFerConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ChemindeFerConfig)

	// GetPhase 現在のフェーズ
	GetPhase() domain.ChemindeFerPhase
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// IsHumanTurn 人間の入力を待っているか
	IsHumanTurn() bool

	// GetBankerIdx 親の席
	GetBankerIdx() int
	// GetBetTurn 次に賭ける子の席 (-1: 賭けは終わっている)
	GetBetTurn() int
	// GetStake 親が張ったバンク額
	GetStake() int
	// GetRemainingStake まだ覆われていないバンク額
	GetRemainingStake() int
	// GetTotalBet 子側の賭け総額
	GetTotalBet() int
	// StakeRangeFor 席が親として張れる額の範囲
	StakeRangeFor(seatIdx int) (int, int)
	// BetRangeFor 席がいま賭けられる額の範囲
	BetRangeFor(seatIdx int) (int, int)
	// GetRepresentativeIdx 子側の代表の席 (-1: 未定)
	GetRepresentativeIdx() int
	// PunterMayChoose 子側がいま自分で引き方を選べるか
	PunterMayChoose() bool

	// GetBankerHand 親の手札
	GetBankerHand() []*domain.Card
	// GetPunterHand 子側の手札
	GetPunterHand() []*domain.Card
	// GetBankerTotal 親の合計値
	GetBankerTotal() int
	// GetPunterTotal 子側の合計値
	GetPunterTotal() int
	// GetPunterDrew 子側が 3 枚目を引いたか
	GetPunterDrew() bool
	// GetResult ラウンドの決着
	GetResult() domain.ChemindeFerResult
	// GetLastNet 直前の決済での席の純増減 (ラウンド中は 0)
	GetLastNet(i int) int

	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetPlayer 席 i のプレイヤー
	GetPlayer(i int) *domain.ChemindeFerPlayer
	// GetRemainingCards シューの残り枚数
	GetRemainingCards() int
	// GetHint 助言
	GetHint() *domain.ChemindeFerHint
}
