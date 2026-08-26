//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaccaratBanqueWebInput はバカラ・バンクの Web インプット。
type BaccaratBanqueWebInput struct {
	BaseWebInput
	// Config はゲーム設定。
	Config *BaccaratBanqueWebConfig `json:"config,omitempty"`
}

// BaccaratBanqueWebConfig はバカラ・バンクの Web 設定。
type BaccaratBanqueWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	StartChips    *int `json:"startChips,omitempty"`
	BetAmount     *int `json:"betAmount,omitempty"`
}

// BaccaratBanqueWebOutputPlayer は 1 席ぶんの出力。
type BaccaratBanqueWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Role は "banker" | "right" | "left"。
	Role string `json:"role"`
	// Cards は手札。**バカラは全部表向き。**
	Cards []*WebOutputCard `json:"cards"`
	Total int              `json:"total"`
	Chips int              `json:"chips"`
	Bet   int              `json:"bet"`
	Drawn bool             `json:"drawn"`
}

// BaccaratBanqueWebOutputSide は 1 つのタブローとの決着。
type BaccaratBanqueWebOutputSide struct {
	SeatIdx int    `json:"seatIdx"`
	Outcome string `json:"outcome"`
	Bet     int    `json:"bet"`
	Delta   int    `json:"delta"`
}

// BaccaratBanqueWebOutputResult は 1 クーの結果。
type BaccaratBanqueWebOutputResult struct {
	BankerTotal   int                            `json:"bankerTotal"`
	Sides         []*BaccaratBanqueWebOutputSide `json:"sides"`
	BankerDelta   int                            `json:"bankerDelta"`
	BankerNatural bool                           `json:"bankerNatural"`
}

// BaccaratBanqueWebOutput はバカラ・バンクの Web アウトプット。
type BaccaratBanqueWebOutput struct {
	Players    []*BaccaratBanqueWebOutputPlayer `json:"players"`
	Phase      string                           `json:"phase"`
	CoupNumber int                              `json:"coupNumber"`
	// BankHeld はこのバンクで続けたクー数。**1 回負けても途切れない。**
	BankHeld int `json:"bankHeld"`
	// ShoeRemaining はシューの残り枚数。配り切るとバンクが終わる。
	ShoeRemaining int `json:"shoeRemaining"`
	// Retired はバンカーが自分から降りたか。
	Retired     bool                           `json:"retired"`
	LastResult  *BaccaratBanqueWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag bool                           `json:"gameEndFlag"`
	WinnerIdx   int                            `json:"winnerIdx"`
	IsHumanTurn bool                           `json:"isHumanTurn"`
	// HintDraw は引くべきか。
	HintDraw bool `json:"hintDraw"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config BaccaratBanqueWebOutputConfig `json:"config"`
}

// BaccaratBanqueWebOutputConfig は設定アウトプット。
type BaccaratBanqueWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	StartChips    int `json:"startChips"`
	BetAmount     int `json:"betAmount"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *BaccaratBanqueWebConfig) ToConfig() domain.BaccaratBanqueConfig {
	cfg := domain.DefaultBaccaratBanqueConfig()
	cfg.CpuDifficulty = domain.BaccaratBanqueCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.BaccaratBanqueCpuDifficultyEasy),
		int(domain.BaccaratBanqueCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.StartChips, c.StartChips,
		domain.BaccaratBanqueMinChips, domain.BaccaratBanqueMaxChips)
	webutil.ApplyBoundedInt(&cfg.BetAmount, c.BetAmount,
		domain.BaccaratBanqueMinBet, domain.BaccaratBanqueMaxBet)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p BaccaratBanqueWebInput) ToConfig() domain.BaccaratBanqueConfig {
	return configOrDefault(p.Config, (*BaccaratBanqueWebConfig).ToConfig,
		domain.DefaultBaccaratBanqueConfig())
}

// BaccaratBanqueWebController はバカラ・バンクの Web コントローラー。
type BaccaratBanqueWebController = GameWebController[usecase.BaccaratBanqueInteractorIF, BaccaratBanqueWebInput, *BaccaratBanqueWebOutput]

// NewBaccaratBanqueWebController, NewBaccaratBanqueWebControllerWithProvider are
// the standard and provider-backed constructors.
var NewBaccaratBanqueWebController, NewBaccaratBanqueWebControllerWithProvider = webControllerPair[usecase.BaccaratBanqueInteractorIF, BaccaratBanqueWebInput, *BaccaratBanqueWebOutput](
	newBaccaratBanqueDefaultOutput, baccaratBanqueDispatch,
)

func newBaccaratBanqueDefaultOutput(msg string) *BaccaratBanqueWebOutput {
	return &BaccaratBanqueWebOutput{
		Players:       make([]*BaccaratBanqueWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func baccaratBanqueDispatch(bc *baseController, w http.ResponseWriter, di usecase.BaccaratBanqueInteractorIF, param BaccaratBanqueWebInput, _ func(string) *BaccaratBanqueWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	// **引くと止まるは別のコマンド。** ひとつの命令に真偽値を積むと、
	// 既定値のまま届いた要求がどちらかに黙って倒れる。
	case "d", "draw":
		bc.writePresenterResponse(w, di.Draw())
	case "s", "stand":
		bc.writePresenterResponse(w, di.Stand())
	case "nc", "nextcoup":
		bc.writePresenterResponse(w, di.NextCoup())
	case "retire":
		bc.writePresenterResponse(w, di.Retire())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
