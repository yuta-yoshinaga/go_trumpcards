//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ConquianWebInput コンキャンWebインプット
type ConquianWebInput struct {
	BaseWebInput
	CardIndex  *int    `json:"cardIndex,omitempty"`
	MeldGroups [][]int `json:"meldGroups,omitempty"`
	// ExtendTargets は meldGroups の各グループの延長先メルド番号 (省略可)。
	ExtendTargets []int              `json:"extendTargets,omitempty"`
	Config        *ConquianWebConfig `json:"config,omitempty"`
}

// ConquianWebConfig コンキャンWeb設定
type ConquianWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetWins    *int `json:"targetWins,omitempty"`
}

// ConquianWebOutputMeld テーブルメルドのアウトプット
type ConquianWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// ConquianWebOutputPlayer コンキャンWebアウトプットプレイヤー
type ConquianWebOutputPlayer struct {
	ID        int                      `json:"id"`
	IsHuman   bool                     `json:"isHuman"`
	CardCount int                      `json:"cardCount"`
	Cards     []*WebOutputCard         `json:"cards"`
	Melds     []*ConquianWebOutputMeld `json:"melds"`
	Wins      int                      `json:"wins"`
}

// ConquianWebOutput コンキャンWebアウトプット
type ConquianWebOutput struct {
	Players          []*ConquianWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard             `json:"discardTop"`
	DrawPileCount    int                        `json:"drawPileCount"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	RoundWinnerIdx   int                        `json:"roundWinnerIdx"`
	TookDiscard      bool                       `json:"tookDiscard"`
	WebOutputBase
	Config ConquianWebOutputConfig `json:"config"`
}

// ConquianWebOutputConfig コンキャン設定アウトプット
type ConquianWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetWins    int `json:"targetWins"`
}

// ToConfig builds a ConquianConfig from the nested web config, applying bounds checking.
func (c *ConquianWebConfig) ToConfig() domain.ConquianConfig {
	cfg := domain.DefaultConquianConfig()
	cfg.CpuDifficulty = domain.ConquianCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ConquianCpuDifficultyEasy), int(domain.ConquianCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetWins, c.TargetWins, 1, 100)
	return cfg
}

// ToConfig builds a ConquianConfig from the web input.
func (p ConquianWebInput) ToConfig() domain.ConquianConfig {
	return configOrDefault(p.Config, (*ConquianWebConfig).ToConfig, domain.DefaultConquianConfig())
}

// ConquianWebController コンキャンWebコントローラークラス
type ConquianWebController = GameWebController[usecase.ConquianInteractorIF, ConquianWebInput, *ConquianWebOutput]

// NewConquianWebController and NewConquianWebControllerWithProvider are
// the standard and provider-backed constructors for ConquianWebController.
var NewConquianWebController, NewConquianWebControllerWithProvider = webControllerPair[usecase.ConquianInteractorIF, ConquianWebInput, *ConquianWebOutput](
	newConquianDefaultOutput, conquianDispatch,
)

func newConquianDefaultOutput(msg string) *ConquianWebOutput {
	return &ConquianWebOutput{
		Players:       make([]*ConquianWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func conquianDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ConquianInteractorIF, param ConquianWebInput, newDefault func(string) *ConquianWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "m", "meld":
		bc.writePresenterResponse(w, ci.MeldWithTargets(param.MeldGroups, param.ExtendTargets))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
