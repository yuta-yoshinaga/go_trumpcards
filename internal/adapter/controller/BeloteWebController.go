//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeloteWebInput ベロートWebインプット
type BeloteWebInput struct {
	BaseWebInput
	OrderUp   *bool            `json:"orderUp,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *BeloteWebConfig `json:"config,omitempty"`
}

// BeloteWebConfig ベロートWeb設定
type BeloteWebConfig struct {
	CpuDifficulty        *int  `json:"cpuDifficulty,omitempty"`
	TargetScore          *int  `json:"targetScore,omitempty"`
	DixDeDer             *int  `json:"dixDeDer,omitempty"`
	EnableBeloteRebelote *bool `json:"enableBeloteRebelote,omitempty"`
}

// BeloteWebOutputPlayer ベロートWebアウトプットプレイヤー
type BeloteWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// BeloteWebOutputHint ヒント出力
type BeloteWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	OrderUp   *bool  `json:"orderUp,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// BeloteWebOutput ベロートWebアウトプット
type BeloteWebOutput struct {
	Players          []*BeloteWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	BidPlayerIdx     int                      `json:"bidPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	TrumpSuit        int                      `json:"trumpSuit"`
	FaceUpCard       *WebOutputCard           `json:"faceUpCard"`
	MakerTeam        int                      `json:"makerTeam"`
	MakerPlayerIdx   int                      `json:"makerPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	TeamScores       [2]int                   `json:"teamScores"`
	RoundPoints      [2]int                   `json:"roundPoints"`
	RoundBeloteBonus [2]int                   `json:"roundBeloteBonus"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *BeloteWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config BeloteWebOutputConfig `json:"config"`
}

// BeloteWebOutputConfig ベロート設定アウトプット
type BeloteWebOutputConfig struct {
	CpuDifficulty        int  `json:"cpuDifficulty"`
	TargetScore          int  `json:"targetScore"`
	DixDeDer             int  `json:"dixDeDer"`
	EnableBeloteRebelote bool `json:"enableBeloteRebelote"`
}

// ToConfig builds a BeloteConfig from the nested web config, applying bounds checking.
func (c *BeloteWebConfig) ToConfig() domain.BeloteConfig {
	cfg := domain.DefaultBeloteConfig()
	cfg.CpuDifficulty = domain.BeloteCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BeloteCpuDifficultyEasy), int(domain.BeloteCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	webutil.ApplyBoundedInt(&cfg.DixDeDer, c.DixDeDer, 0, 100)
	if c.EnableBeloteRebelote != nil {
		cfg.EnableBeloteRebelote = *c.EnableBeloteRebelote
	}
	return cfg
}

// ToConfig builds a BeloteConfig from the web input.
func (p BeloteWebInput) ToConfig() domain.BeloteConfig {
	return configOrDefault(p.Config, (*BeloteWebConfig).ToConfig, domain.DefaultBeloteConfig())
}

// BeloteWebController ベロートWebコントローラークラス
type BeloteWebController = GameWebController[usecase.BeloteInteractorIF, BeloteWebInput, *BeloteWebOutput]

// NewBeloteWebController and NewBeloteWebControllerWithProvider are
// the standard and provider-backed constructors for BeloteWebController.
var NewBeloteWebController, NewBeloteWebControllerWithProvider = webControllerPair[usecase.BeloteInteractorIF, BeloteWebInput, *BeloteWebOutput](
	newBeloteDefaultOutput, beloteDispatch,
)

func newBeloteDefaultOutput(msg string) *BeloteWebOutput {
	return &BeloteWebOutput{
		Players:       make([]*BeloteWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func beloteDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BeloteInteractorIF, param BeloteWebInput, newDefault func(string) *BeloteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "o", "orderup":
		bc.writePresenterResponse(w, bi.PickUp(true))
	case "c", "calltrump":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.CallTrump(*param.Suit))
	case "pa", "pass":
		bc.writePresenterResponse(w, bi.Pass())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
