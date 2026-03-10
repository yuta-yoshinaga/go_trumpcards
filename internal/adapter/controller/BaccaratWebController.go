package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// BaccaratWebInput バカラWebインプット
type BaccaratWebInput struct {
	BaseWebInput
	Amount  int  `json:"amount,omitempty"`
	BetType *int `json:"betType,omitempty"`
}

// BaccaratWebOutput バカラWebアウトプット
type BaccaratWebOutput struct {
	PlayerHand      []*WebOutputCard  `json:"playerHand"`
	BankerHand      []*WebOutputCard  `json:"bankerHand"`
	PlayerHandValue int               `json:"playerHandValue"`
	BankerHandValue int               `json:"bankerHandValue"`
	Phase           int               `json:"phase"`
	Chips           int               `json:"chips"`
	BetAmount       int               `json:"betAmount"`
	BetType         int               `json:"betType"`
	Result          int               `json:"result"`
	Payout          int               `json:"payout"`
	Message         string            `json:"message"`
	MessageCode     string            `json:"messageCode,omitempty"`
	MessageParams   map[string]string `json:"messageParams,omitempty"`
}

// BaccaratWebController バカラWebコントローラークラス
type BaccaratWebController = GameWebController[usecase.BaccaratInteractorIF, BaccaratWebInput, *BaccaratWebOutput]

// NewBaccaratWebController コンストラクタ
func NewBaccaratWebController(factory func() usecase.BaccaratInteractorIF) *BaccaratWebController {
	return NewGameWebController(factory, newBaccaratDefaultOutput, baccaratDispatch)
}

func newBaccaratDefaultOutput(msg string) *BaccaratWebOutput {
	return &BaccaratWebOutput{
		PlayerHand: make([]*WebOutputCard, 0),
		BankerHand: make([]*WebOutputCard, 0),
		Message:    msg,
	}
}

func baccaratDispatch(bc *baseController, w rest.ResponseWriter, bi usecase.BaccaratInteractorIF, param BaccaratWebInput, _ func(string) *BaccaratWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.Reset())
	case "b", "bet":
		bt := derefInt(param.BetType)
		bc.writePresenterResponse(w, bi.Bet(param.Amount, bt))
	case "log", "l":
		bc.writePresenterResponse(w, bi.ActionLog())
	default:
		return false
	}
	return true
}
