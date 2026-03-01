package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// BlackJackWebInput ブラックジャックWebインプット
type BlackJackWebInput struct {
	Command   string `json:"command"`
	Amount    int    `json:"amount,omitempty"`
	SessionId string `json:"sessionId"`
}

// GetCommand returns the command string.
func (i BlackJackWebInput) GetCommand() string { return i.Command }

// GetSessionID returns the session ID string.
func (i BlackJackWebInput) GetSessionID() string { return i.SessionId }

// BlackJackWebOutputHand ブラックジャックWebアウトプットハンド
type BlackJackWebOutputHand struct {
	Score        int              `json:"score"`
	Cards        []*WebOutputCard `json:"cards"`
	Bet          int              `json:"bet"`
	Stood        bool             `json:"stood"`
	Doubled      bool             `json:"doubled"`
	Busted       bool             `json:"busted"`
	IsBlackJack  bool             `json:"isBlackJack"`
	CanSplit     bool             `json:"canSplit"`
	Surrendered  bool             `json:"surrendered"`
	CanSurrender bool             `json:"canSurrender"`
}

// BlackJackWebOutputPlayer ブラックジャックWebアウトプットプレイヤー
type BlackJackWebOutputPlayer struct {
	Score int              `json:"score"`
	Cards []*WebOutputCard `json:"cards"`
	Chips int              `json:"chips"`
}

// BlackJackWebOutput ブラックジャックWebアウトプット
type BlackJackWebOutput struct {
	Dealer             *BlackJackWebOutputPlayer `json:"dealer"`
	Player             *BlackJackWebOutputPlayer `json:"player"`
	Hands              []*BlackJackWebOutputHand `json:"hands,omitempty"`
	CurrentHandIdx     int                       `json:"currentHandIdx"`
	Phase              int                       `json:"phase"`
	InsuranceBet       int                       `json:"insuranceBet"`
	InsuranceAvailable bool                      `json:"insuranceAvailable"`
	Message            string                    `json:"message"`
	HintEnabled        bool                      `json:"hintEnabled"`
	SuggestedAction    int                       `json:"suggestedAction"`
	DeckCount          int                       `json:"deckCount"`
}

// BlackJackWebController ブラックジャックWebコントローラークラス
type BlackJackWebController struct {
	baseController
	factory func() usecase.BlackJackInteractorIF
	store   *SessionStore[usecase.BlackJackInteractorIF]
}

// NewBlackJackWebController コンストラクタ
func NewBlackJackWebController(factory func() usecase.BlackJackInteractorIF) *BlackJackWebController {
	return &BlackJackWebController{
		factory: factory,
		store:   NewSessionStore[usecase.BlackJackInteractorIF](),
	}
}

// Exec ゲーム実行
func (bwc *BlackJackWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&bwc.baseController, w, r, bwc.store, bwc.factory,
		func(msg string) any { return bwc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, bji usecase.BlackJackInteractorIF, param BlackJackWebInput, errOutput any) bool {
			switch param.Command {
			case "r", "reset":
				bwc.writePresenterResponse(w, bji.Reset(), errOutput)
			case "h", "hit":
				bwc.writePresenterResponse(w, bji.Hit(), errOutput)
			case "s", "stand":
				bwc.writePresenterResponse(w, bji.Stand(), errOutput)
			case "b", "bet":
				bwc.writePresenterResponse(w, bji.Bet(param.Amount), errOutput)
			case "d", "doubledown":
				bwc.writePresenterResponse(w, bji.DoubleDown(), errOutput)
			case "sp", "split":
				bwc.writePresenterResponse(w, bji.Split(), errOutput)
			case "i", "insurance":
				bwc.writePresenterResponse(w, bji.Insurance(), errOutput)
			case "di", "declineinsurance":
				bwc.writePresenterResponse(w, bji.DeclineInsurance(), errOutput)
			case "sur", "surrender":
				bwc.writePresenterResponse(w, bji.Surrender(), errOutput)
			case "togglehint":
				bwc.writePresenterResponse(w, bji.ToggleHint(), errOutput)
			case "sd", "setdeckcount":
				bwc.writePresenterResponse(w, bji.SetDeckCount(param.Amount), errOutput)
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (bwc *BlackJackWebController) Stop() {
	bwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (bwc *BlackJackWebController) newDefaultOutput(msg string) *BlackJackWebOutput {
	return &BlackJackWebOutput{
		Dealer:    &BlackJackWebOutputPlayer{},
		Player:    &BlackJackWebOutputPlayer{},
		Message:   msg,
		DeckCount: 1,
	}
}
