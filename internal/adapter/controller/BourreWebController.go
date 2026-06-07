//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BourreWebConfig ブーレ設定 (入力・出力共用)
type BourreWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts BourreWebConfig to domain.BourreConfig.
func (c BourreWebConfig) ToConfig() domain.BourreConfig {
	return domain.BourreConfig{
		CpuDifficulty: domain.BourreCpuDifficulty(c.CpuDifficulty),
	}
}

// BourreWebInput ブーレWebインプット
type BourreWebInput struct {
	BaseWebInput
	Decide    *bool            `json:"decide"`    // decide コマンド用 (true=参加, false=フォールド)
	Indices   []int            `json:"indices"`   // draw コマンド用 (捨てるカードのインデックス)
	CardIndex *int             `json:"cardIndex"` // play コマンド用
	Config    *BourreWebConfig `json:"config"`    // リセット時の設定 (省略可)
}

// BourreWebOutputPlayer ブーレWebアウトプットプレイヤー
type BourreWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Folded     bool             `json:"folded"`
	Decided    bool             `json:"decided"`
	Drawn      bool             `json:"drawn"`
	Bourreed   bool             `json:"bourreed"`
	Chips      int              `json:"chips"`
	Tricks     int              `json:"tricks"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// BourreWebTrickCard ブーレのトリック中の1枚
type BourreWebTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// BourreWebResult ブーレのハンド結果
type BourreWebResult struct {
	PlayerIdx int  `json:"playerIdx"`
	Tricks    int  `json:"tricks"`
	WonAmount int  `json:"wonAmount"`
	Bourreed  bool `json:"bourreed"`
	Folded    bool `json:"folded"`
}

// BourreWebOutput ブーレWebアウトプット
type BourreWebOutput struct {
	Players          []*BourreWebOutputPlayer `json:"players"`
	Phase            string                   `json:"phase"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	Pot              int                      `json:"pot"`
	CarryPot         int                      `json:"carryPot"`
	TrumpSuit        string                   `json:"trumpSuit"`
	TrumpCard        *WebOutputCard           `json:"trumpCard"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentTrick     []*BourreWebTrickCard    `json:"currentTrick"`
	LastTrick        []*BourreWebTrickCard    `json:"lastTrick"`
	LastTrickWinner  int                      `json:"lastTrickWinner"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	HandNumber       int                      `json:"handNumber"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	ValidPlays       []int                    `json:"validPlays"`
	Results          []*BourreWebResult       `json:"results"`
	Config           BourreWebConfig          `json:"config"`
	WebOutputBase
}

// BourreWebController ブーレWebコントローラー
type BourreWebController = GameWebController[usecase.BourreInteractorIF, BourreWebInput, *BourreWebOutput]

// NewBourreWebController and NewBourreWebControllerWithProvider are the standard
// and provider-backed constructors for BourreWebController.
var NewBourreWebController, NewBourreWebControllerWithProvider = webControllerPair[usecase.BourreInteractorIF, BourreWebInput, *BourreWebOutput](
	newBourreDefaultOutput, bourreDispatch,
)

func newBourreDefaultOutput(msg string) *BourreWebOutput {
	return &BourreWebOutput{
		Players:         make([]*BourreWebOutputPlayer, 0),
		CurrentTrick:    make([]*BourreWebTrickCard, 0),
		LastTrick:       make([]*BourreWebTrickCard, 0),
		ValidPlays:      make([]int, 0),
		Results:         make([]*BourreWebResult, 0),
		LastTrickWinner: -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func bourreDispatch(bc *baseController, w http.ResponseWriter, bgi usecase.BourreInteractorIF, param BourreWebInput, _ func(string) *BourreWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, bgi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, bgi.Reset())
		}
	case "decide":
		bc.writePresenterResponse(w, bgi.Decide(derefDefault(param.Decide, true)))
	case "draw":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, bgi.Draw(indices))
	case "p", "play":
		bc.writePresenterResponse(w, bgi.Play(deref(param.CardIndex)))
	case "next":
		bc.writePresenterResponse(w, bgi.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, bgi.ActionLog)
	}
	return true
}
