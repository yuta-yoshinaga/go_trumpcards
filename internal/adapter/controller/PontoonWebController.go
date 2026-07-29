//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PontoonWebInput ポンツーン Web インプット
type PontoonWebInput struct {
	BaseWebInput
	// Amount is the stake for "bet" and the extra stake for "buy".
	Amount *int `json:"amount,omitempty"`
}

// PontoonWebOutputHand 1 つの手の出力
type PontoonWebOutputHand struct {
	Cards []*WebOutputCard `json:"cards"`
	Bet   int              `json:"bet"`
	Total int              `json:"total"`
	// Rank は 0=バースト / 1=点数 / 2=ファイブカード / 3=ポンツーン。
	Rank    int  `json:"rank"`
	Twisted bool `json:"twisted"`
	Stuck   bool `json:"stuck"`
	Payout  int  `json:"payout"`
}

// PontoonWebOutputSeat 1 席の出力
type PontoonWebOutputSeat struct {
	Name  string                  `json:"name"`
	IsCPU bool                    `json:"isCpu"`
	Hands []*PontoonWebOutputHand `json:"hands"`
}

// PontoonWebOutput ポンツーン Web アウトプット
type PontoonWebOutput struct {
	Seats []*PontoonWebOutputSeat `json:"seats"`
	// BankerHand は親の手。配る前は null。
	BankerHand *PontoonWebOutputHand `json:"bankerHand,omitempty"`
	BankerIdx  int                   `json:"bankerIdx"`
	// IsHumanBanker が真の局は、人間がベットせず最後に引き止めを決める。
	IsHumanBanker bool `json:"isHumanBanker"`
	Chips         int  `json:"chips"`
	ActiveSeat    int  `json:"activeSeat"`
	ActiveHand    int  `json:"activeHand"`
	// NextBanker は次局の親（未定なら -1）。
	NextBanker int    `json:"nextBanker"`
	LastResult string `json:"lastResult"`
	Phase      int    `json:"phase"`
	// CanStick などは、規則が複雑なので都度サーバーが判断して渡す。
	// クライアントに「15 未満は宣言できない」「Twist 後は Buy 不可」を
	// 再実装させないため。
	CanStick bool `json:"canStick"`
	CanTwist bool `json:"canTwist"`
	CanBuy   bool `json:"canBuy"`
	CanSplit bool `json:"canSplit"`
	WebOutputBase
}

// PontoonWebController ポンツーン Web コントローラークラス
type PontoonWebController = GameWebController[usecase.PontoonInteractorIF, PontoonWebInput, *PontoonWebOutput]

// NewPontoonWebController and NewPontoonWebControllerWithProvider are the
// standard and provider-backed constructors for PontoonWebController.
var NewPontoonWebController, NewPontoonWebControllerWithProvider = webControllerPair[usecase.PontoonInteractorIF, PontoonWebInput, *PontoonWebOutput](
	newPontoonDefaultOutput, pontoonDispatch,
)

func newPontoonDefaultOutput(msg string) *PontoonWebOutput {
	return &PontoonWebOutput{
		Seats:         make([]*PontoonWebOutputSeat, 0),
		NextBanker:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pontoonDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PontoonInteractorIF, param PontoonWebInput, newDefault func(string) *PontoonWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newDefault, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Bet(*param.Amount))
	case "deal":
		bc.writePresenterResponse(w, pi.Deal())
	case "s", "stick":
		bc.writePresenterResponse(w, pi.Stick())
	case "t", "twist":
		bc.writePresenterResponse(w, pi.Twist())
	case "buy":
		if !requireParam(bc, w, newDefault, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Buy(*param.Amount))
	case "sp", "split":
		bc.writePresenterResponse(w, pi.Split())
	case "bt", "bankertwist":
		bc.writePresenterResponse(w, pi.BankerTwist())
	case "bs", "bankerstay":
		bc.writePresenterResponse(w, pi.BankerStay())
	case "log", "l":
		bc.writePresenterResponse(w, pi.ActionLog())
	case "r", "reset":
		bc.writePresenterResponse(w, pi.Reset())
	default:
		return false
	}
	return true
}
