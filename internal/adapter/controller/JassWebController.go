//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// JassWebInput ヤスWebインプット
type JassWebInput struct {
	BaseWebInput
	Suit      *int           `json:"suit,omitempty"`
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *JassWebConfig `json:"config,omitempty"`
}

// JassWebConfig ヤスWeb設定
type JassWebConfig struct {
	CpuDifficulty  *int  `json:"cpuDifficulty,omitempty"`
	TargetScore    *int  `json:"targetScore,omitempty"`
	LastTrickBonus *int  `json:"lastTrickBonus,omitempty"`
	EnableWeis     *bool `json:"enableWeis,omitempty"`
}

// JassWebOutputPlayer ヤスWebアウトプットプレイヤー
type JassWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// JassWebOutputHint ヒント出力
type JassWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Schieben  *bool  `json:"schieben,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// JassWebOutput ヤスWebアウトプット
type JassWebOutput struct {
	Players          []*JassWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	TrickNumber      int                    `json:"trickNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	BidPlayerIdx     int                    `json:"bidPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	ForehandIdx      int                    `json:"forehandIdx"`
	TrumpSuit        int                    `json:"trumpSuit"`
	Schieben         bool                   `json:"schieben"`
	MakerTeam        int                    `json:"makerTeam"`
	MakerPlayerIdx   int                    `json:"makerPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard  `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard  `json:"lastTrick"`
	LastTrickWinner  int                    `json:"lastTrickWinner"`
	TeamScores       [2]int                 `json:"teamScores"`
	RoundPoints      [2]int                 `json:"roundPoints"`
	RoundWeisPoints  [2]int                 `json:"roundWeisPoints"`
	RoundStockPoints [2]int                 `json:"roundStockPoints"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerTeam       int                    `json:"winnerTeam"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	Hint             *JassWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config JassWebOutputConfig `json:"config"`
}

// JassWebOutputConfig ヤス設定アウトプット
type JassWebOutputConfig struct {
	CpuDifficulty  int  `json:"cpuDifficulty"`
	TargetScore    int  `json:"targetScore"`
	LastTrickBonus int  `json:"lastTrickBonus"`
	EnableWeis     bool `json:"enableWeis"`
}

// ToConfig builds a JassConfig from the nested web config, applying bounds checking.
func (c *JassWebConfig) ToConfig() domain.JassConfig {
	cfg := domain.DefaultJassConfig()
	cfg.CpuDifficulty = domain.JassCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.JassCpuDifficultyEasy), int(domain.JassCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	webutil.ApplyBoundedInt(&cfg.LastTrickBonus, c.LastTrickBonus, 0, 100)
	if c.EnableWeis != nil {
		cfg.EnableWeis = *c.EnableWeis
	}
	return cfg
}

// ToConfig builds a JassConfig from the web input.
func (p JassWebInput) ToConfig() domain.JassConfig {
	return configOrDefault(p.Config, (*JassWebConfig).ToConfig, domain.DefaultJassConfig())
}

// JassWebController ヤスWebコントローラークラス
type JassWebController = GameWebController[usecase.JassInteractorIF, JassWebInput, *JassWebOutput]

// NewJassWebController and NewJassWebControllerWithProvider are
// the standard and provider-backed constructors for JassWebController.
var NewJassWebController, NewJassWebControllerWithProvider = webControllerPair[usecase.JassInteractorIF, JassWebInput, *JassWebOutput](
	newJassDefaultOutput, jassDispatch,
)

func newJassDefaultOutput(msg string) *JassWebOutput {
	return &JassWebOutput{
		Players:         make([]*JassWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func jassDispatch(bc *baseController, w http.ResponseWriter, ji usecase.JassInteractorIF, param JassWebInput, newDefault func(string) *JassWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ji.ResetWithConfig(param.ToConfig()))
	case "c", "calltrump":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ji.ChooseTrump(*param.Suit))
	case "sc", "schieben":
		bc.writePresenterResponse(w, ji.Schieben())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ji.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ji.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ji.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ji.Hint, ji.ActionLog)
	}
	return true
}
