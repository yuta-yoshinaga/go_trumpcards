//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuinzeWebInput カーンズ Web インプット
type QuinzeWebInput struct {
	BaseWebInput
	// Amount is the stake for "bet".
	Amount *int `json:"amount,omitempty"`
}

// QuinzeWebOutputHand 1 つの手の出力
type QuinzeWebOutputHand struct {
	Cards []*WebOutputCard `json:"cards"`
	Bet   int              `json:"bet"`
	// TotalHalves は合計を**半点単位**で表したもの。0.5 点札があるので、
	// 小数をワイヤに載せずに正確な等値比較ができる形で渡す。
	TotalHalves int `json:"totalHalves"`
	// TotalLabel は "15" のような表示用の文字列。
	TotalLabel string `json:"totalLabel"`
	Stood      bool   `json:"stood"`
	Payout     int    `json:"payout"`
	// Hidden が真のとき、Cards の要素は null で合計も伏せられる。枚数だけが残る。
	Hidden bool `json:"hidden"`
}

// QuinzeWebOutputSeat 1 席の出力
type QuinzeWebOutputSeat struct {
	Name  string               `json:"name"`
	IsCPU bool                 `json:"isCpu"`
	Hand  *QuinzeWebOutputHand `json:"hand,omitempty"`
}

// QuinzeWebOutput カーンズ Web アウトプット
type QuinzeWebOutput struct {
	Seats         []*QuinzeWebOutputSeat `json:"seats"`
	BankerHand    *QuinzeWebOutputHand   `json:"bankerHand,omitempty"`
	BankerIdx     int                    `json:"bankerIdx"`
	IsHumanBanker bool                   `json:"isHumanBanker"`
	Chips         int                    `json:"chips"`
	ActiveSeat    int                    `json:"activeSeat"`
	NextBanker    int                    `json:"nextBanker"`
	LastResult    string                 `json:"lastResult"`
	Phase         int                    `json:"phase"`
	// Target は 15 を半点単位で表したもの（15）。
	Target int `json:"targetHalves"`
	// CpuStandPoints は CPU 席と親が止まる合計（11 = 5.5 点、#5566）。
	// 数字を訳文に焼き込むと、閾値を変えたとき案内だけが嘘になる。
	CpuStandPoints int  `json:"cpuStandHalves"`
	CanHit         bool `json:"canHit"`
	CanStand       bool `json:"canStand"`
	WebOutputBase
}

// QuinzeWebController カーンズ Web コントローラークラス
type QuinzeWebController = GameWebController[usecase.QuinzeInteractorIF, QuinzeWebInput, *QuinzeWebOutput]

// NewQuinzeWebController and NewQuinzeWebControllerWithProvider are the
// standard and provider-backed constructors for QuinzeWebController.
var NewQuinzeWebController, NewQuinzeWebControllerWithProvider = webControllerPair[usecase.QuinzeInteractorIF, QuinzeWebInput, *QuinzeWebOutput](
	newQuinzeDefaultOutput, quinzeDispatch,
)

func newQuinzeDefaultOutput(msg string) *QuinzeWebOutput {
	return &QuinzeWebOutput{
		Seats:         make([]*QuinzeWebOutputSeat, 0),
		NextBanker:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func quinzeDispatch(bc *baseController, w http.ResponseWriter, si usecase.QuinzeInteractorIF, param QuinzeWebInput, newDefault func(string) *QuinzeWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newDefault, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bet(*param.Amount))
	case "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "h", "hit":
		bc.writePresenterResponse(w, si.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, si.Stand())
	case "bh", "bankerhit":
		bc.writePresenterResponse(w, si.BankerHit())
	case "bs", "bankerstand":
		bc.writePresenterResponse(w, si.BankerStand())
	case "log", "l":
		bc.writePresenterResponse(w, si.ActionLog())
	case "r", "reset":
		bc.writePresenterResponse(w, si.Reset())
	default:
		return false
	}
	return true
}
