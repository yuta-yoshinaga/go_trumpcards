//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SetteEMezzoWebInput セッテ・エ・メッツォ Web インプット
type SetteEMezzoWebInput struct {
	BaseWebInput
	// Amount is the stake for "bet" and the matta's value in HALVES for "matta".
	Amount *int `json:"amount,omitempty"`
}

// SetteEMezzoWebOutputHand 1 つの手の出力
type SetteEMezzoWebOutputHand struct {
	Cards []*WebOutputCard `json:"cards"`
	Bet   int              `json:"bet"`
	// TotalHalves は合計を**半点単位**で表したもの。0.5 点札があるので、
	// 小数をワイヤに載せずに正確な等値比較ができる形で渡す。
	TotalHalves int `json:"totalHalves"`
	// TotalLabel は "7.5" のような表示用の文字列。
	TotalLabel string `json:"totalLabel"`
	// MattaHalves はマッタに割り当てた値（半点単位・マッタが無ければ 0）。
	MattaHalves int  `json:"mattaHalves"`
	HasMatta    bool `json:"hasMatta"`
	Stood       bool `json:"stood"`
	Payout      int  `json:"payout"`
	// Hidden が真のとき、Cards の要素は null で合計も伏せられる。枚数だけが残る。
	Hidden bool `json:"hidden"`
}

// SetteEMezzoWebOutputSeat 1 席の出力
type SetteEMezzoWebOutputSeat struct {
	Name  string                    `json:"name"`
	IsCPU bool                      `json:"isCpu"`
	Hand  *SetteEMezzoWebOutputHand `json:"hand,omitempty"`
}

// SetteEMezzoWebOutput セッテ・エ・メッツォ Web アウトプット
type SetteEMezzoWebOutput struct {
	Seats         []*SetteEMezzoWebOutputSeat `json:"seats"`
	BankerHand    *SetteEMezzoWebOutputHand   `json:"bankerHand,omitempty"`
	BankerIdx     int                         `json:"bankerIdx"`
	IsHumanBanker bool                        `json:"isHumanBanker"`
	Chips         int                         `json:"chips"`
	ActiveSeat    int                         `json:"activeSeat"`
	NextBanker    int                         `json:"nextBanker"`
	LastResult    string                      `json:"lastResult"`
	Phase         int                         `json:"phase"`
	// TargetHalves は 7.5 を半点単位で表したもの（15）。
	TargetHalves int  `json:"targetHalves"`
	CanHit       bool `json:"canHit"`
	CanStand     bool `json:"canStand"`
	CanSetMatta  bool `json:"canSetMatta"`
	WebOutputBase
}

// SetteEMezzoWebController セッテ・エ・メッツォ Web コントローラークラス
type SetteEMezzoWebController = GameWebController[usecase.SetteEMezzoInteractorIF, SetteEMezzoWebInput, *SetteEMezzoWebOutput]

// NewSetteEMezzoWebController and NewSetteEMezzoWebControllerWithProvider are the
// standard and provider-backed constructors for SetteEMezzoWebController.
var NewSetteEMezzoWebController, NewSetteEMezzoWebControllerWithProvider = webControllerPair[usecase.SetteEMezzoInteractorIF, SetteEMezzoWebInput, *SetteEMezzoWebOutput](
	newSetteEMezzoDefaultOutput, setteEMezzoDispatch,
)

func newSetteEMezzoDefaultOutput(msg string) *SetteEMezzoWebOutput {
	return &SetteEMezzoWebOutput{
		Seats:         make([]*SetteEMezzoWebOutputSeat, 0),
		NextBanker:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func setteEMezzoDispatch(bc *baseController, w http.ResponseWriter, si usecase.SetteEMezzoInteractorIF, param SetteEMezzoWebInput, newDefault func(string) *SetteEMezzoWebOutput) bool {
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
	case "matta":
		if !requireParam(bc, w, newDefault, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Matta(*param.Amount))
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
