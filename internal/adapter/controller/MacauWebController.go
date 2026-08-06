//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MacauWebInput マカオWebインプット
type MacauWebInput struct {
	BaseWebInput
	CardIndex *int            `json:"cardIndex,omitempty"`
	Suit      *int            `json:"suit,omitempty"`
	Config    *MacauWebConfig `json:"config,omitempty"`
}

// MacauWebConfig マカオWeb設定
type MacauWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// MacauWebOutputPlayer マカオWebアウトプットプレイヤー
type MacauWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	HasDeclared     bool             `json:"hasDeclared"`
}

// MacauWebOutput マカオWebアウトプット
type MacauWebOutput struct {
	Players          []*MacauWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard          `json:"discardTop"`
	DrawPileCount    int                     `json:"drawPileCount"`
	ChosenSuit       int                     `json:"chosenSuit"`
	PenaltyDrawCount int                     `json:"penaltyDrawCount"`
	// PlayableIndices は人間がいま出せる手札の位置。マジックカードやチョウズド
	// スートの絡む合法判定を画面が示すために使う (#4805)。
	PlayableIndices []int `json:"playableIndices"`
	Direction       int   `json:"direction"`
	GameEndFlag     bool  `json:"gameEndFlag"`
	WinnerIdx       int   `json:"winnerIdx"`
	WebOutputBase
	Config MacauWebOutputConfig `json:"config"`
}

// MacauWebOutputConfig マカオ設定アウトプット
type MacauWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a MacauConfig from the nested web config, applying bounds checking.
func (c *MacauWebConfig) ToConfig() domain.MacauConfig {
	cfg := domain.DefaultMacauConfig()
	cfg.CpuDifficulty = domain.MacauCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MacauCpuDifficultyEasy), int(domain.MacauCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a MacauConfig from the web input.
func (p MacauWebInput) ToConfig() domain.MacauConfig {
	return configOrDefault(p.Config, (*MacauWebConfig).ToConfig, domain.DefaultMacauConfig())
}

// MacauWebController マカオWebコントローラークラス
type MacauWebController = GameWebController[usecase.MacauInteractorIF, MacauWebInput, *MacauWebOutput]

// NewMacauWebController and NewMacauWebControllerWithProvider are
// the standard and provider-backed constructors for MacauWebController.
var NewMacauWebController, NewMacauWebControllerWithProvider = webControllerPair[usecase.MacauInteractorIF, MacauWebInput, *MacauWebOutput](
	newMacauDefaultOutput, macauDispatch,
)

func newMacauDefaultOutput(msg string) *MacauWebOutput {
	return &MacauWebOutput{
		Players:       make([]*MacauWebOutputPlayer, 0),
		WinnerIdx:     -1,
		Direction:     1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func macauDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MacauInteractorIF, param MacauWebInput, newDefault func(string) *MacauWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "s", "suit":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.ChooseSuit(*param.Suit))
	case "dc", "declare":
		bc.writePresenterResponse(w, ci.Declare())
	case "sk", "skipdeclare":
		bc.writePresenterResponse(w, ci.SkipDeclare())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
