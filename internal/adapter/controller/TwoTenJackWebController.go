package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TwoTenJackWebInput ツーテンジャックWebインプット
type TwoTenJackWebInput struct {
	BaseWebInput
	TrumpSuit *int                 `json:"trumpSuit,omitempty"`
	CardIndex *int                 `json:"cardIndex,omitempty"`
	Config    *TwoTenJackWebConfig `json:"config,omitempty"`
}

// TwoTenJackWebConfig ツーテンジャックWeb設定
type TwoTenJackWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// TwoTenJackWebOutputPlayer ツーテンジャックWebアウトプットプレイヤー
type TwoTenJackWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	CapturedPoints  int              `json:"capturedPoints"`
}

// TwoTenJackWebOutputHint ヒント出力
type TwoTenJackWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	TrumpSuit *int   `json:"trumpSuit,omitempty"`
	Reason    string `json:"reason"`
}

// TwoTenJackWebOutput ツーテンジャックWebアウトプット
type TwoTenJackWebOutput struct {
	Players          []*TwoTenJackWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	DeclarerIdx      int                          `json:"declarerIdx"`
	TrumpSuit        int                          `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	WinnerTeam       int                          `json:"winnerTeam"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	Hint             *TwoTenJackWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config TwoTenJackWebOutputConfig `json:"config"`
}

// TwoTenJackWebOutputConfig ツーテンジャック設定アウトプット
type TwoTenJackWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a TwoTenJackConfig from the nested web config.
func (c *TwoTenJackWebConfig) ToConfig() domain.TwoTenJackConfig {
	cfg := domain.DefaultTwoTenJackConfig()
	cfg.CpuDifficulty = domain.TwoTenJackCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TwoTenJackCpuDifficultyEasy), int(domain.TwoTenJackCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a TwoTenJackConfig from the web input.
func (p TwoTenJackWebInput) ToConfig() domain.TwoTenJackConfig {
	return configOrDefault(p.Config, (*TwoTenJackWebConfig).ToConfig, domain.DefaultTwoTenJackConfig())
}

// TwoTenJackWebController ツーテンジャックWebコントローラークラス
type TwoTenJackWebController = GameWebController[usecase.TwoTenJackInteractorIF, TwoTenJackWebInput, *TwoTenJackWebOutput]

// NewTwoTenJackWebController and NewTwoTenJackWebControllerWithProvider.
var NewTwoTenJackWebController, NewTwoTenJackWebControllerWithProvider = webControllerPair[usecase.TwoTenJackInteractorIF, TwoTenJackWebInput, *TwoTenJackWebOutput](
	newTwoTenJackDefaultOutput, twoTenJackDispatch,
)

func newTwoTenJackDefaultOutput(msg string) *TwoTenJackWebOutput {
	return &TwoTenJackWebOutput{
		Players:       make([]*TwoTenJackWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		TrumpSuit:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func twoTenJackDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TwoTenJackInteractorIF, param TwoTenJackWebInput, newDefault func(string) *TwoTenJackWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "d", "declare":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.DeclareTrump(*param.TrumpSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
