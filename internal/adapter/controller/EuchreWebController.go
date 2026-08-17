//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EuchreWebInput ユーカーWebインプット
type EuchreWebInput struct {
	BaseWebInput
	OrderUp   *bool            `json:"orderUp,omitempty"`
	GoAlone   *bool            `json:"goAlone,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *EuchreWebConfig `json:"config,omitempty"`
}

// EuchreWebConfig ユーカーWeb設定
type EuchreWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// EuchreWebOutputPlayer ユーカーWebアウトプットプレイヤー
type EuchreWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// EuchreWebOutputHint ヒント出力
type EuchreWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	OrderUp   *bool  `json:"orderUp,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	GoAlone   *bool  `json:"goAlone,omitempty"`
	Reason    string `json:"reason"`
	// Score は判断の元になったハンド強度 (ビッド局面のみ)。CUI と同じ値。
	Score *int `json:"score,omitempty"`
}

// EuchreWebOutput ユーカーWebアウトプット
type EuchreWebOutput struct {
	Players             []*EuchreWebOutputPlayer `json:"players"`
	Phase               int                      `json:"phase"`
	RoundNumber         int                      `json:"roundNumber"`
	TrickNumber         int                      `json:"trickNumber"`
	CurrentPlayerIdx    int                      `json:"currentPlayerIdx"`
	BidPlayerIdx        int                      `json:"bidPlayerIdx"`
	DealerIdx           int                      `json:"dealerIdx"`
	TrumpSuit           int                      `json:"trumpSuit"`
	FaceUpCard          *WebOutputCard           `json:"faceUpCard"`
	MakerTeam           int                      `json:"makerTeam"`
	GoingAlone          bool                     `json:"goingAlone"`
	GoingAlonePlayerIdx int                      `json:"goingAlonePlayerIdx"`
	CurrentTrick        []*WebOutputTrickCard    `json:"currentTrick"`
	TeamScores          [2]int                   `json:"teamScores"`
	GameEndFlag         bool                     `json:"gameEndFlag"`
	WinnerTeam          int                      `json:"winnerTeam"`
	LeadPlayerIdx       int                      `json:"leadPlayerIdx"`
	Hint                *EuchreWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config EuchreWebOutputConfig `json:"config"`
}

// EuchreWebOutputConfig ユーカー設定アウトプット
type EuchreWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds an EuchreConfig from the nested web config, applying bounds checking.
func (c *EuchreWebConfig) ToConfig() domain.EuchreConfig {
	cfg := domain.DefaultEuchreConfig()
	cfg.CpuDifficulty = domain.EuchreCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.EuchreCpuDifficultyEasy), int(domain.EuchreCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds an EuchreConfig from the web input.
func (p EuchreWebInput) ToConfig() domain.EuchreConfig {
	return configOrDefault(p.Config, (*EuchreWebConfig).ToConfig, domain.DefaultEuchreConfig())
}

// EuchreWebController ユーカーWebコントローラークラス
type EuchreWebController = GameWebController[usecase.EuchreInteractorIF, EuchreWebInput, *EuchreWebOutput]

// NewEuchreWebController and NewEuchreWebControllerWithProvider are
// the standard and provider-backed constructors for EuchreWebController.
var NewEuchreWebController, NewEuchreWebControllerWithProvider = webControllerPair[usecase.EuchreInteractorIF, EuchreWebInput, *EuchreWebOutput](
	newEuchreDefaultOutput, euchreDispatch,
)

func newEuchreDefaultOutput(msg string) *EuchreWebOutput {
	return &EuchreWebOutput{
		Players:       make([]*EuchreWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func euchreDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EuchreInteractorIF, param EuchreWebInput, newDefault func(string) *EuchreWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ei.ResetWithConfig(param.ToConfig()))
	case "o", "orderup":
		goAlone := param.GoAlone != nil && *param.GoAlone
		bc.writePresenterResponse(w, ei.PickUp(true, goAlone))
	case "c", "calltrump":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		goAlone := param.GoAlone != nil && *param.GoAlone
		bc.writePresenterResponse(w, ei.CallTrump(*param.Suit, goAlone))
	case "pa", "pass":
		bc.writePresenterResponse(w, ei.Pass())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.Discard(*param.CardIndex))
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
