//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShelemWebInput シェレムWebインプット
type ShelemWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	Bid       *int             `json:"bid,omitempty"`
	Discards  []int            `json:"discards,omitempty"`
	Config    *ShelemWebConfig `json:"config,omitempty"`
}

// ShelemWebConfig シェレムWeb設定
type ShelemWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// ShelemWebOutputPlayer シェレムWebアウトプットプレイヤー
type ShelemWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Bid は競りで出した点数 (-1: 未入札/パス)。
	Bid            int  `json:"bid"`
	Passed         bool `json:"passed"`
	DeclaredShelem bool `json:"declaredShelem"`
	TrickCount     int  `json:"trickCount"`
}

// ShelemWebOutputHint ヒント出力
type ShelemWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	Value     int    `json:"value"`
	Suit      int    `json:"suit"`
}

// ShelemWebOutput シェレムWebアウトプット
type ShelemWebOutput struct {
	Players     []*ShelemWebOutputPlayer `json:"players"`
	Phase       int                      `json:"phase"`
	RoundNumber int                      `json:"roundNumber"`
	TrickNumber int                      `json:"trickNumber"`
	TrumpSuit   int                      `json:"trumpSuit"`
	// DeclarerIdx / Contract / ShelemBid は競りの結果。
	DeclarerIdx int  `json:"declarerIdx"`
	Contract    int  `json:"contract"`
	ShelemBid   bool `json:"shelemBid"`
	// MinBid は次に出せる最小の入札額（競りが進むと上がる）。
	MinBid int `json:"minBid"`
	// WidowSize は伏せられている枚数（落札後は 0）、DiscardCount は捨てる枚数。
	WidowSize    int `json:"widowSize"`
	DiscardCount int `json:"discardCount"`
	// Scores は累計、RoundPoints は現ラウンドのカード点。どちらもチーム単位。
	Scores           []int                 `json:"scores"`
	RoundPoints      []int                 `json:"roundPoints"`
	TeamTricks       []int                 `json:"teamTricks"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	BidPlayerIdx     int                   `json:"bidPlayerIdx"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	ValidPlays       []int                 `json:"validPlays"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerTeam       int                   `json:"winnerTeam"`
	Hint             *ShelemWebOutputHint  `json:"hint,omitempty"`
	WebOutputBase
	Config ShelemWebOutputConfig `json:"config"`
}

// ShelemWebOutputConfig シェレム設定アウトプット
type ShelemWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a ShelemConfig from the nested web config, applying bounds checking.
func (c *ShelemWebConfig) ToConfig() domain.ShelemConfig {
	cfg := domain.DefaultShelemConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.ShelemTargetMin, domain.ShelemTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a ShelemConfig from the web input.
func (p ShelemWebInput) ToConfig() domain.ShelemConfig {
	return configOrDefault(p.Config, (*ShelemWebConfig).ToConfig, domain.DefaultShelemConfig())
}

// ShelemWebController シェレムWebコントローラークラス
type ShelemWebController = GameWebController[usecase.ShelemInteractorIF, ShelemWebInput, *ShelemWebOutput]

// NewShelemWebController and NewShelemWebControllerWithProvider are
// the standard and provider-backed constructors for ShelemWebController.
var NewShelemWebController, NewShelemWebControllerWithProvider = webControllerPair[usecase.ShelemInteractorIF, ShelemWebInput, *ShelemWebOutput](
	newShelemDefaultOutput, shelemDispatch,
)

func newShelemDefaultOutput(msg string) *ShelemWebOutput {
	return &ShelemWebOutput{
		Players:       make([]*ShelemWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		Scores:        make([]int, 0),
		RoundPoints:   make([]int, 0),
		TeamTricks:    make([]int, 0),
		DeclarerIdx:   -1,
		MinBid:        domain.ShelemMinBid,
		DiscardCount:  domain.ShelemWidowSize,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func shelemDispatch(bc *baseController, w http.ResponseWriter, si usecase.ShelemInteractorIF, param ShelemWebInput, newDefault func(string) *ShelemWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		// **入札額は既定値で埋めない。** 埋めると出していない額で落札する。
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.Bid))
	case "shelem":
		bc.writePresenterResponse(w, si.BidShelem())
	case "pass":
		bc.writePresenterResponse(w, si.Pass())
	case "d", "discard":
		// **捨て札とスートの両方が要る。** 片方でも欠けたら通さない。
		if !requireParam(bc, w, newDefault, len(param.Discards) == 0 || param.Suit == nil,
			"param error: discards and suit are required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Discard(param.Discards, *param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
