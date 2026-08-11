//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HasenpfefferWebInput ハーゼンプフェファーWebインプット
type HasenpfefferWebInput struct {
	BaseWebInput
	CardIndex *int                   `json:"cardIndex,omitempty"`
	Suit      *int                   `json:"suit,omitempty"`
	Bid       *int                   `json:"bid,omitempty"`
	Config    *HasenpfefferWebConfig `json:"config,omitempty"`
}

// HasenpfefferWebConfig ハーゼンプフェファーWeb設定
type HasenpfefferWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// HasenpfefferWebOutputPlayer ハーゼンプフェファーWebアウトプットプレイヤー
type HasenpfefferWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Bid は宣言額 (-1: 未宣言、0: 降りた)。
	Bid        int `json:"bid"`
	TrickCount int `json:"trickCount"`
}

// HasenpfefferWebOutputHint ヒント出力
type HasenpfefferWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Value は宣言すべきトリック数（プレイ中は 0）。
	Value int `json:"value"`
	// Suit は切り札に勧めるスート（プレイ中は 0）。
	Suit int `json:"suit"`
}

// HasenpfefferWebOutput ハーゼンプフェファーWebアウトプット
type HasenpfefferWebOutput struct {
	Players     []*HasenpfefferWebOutputPlayer `json:"players"`
	Phase       int                            `json:"phase"`
	HandNumber  int                            `json:"handNumber"`
	TrickNumber int                            `json:"trickNumber"`
	// TrumpSuit は 0 のあいだ未宣言。落札者が捨て札と一緒に決める。
	TrumpSuit   int `json:"trumpSuit"`
	DeclarerIdx int `json:"declarerIdx"`
	Contract    int `json:"contract"`
	// MinBid は次に出せる最小額 (0: もう宣言できない = 降りるしかない)。
	MinBid int `json:"minBid"`
	// MustBid は人間が降りられない（義務競り）かどうか。
	MustBid bool `json:"mustBid"`
	// BlindSize は伏せ札の枚数（落札者が取り込むと 0）。
	BlindSize        int                        `json:"blindSize"`
	Scores           []int                      `json:"scores"`
	TeamTricks       []int                      `json:"teamTricks"`
	LastHandEuchred  bool                       `json:"lastHandEuchred"`
	LastHandTricks   int                        `json:"lastHandTricks"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	ValidPlays       []int                      `json:"validPlays"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerTeam       int                        `json:"winnerTeam"`
	Hint             *HasenpfefferWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config HasenpfefferWebOutputConfig `json:"config"`
}

// HasenpfefferWebOutputConfig ハーゼンプフェファー設定アウトプット
type HasenpfefferWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a HasenpfefferConfig from the nested web config, applying bounds checking.
func (c *HasenpfefferWebConfig) ToConfig() domain.HasenpfefferConfig {
	cfg := domain.DefaultHasenpfefferConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.HasenpfefferTargetMin, domain.HasenpfefferTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a HasenpfefferConfig from the web input.
func (p HasenpfefferWebInput) ToConfig() domain.HasenpfefferConfig {
	return configOrDefault(p.Config, (*HasenpfefferWebConfig).ToConfig, domain.DefaultHasenpfefferConfig())
}

// HasenpfefferWebController ハーゼンプフェファーWebコントローラークラス
type HasenpfefferWebController = GameWebController[usecase.HasenpfefferInteractorIF, HasenpfefferWebInput, *HasenpfefferWebOutput]

// NewHasenpfefferWebController and NewHasenpfefferWebControllerWithProvider are
// the standard and provider-backed constructors for HasenpfefferWebController.
var NewHasenpfefferWebController, NewHasenpfefferWebControllerWithProvider = webControllerPair[usecase.HasenpfefferInteractorIF, HasenpfefferWebInput, *HasenpfefferWebOutput](
	newHasenpfefferDefaultOutput, hasenpfefferDispatch,
)

func newHasenpfefferDefaultOutput(msg string) *HasenpfefferWebOutput {
	return &HasenpfefferWebOutput{
		Players:       make([]*HasenpfefferWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		Scores:        make([]int, 0),
		TeamTricks:    make([]int, 0),
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func hasenpfefferDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HasenpfefferInteractorIF, param HasenpfefferWebInput, newDefault func(string) *HasenpfefferWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, hi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		// **0（降りる）と未指定は違う。** 未指定を 0 で埋めると、降りるつもりの
		// 無い人が勝手に降ろされる。
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Bid(*param.Bid))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Discard(*param.CardIndex, *param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, hi.NextHand())
	case "g", "giveup":
		bc.writePresenterResponse(w, hi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, hi.Hint, hi.ActionLog)
	}
	return true
}
