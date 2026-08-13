//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BanLuckWebInput バンラックWebインプット
type BanLuckWebInput struct {
	BaseWebInput
	// Bet は賭け金。
	//
	// **親のラウンドでは 0 が正当。** 親は賭ける側ではないので、省略と
	// 「0 を送った」を区別する必要がある。
	Bet *int `json:"bet,omitempty"`
}

// BanLuckWebOutCfg はバンラックの設定
type BanLuckWebOutCfg struct {
	Seats        int `json:"seats"`
	InitialChips int `json:"initialChips"`
	Rounds       int `json:"rounds"`
	DefaultBet   int `json:"defaultBet"`
}

// BanLuckWebOutputSeat は 1 席の状態
type BanLuckWebOutputSeat struct {
	Name    string           `json:"name"`
	IsHuman bool             `json:"isHuman"`
	Chips   int              `json:"chips"`
	Bet     int              `json:"bet"`
	Cards   []*WebOutputCard `json:"cards"`
	Score   int              `json:"score"`
	// Rank は役 (0=bust, 1=point, 2=fiveDragon, 3=banLuck, 4=banBan)。
	Rank int `json:"rank"`
	// Outcome は親から見た決着 (0=lose, 1=push, 2=win)。親自身は push。
	Outcome int `json:"outcome"`
	// RoundBet はそのラウンドに置いた額。**精算後は席の bet が 0 に戻る**ので、
	// 表示にはこちらを使う。
	RoundBet int  `json:"roundBet"`
	Delta    int  `json:"delta"`
	Busted   bool `json:"busted"`
	Stood    bool `json:"stood"`
	IsBanker bool `json:"isBanker"`
	IsTurn   bool `json:"isTurn"`
}

// BanLuckWebOutput バンラックWebアウトプット
type BanLuckWebOutput struct {
	Phase int                     `json:"phase"`
	Seats []*BanLuckWebOutputSeat `json:"seats"`
	// BankerSeat / TurnSeat / HumanSeat は席の添字。
	BankerSeat int `json:"bankerSeat"`
	TurnSeat   int `json:"turnSeat"`
	HumanSeat  int `json:"humanSeat"`
	// IsHumanTurn は人間の操作待ちか。
	IsHumanTurn bool `json:"isHumanTurn"`
	// MustHit は人間が親で、いま引く義務を負っているか。
	//
	// **ページに 15 未満かを計算し直させない。** 規則が 2 か所に増えると必ずずれる。
	MustHit        bool `json:"mustHit"`
	RoundNumber    int  `json:"roundNumber"`
	RemainingCards int  `json:"remainingCards"`
	WinnerSeat     int  `json:"winnerSeat"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *BanLuckWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// BanLuckWebController バンラックWebコントローラークラス
type BanLuckWebController = GameWebController[usecase.BanLuckInteractorIF, BanLuckWebInput, *BanLuckWebOutput]

// NewBanLuckWebController and NewBanLuckWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewBanLuckWebController, NewBanLuckWebControllerWithProvider = webControllerPair[usecase.BanLuckInteractorIF, BanLuckWebInput, *BanLuckWebOutput](
	newBanLuckDefaultOutput, banLuckDispatch,
)

func newBanLuckDefaultOutput(msg string) *BanLuckWebOutput {
	return &BanLuckWebOutput{
		Seats:         make([]*BanLuckWebOutputSeat, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func banLuckDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BanLuckInteractorIF, param BanLuckWebInput, newOut func(string) *BanLuckWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		// **0 は「親なので賭けない」で、正当な値。** 省略とは違う。
		if !requireParam(bc, w, newOut, param.Bet == nil, "param error: bet is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlaceBet(*param.Bet))
	case "h", "hit":
		bc.writePresenterResponse(w, ci.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, ci.Stand())
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
