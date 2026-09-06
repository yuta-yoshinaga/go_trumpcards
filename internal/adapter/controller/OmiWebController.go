//go:build !js || !wasm || extra5

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmiWebInput オミWebインプット
type OmiWebInput struct {
	BaseWebInput
	Suit      *int          `json:"suit,omitempty"`
	CardIndex *int          `json:"cardIndex,omitempty"`
	Config    *OmiWebConfig `json:"config,omitempty"`
}

// OmiWebConfig オミWeb設定
type OmiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// OmiWebOutputPlayer オミWebアウトプットプレイヤー
type OmiWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// OmiWebOutputHint ヒント出力
type OmiWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// OmiWebOutput オミWebアウトプット
type OmiWebOutput struct {
	Players             []*OmiWebOutputPlayer `json:"players"`
	Phase               int                   `json:"phase"`
	RoundNumber         int                   `json:"roundNumber"`
	TrickNumber         int                   `json:"trickNumber"`
	CurrentPlayerIdx    int                   `json:"currentPlayerIdx"`
	TrumpCallerIdx      int                   `json:"trumpCallerIdx"`
	BidPlayerIdx        int                   `json:"bidPlayerIdx"`
	DealerIdx           int                   `json:"dealerIdx"`
	TrumpSuit           int                   `json:"trumpSuit"`
	DealStage           int                   `json:"dealStage"`
	FaceUpCard          *WebOutputCard        `json:"faceUpCard"`
	MakerTeam           int                   `json:"makerTeam"`
	GoingAlone          bool                  `json:"goingAlone"`
	GoingAlonePlayerIdx int                   `json:"goingAlonePlayerIdx"`
	CurrentTrick        []*WebOutputTrickCard `json:"currentTrick"`
	TeamScores          [2]int                `json:"teamScores"`
	TeamTricks          [2]int                `json:"teamTricks"`
	GameEndFlag         bool                  `json:"gameEndFlag"`
	WinnerTeam          int                   `json:"winnerTeam"`
	LeadPlayerIdx       int                   `json:"leadPlayerIdx"`
	Hint                *OmiWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config OmiWebOutputConfig `json:"config"`
}

// OmiWebOutputConfig オミ設定アウトプット
type OmiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds an OmiConfig from the nested web config, applying bounds checking.
func (c *OmiWebConfig) ToConfig() domain.OmiConfig {
	cfg := domain.DefaultOmiConfig()
	cfg.CpuDifficulty = domain.OmiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.OmiCpuDifficultyEasy), int(domain.OmiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds an OmiConfig from the web input.
func (p OmiWebInput) ToConfig() domain.OmiConfig {
	return configOrDefault(p.Config, (*OmiWebConfig).ToConfig, domain.DefaultOmiConfig())
}

// OmiWebController オミWebコントローラークラス
type OmiWebController = GameWebController[usecase.OmiInteractorIF, OmiWebInput, *OmiWebOutput]

// NewOmiWebController and NewOmiWebControllerWithProvider are
// the standard and provider-backed constructors for OmiWebController.
var NewOmiWebController, NewOmiWebControllerWithProvider = webControllerPair[usecase.OmiInteractorIF, OmiWebInput, *OmiWebOutput](
	newOmiDefaultOutput, omiDispatch,
)

func newOmiDefaultOutput(msg string) *OmiWebOutput {
	return &OmiWebOutput{
		Players:       make([]*OmiWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func omiDispatch(bc *baseController, w http.ResponseWriter, ei usecase.OmiInteractorIF, param OmiWebInput, newDefault func(string) *OmiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ei.ResetWithConfig(param.ToConfig()))
	case "t", "trump", "c", "calltrump":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.CallTrump(*param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ei.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ei.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ei.Hint, ei.ActionLog)
	}
	return true
}
