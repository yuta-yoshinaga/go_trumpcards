//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GermanWhistWebInput ジャーマンホイストWebインプット
type GermanWhistWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
}

// GermanWhistWebOutputPlayer ジャーマンホイストWebアウトプットプレイヤー
type GermanWhistWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は伏せられている相手の手札枚数を出すために要る。Cards は
	// 人間の分しか入らない。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// TrickCount は前半も含む総獲得トリック数、ScoringTricks が**得点になる
	// 後半だけ**の数。表示上どちらも要るので両方返す。
	TrickCount    int `json:"trickCount"`
	ScoringTricks int `json:"scoringTricks"`
}

// GermanWhistWebOutputHint ヒント出力
type GermanWhistWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// GermanWhistWebOutput ジャーマンホイストWebアウトプット
type GermanWhistWebOutput struct {
	Players          []*GermanWhistWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	TrickNumber      int                           `json:"trickNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard         `json:"currentTrick"`
	TrumpSuit        int                           `json:"trumpSuit"`
	// UpCard は前半で奪い合う表向きの 1 枚。後半は無いので omitempty。
	UpCard      *WebOutputCard            `json:"upCard,omitempty"`
	StockCount  int                       `json:"stockCount"`
	ValidPlays  []int                     `json:"validPlays"`
	GameEndFlag bool                      `json:"gameEndFlag"`
	WinnerIdx   int                       `json:"winnerIdx"`
	Hint        *GermanWhistWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// GermanWhistWebController ジャーマンホイストWebコントローラークラス
type GermanWhistWebController = GameWebController[usecase.GermanWhistInteractorIF, GermanWhistWebInput, *GermanWhistWebOutput]

// NewGermanWhistWebController and NewGermanWhistWebControllerWithProvider are
// the standard and provider-backed constructors for GermanWhistWebController.
var NewGermanWhistWebController, NewGermanWhistWebControllerWithProvider = webControllerPair[usecase.GermanWhistInteractorIF, GermanWhistWebInput, *GermanWhistWebOutput](
	newGermanWhistDefaultOutput, germanWhistDispatch,
)

func newGermanWhistDefaultOutput(msg string) *GermanWhistWebOutput {
	return &GermanWhistWebOutput{
		Players:       make([]*GermanWhistWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func germanWhistDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GermanWhistInteractorIF, param GermanWhistWebInput, newDefault func(string) *GermanWhistWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.Reset())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Play(*param.CardIndex))
	case "g", "giveup":
		bc.writePresenterResponse(w, gi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, gi.Hint, gi.ActionLog)
	}
	return true
}
